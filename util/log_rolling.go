package util

import (
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Log output is configuration-owned: the directory, the file-name template and
// the rolling policy all come from the environment, so one image serves local,
// testing and production without a rebuild. The logger initializes before
// configurationMap is populated, so these helpers read the environment first and
// fall back to the configuration map only after startup.
const (
	defaultLogDir         = "./logs"
	defaultLogPattern     = "cashlenx_{date}{shard}.log"
	defaultLogRotation    = "daily"
	defaultLogConsoleForm = "json"
)

type logSettings struct {
	Dir           string
	FilePattern   string
	Rotation      string
	MaxSizeBytes  int64
	RetentionDays int
	ConsoleFormat string
}

// loadLogSettings returns the effective logging configuration.
func loadLogSettings() logSettings {
	settings := logSettings{
		Dir:           configuredLogFolder(),
		FilePattern:   firstNonEmpty(os.Getenv("LOG_FILE_PATTERN"), GetConfigByKey("logger.pattern"), defaultLogPattern),
		Rotation:      firstNonEmpty(os.Getenv("LOG_ROTATION"), GetConfigByKey("logger.rotation"), defaultLogRotation),
		ConsoleFormat: firstNonEmpty(os.Getenv("LOG_CONSOLE_FORMAT"), GetConfigByKey("logger.console_format"), defaultLogConsoleForm),
	}

	if settings.Dir == "" {
		settings.Dir = defaultLogDir
	}
	if !validLogRotation(settings.Rotation) {
		log.Printf("Invalid LOG_ROTATION '%s', falling back to %s", settings.Rotation, defaultLogRotation)
		settings.Rotation = defaultLogRotation
	}
	if settings.ConsoleFormat != "text" {
		settings.ConsoleFormat = defaultLogConsoleForm
	}

	maxSizeMB := firstPositiveInt(os.Getenv("LOG_MAX_SIZE_MB"), GetConfigByKey("logger.max_size_mb"))
	settings.MaxSizeBytes = int64(maxSizeMB) * 1024 * 1024
	settings.RetentionDays = firstPositiveInt(os.Getenv("LOG_RETENTION_DAYS"), GetConfigByKey("logger.retention_days"))

	return settings
}

func validLogRotation(value string) bool {
	switch value {
	case "daily", "hourly", "size":
		return true
	default:
		return false
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func firstPositiveInt(values ...string) int {
	for _, value := range values {
		if value == "" {
			continue
		}
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed <= 0 {
			continue
		}
		return parsed
	}
	return 0
}

// renderLogPath expands the file-name template. {date} is YYYYMMDD, {time} is
// HHMMSS, {datetime} is YYYYMMDD-HHMMSS, {pid} is the process id and {shard}
// becomes _1, _2 ... for the second and later files of a period. A template
// without {shard} still gets a numeric suffix before the extension, so size
// rolling never appends to the file it just closed.
func renderLogPath(pattern, dir string, now time.Time, shard int) string {
	name := pattern
	name = strings.ReplaceAll(name, "{datetime}", now.Format("20060102-150405"))
	name = strings.ReplaceAll(name, "{date}", now.Format("20060102"))
	name = strings.ReplaceAll(name, "{time}", now.Format("150405"))
	name = strings.ReplaceAll(name, "{pid}", strconv.Itoa(os.Getpid()))

	if strings.Contains(pattern, "{shard}") {
		suffix := ""
		if shard > 0 {
			suffix = "_" + strconv.Itoa(shard)
		}
		name = strings.ReplaceAll(name, "{shard}", suffix)
	} else if shard > 0 {
		extension := filepath.Ext(name)
		name = strings.TrimSuffix(name, extension) + "_" + strconv.Itoa(shard) + extension
	}

	return filepath.Join(dir, name)
}

func logPeriodKey(rotation string, now time.Time) string {
	switch rotation {
	case "hourly":
		return now.Format("2006010215")
	case "size":
		return ""
	default:
		return now.Format("20060102")
	}
}

// rollingWriter appends JSON entries to one file per period and, when a size
// limit is configured, to further shards of the same period.
type rollingWriter struct {
	mu        sync.Mutex
	dir       string
	pattern   string
	rotation  string
	maxSize   int64
	retention int
	now       func() time.Time

	period string
	shard  int
	size   int64
	path   string
	file   *os.File
}

func newRollingWriter(settings logSettings) (*rollingWriter, error) {
	return newRollingWriterWithClock(settings, time.Now)
}

func newRollingWriterWithClock(settings logSettings, now func() time.Time) (*rollingWriter, error) {
	writer := &rollingWriter{
		dir:       settings.Dir,
		pattern:   settings.FilePattern,
		rotation:  settings.Rotation,
		maxSize:   settings.MaxSizeBytes,
		retention: settings.RetentionDays,
		now:       now,
	}
	if err := os.MkdirAll(writer.dir, 0o755); err != nil {
		return nil, err
	}
	if err := writer.openLocked(writer.now(), 0); err != nil {
		return nil, err
	}
	return writer, nil
}

func (writer *rollingWriter) Write(p []byte) (int, error) {
	writer.mu.Lock()
	defer writer.mu.Unlock()

	now := writer.now()
	period := logPeriodKey(writer.rotation, now)
	shard := writer.shard
	rotate := writer.file == nil

	if !rotate && writer.rotation != "size" && period != writer.period {
		rotate = true
		shard = 0
	}
	if !rotate && writer.maxSize > 0 && writer.size+int64(len(p)) > writer.maxSize {
		rotate = true
		shard++
	}

	if rotate {
		if err := writer.openLocked(now, shard); err != nil {
			return 0, err
		}
	}

	written, err := writer.file.Write(p)
	writer.size += int64(written)
	return written, err
}

func (writer *rollingWriter) Sync() error {
	writer.mu.Lock()
	defer writer.mu.Unlock()

	if writer.file == nil {
		return nil
	}
	return writer.file.Sync()
}

func (writer *rollingWriter) openLocked(now time.Time, shard int) error {
	path := renderLogPath(writer.pattern, writer.dir, now, shard)

	file, err := os.OpenFile(path, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0o644)
	if err != nil {
		return err
	}

	if writer.file != nil {
		_ = writer.file.Close()
	}

	size := int64(0)
	if info, statErr := file.Stat(); statErr == nil {
		size = info.Size()
	}

	previousPeriod := writer.period
	writer.file = file
	writer.path = path
	writer.shard = shard
	writer.size = size
	writer.period = logPeriodKey(writer.rotation, now)

	if previousPeriod != writer.period {
		writer.pruneLocked(now)
	}
	return nil
}

// pruneLocked deletes rolled files that fell out of the retention window. It
// only touches files that match the configured template's prefix and suffix, so
// an unrelated file in the same directory is never removed.
func (writer *rollingWriter) pruneLocked(now time.Time) {
	if writer.retention <= 0 {
		return
	}

	prefix, suffix := logFileNameBounds(writer.pattern)
	entries, err := os.ReadDir(writer.dir)
	if err != nil {
		return
	}
	cutoff := now.AddDate(0, 0, -writer.retention)
	current := filepath.Base(writer.path)

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if name == current || !strings.HasPrefix(name, prefix) || !strings.HasSuffix(name, suffix) {
			continue
		}
		info, infoErr := entry.Info()
		if infoErr != nil || !info.ModTime().Before(cutoff) {
			continue
		}
		_ = os.Remove(filepath.Join(writer.dir, name))
	}
}

func logFileNameBounds(pattern string) (string, string) {
	prefix := pattern
	suffix := ""
	if open := strings.Index(pattern, "{"); open >= 0 {
		prefix = pattern[:open]
	}
	if close := strings.LastIndex(pattern, "}"); close >= 0 {
		suffix = pattern[close+1:]
	}
	if prefix == "" {
		prefix = "cashlenx"
	}
	return prefix, suffix
}
