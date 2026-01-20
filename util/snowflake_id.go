package util

import (
	"fmt"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Snowflake ID generator
// Generates 64-bit unique IDs with structure:
// - 41 bits: timestamp (milliseconds since custom epoch)
// - 10 bits: machine/worker ID
// - 12 bits: sequence number within the same millisecond
// This guarantees 4096 IDs per millisecond per machine

const (
	// Custom epoch (2020-01-01 00:00:00 UTC) - gives us ~69 years from 2020
	customEpoch int64 = 1577836800000

	// Bit allocation
	workerIDBits     = uint(10)
	sequenceBits     = uint(12)
	maxWorkerID      = int64(-1) ^ (int64(-1) << workerIDBits)  // 1023
	maxSequence      = int64(-1) ^ (int64(-1) << sequenceBits)  // 4095
	workerIDShift    = sequenceBits                              // 12
	timestampShift   = sequenceBits + workerIDBits               // 22
)

// SnowflakeIDGenerator generates unique distributed IDs
type SnowflakeIDGenerator struct {
	mu           sync.Mutex
	workerID     int64
	sequence     int64
	lastTimestamp int64
}

var (
	defaultGenerator *SnowflakeIDGenerator
	generatorOnce    sync.Once
)

// InitSnowflakeGenerator initializes the default Snowflake generator
// workerID should be unique per instance/machine (0-1023)
func InitSnowflakeGenerator(workerID int64) error {
	if workerID < 0 || workerID > maxWorkerID {
		return fmt.Errorf("worker ID must be between 0 and %d", maxWorkerID)
	}

	generatorOnce.Do(func() {
		defaultGenerator = &SnowflakeIDGenerator{
			workerID:     workerID,
			sequence:     0,
			lastTimestamp: -1,
		}
		Logger.Infof("Snowflake ID generator initialized with worker ID: %d", workerID)
	})

	return nil
}

// getDefaultGenerator returns the default generator, initializing with workerID 0 if needed
func getDefaultGenerator() *SnowflakeIDGenerator {
	if defaultGenerator == nil {
		// Auto-initialize with workerID 0 if not explicitly initialized
		_ = InitSnowflakeGenerator(0)
	}
	return defaultGenerator
}

// Generate creates a new unique ID
func (g *SnowflakeIDGenerator) Generate() (int64, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	timestamp := timeInMillis()

	// Clock moved backwards - wait until it catches up
	if timestamp < g.lastTimestamp {
		offset := g.lastTimestamp - timestamp
		if offset <= 5 {
			// Small clock drift, wait it out
			time.Sleep(time.Duration(offset) * time.Millisecond)
			timestamp = timeInMillis()
			if timestamp < g.lastTimestamp {
				return 0, fmt.Errorf("clock moved backwards, refusing to generate ID")
			}
		} else {
			return 0, fmt.Errorf("clock moved backwards by %dms, refusing to generate ID", offset)
		}
	}

	if timestamp == g.lastTimestamp {
		// Same millisecond - increment sequence
		g.sequence = (g.sequence + 1) & maxSequence
		if g.sequence == 0 {
			// Sequence overflow - wait for next millisecond
			timestamp = waitNextMillis(g.lastTimestamp)
		}
	} else {
		// New millisecond - reset sequence
		g.sequence = 0
	}

	g.lastTimestamp = timestamp

	// Compose the ID
	id := ((timestamp - customEpoch) << timestampShift) |
		(g.workerID << workerIDShift) |
		g.sequence

	return id, nil
}

// GenerateHexString generates a Snowflake ID and returns it as a hex string
// This maintains compatibility with MongoDB ObjectID string format
func (g *SnowflakeIDGenerator) GenerateHexString() (string, error) {
	id, err := g.Generate()
	if err != nil {
		return "", err
	}
	// Convert to 16-character hex string (64-bit)
	return fmt.Sprintf("%016x", id), nil
}

// GenerateObjectIDCompatible generates a Snowflake ID wrapped as MongoDB ObjectID
// The first 8 bytes contain the Snowflake ID, remaining bytes are padding
func (g *SnowflakeIDGenerator) GenerateObjectIDCompatible() (primitive.ObjectID, error) {
	id, err := g.Generate()
	if err != nil {
		// Fallback to standard ObjectID on error
		return primitive.NewObjectID(), nil
	}

	// Create a 12-byte array for ObjectID
	var objectIDBytes [12]byte

	// First 8 bytes: Snowflake ID
	objectIDBytes[0] = byte(id >> 56)
	objectIDBytes[1] = byte(id >> 48)
	objectIDBytes[2] = byte(id >> 40)
	objectIDBytes[3] = byte(id >> 32)
	objectIDBytes[4] = byte(id >> 24)
	objectIDBytes[5] = byte(id >> 16)
	objectIDBytes[6] = byte(id >> 8)
	objectIDBytes[7] = byte(id)

	// Last 4 bytes: padding (could use random bytes or counter)
	objectIDBytes[8] = byte(timestamp() >> 24)
	objectIDBytes[9] = byte(timestamp() >> 16)
	objectIDBytes[10] = byte(timestamp() >> 8)
	objectIDBytes[11] = byte(timestamp())

	return objectIDBytes, nil
}

// Helper functions
func timeInMillis() int64 {
	return time.Now().UnixNano() / 1e6
}

func timestamp() int64 {
	return time.Now().Unix()
}

func waitNextMillis(lastTimestamp int64) int64 {
	timestamp := timeInMillis()
	for timestamp <= lastTimestamp {
		timestamp = timeInMillis()
	}
	return timestamp
}

// Public API functions

// GenerateSnowflakeID generates a unique Snowflake ID using the default generator
func GenerateSnowflakeID() (int64, error) {
	return getDefaultGenerator().Generate()
}

// GenerateSnowflakeIDString generates a Snowflake ID as hex string
func GenerateSnowflakeIDString() (string, error) {
	return getDefaultGenerator().GenerateHexString()
}

// GenerateSnowflakeObjectID generates a Snowflake-based MongoDB ObjectID
func GenerateSnowflakeObjectID() (primitive.ObjectID, error) {
	return getDefaultGenerator().GenerateObjectIDCompatible()
}

// GenerateIDWithRetry generates an ID with retry logic for duplicate handling
// It will retry up to maxRetries times if the provided check function returns false
func GenerateIDWithRetry(checkUnique func(string) bool, maxRetries int) (string, error) {
	if maxRetries <= 0 {
		maxRetries = 3
	}

	for i := 0; i < maxRetries; i++ {
		id, err := GenerateSnowflakeIDString()
		if err != nil {
			Logger.Warnw("Failed to generate Snowflake ID, falling back to ObjectID",
				"error", err,
				"attempt", i+1)
			id = primitive.NewObjectID().Hex()
		}

		if checkUnique(id) {
			return id, nil
		}

		Logger.Warnw("Generated ID already exists, retrying",
			"id", id,
			"attempt", i+1)
	}

	return "", fmt.Errorf("failed to generate unique ID after %d attempts", maxRetries)
}
