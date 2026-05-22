package email

import (
	"fmt"
	"strings"
	"sync"
	"time"

	appErrors "github.com/macar-x/cashlenx-server/errors"
	"github.com/macar-x/cashlenx-server/util"
)

type dailyRateLimitEntry struct {
	day   string
	count int
}

type dailyRateLimiter struct {
	mu      sync.Mutex
	byIP    map[string]dailyRateLimitEntry
	byEmail map[string]dailyRateLimitEntry
}

var limiter = &dailyRateLimiter{
	byIP:    map[string]dailyRateLimitEntry{},
	byEmail: map[string]dailyRateLimitEntry{},
}

var (
	getDailyPerIPLimit = func() int {
		return int(util.GetConfigInt("smtp.rate_limit.daily_per_ip", 5))
	}
	getDailyPerEmailLimit = func() int {
		return int(util.GetConfigInt("smtp.rate_limit.daily_per_email", 5))
	}
)

// CheckAndRecordPurposeEmailAllowance applies daily abuse limits for a purpose-scoped email attempt.
func CheckAndRecordPurposeEmailAllowance(purpose string, ipAddress string, recipients []string) error {
	purpose = normalizeLimitPart(purpose)
	if purpose == "" {
		purpose = "email"
	}

	perIPLimit := getDailyPerIPLimit()
	perEmailLimit := getDailyPerEmailLimit()
	if perIPLimit <= 0 && perEmailLimit <= 0 {
		return nil
	}

	day := time.Now().Format("2006-01-02")
	ipAddress = normalizeLimitPart(ipAddress)

	emailKeys := make([]string, 0, len(recipients))
	for _, recipient := range recipients {
		recipient = strings.ToLower(normalizeLimitPart(recipient))
		if recipient != "" {
			emailKeys = append(emailKeys, fmt.Sprintf("%s|%s", purpose, recipient))
		}
	}

	limiter.mu.Lock()
	defer limiter.mu.Unlock()

	limiter.pruneLocked(day)

	if perIPLimit > 0 && ipAddress != "" {
		key := fmt.Sprintf("%s|%s", purpose, ipAddress)
		if entry := limiter.byIP[key]; entry.day == day && entry.count >= perIPLimit {
			return appErrors.NewRateLimitedError("too many email requests for this purpose from this IP address; try again tomorrow")
		}
	}

	if perEmailLimit > 0 {
		for _, key := range emailKeys {
			if entry := limiter.byEmail[key]; entry.day == day && entry.count >= perEmailLimit {
				return appErrors.NewRateLimitedError("too many email requests for this purpose to this email address; try again tomorrow")
			}
		}
	}

	if perIPLimit > 0 && ipAddress != "" {
		key := fmt.Sprintf("%s|%s", purpose, ipAddress)
		entry := limiter.byIP[key]
		if entry.day != day {
			entry = dailyRateLimitEntry{day: day}
		}
		entry.count++
		limiter.byIP[key] = entry
	}

	if perEmailLimit > 0 {
		for _, key := range emailKeys {
			entry := limiter.byEmail[key]
			if entry.day != day {
				entry = dailyRateLimitEntry{day: day}
			}
			entry.count++
			limiter.byEmail[key] = entry
		}
	}

	return nil
}

func (l *dailyRateLimiter) pruneLocked(currentDay string) {
	for key, entry := range l.byIP {
		if entry.day != currentDay {
			delete(l.byIP, key)
		}
	}
	for key, entry := range l.byEmail {
		if entry.day != currentDay {
			delete(l.byEmail, key)
		}
	}
}

func normalizeLimitPart(value string) string {
	return strings.TrimSpace(value)
}
