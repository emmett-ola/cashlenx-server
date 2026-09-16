package util

import (
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
)

// ValidateRuntimeConfiguration rejects invalid and unsafe runtime settings
// without returning configured values in the error.
func ValidateRuntimeConfiguration() error {
	invalid := make([]string, 0)
	addInvalid := func(key string) {
		for _, existing := range invalid {
			if existing == key {
				return
			}
		}
		invalid = append(invalid, key)
	}

	env := GetConfigByKey("env")
	if env != "dev" && env != "test" && env != "prod" {
		addInvalid("ENV")
	}

	for key, configKey := range map[string]string{
		"SCHEMA_VALIDATION":         "api.schema.validation",
		"AUTH_REGISTRATION_ENABLED": "auth.registration.enabled",
		"SMTP_ENABLED":              "smtp.enabled",
		"METRICS_ENABLED":           "metrics.enabled",
	} {
		if !isBoolean(GetConfigByKey(configKey)) {
			addInvalid(key)
		}
	}

	for key, configKey := range map[string]string{
		"JWT_EXPIRATION_MINUTES":                  "auth.jwt.expiration_minutes",
		"REFRESH_TOKEN_EXPIRATION_DAYS":           "auth.refresh_token.expiration_days",
		"VERIFICATION_CODE_EXPIRE_MINUTES":        "verification.code.expire_minutes",
		"VERIFICATION_CODE_SEND_INTERVAL_SECONDS": "verification.code.send_interval_seconds",
		"API_RATE_LIMIT_REQUESTS_PER_MINUTE":      "api.rate_limit.requests_per_minute",
		"API_RATE_LIMIT_BURST":                    "api.rate_limit.burst",
	} {
		if !isPositiveInteger(GetConfigByKey(configKey)) {
			addInvalid(key)
		}
	}

	if !isSupportedLogLevel(GetConfigByKey("logger.level")) {
		addInvalid("LOG_LEVEL")
	}

	if dbType := GetConfigByKey("db.type"); dbType == "mongodb" {
		if isUnsafeOrEmpty(GetConfigByKey("db.mongodb.url")) {
			addInvalid("MONGO_DB_URI")
		}
	} else if dbType == "mysql" {
		if isUnsafeOrEmpty(GetConfigByKey("db.mysql.url")) {
			addInvalid("MYSQL_DB_URI")
		}
	} else {
		addInvalid("DB_TYPE")
	}

	if GetConfigByKey("smtp.enabled") == "true" {
		for key, configKey := range map[string]string{
			"SMTP_HOST":         "smtp.host",
			"SMTP_PORT":         "smtp.port",
			"SMTP_USERNAME":     "smtp.username",
			"SMTP_PASSWORD":     "smtp.password",
			"SMTP_FROM_ADDRESS": "smtp.from_address",
		} {
			if isUnsafeOrEmpty(GetConfigByKey(configKey)) {
				addInvalid(key)
			}
		}
		if !isPositiveInteger(GetConfigByKey("smtp.port")) {
			addInvalid("SMTP_PORT")
		}
	}

	if secret := GetConfigByKey("auth.jwt.secret"); isUnsafeOrEmpty(secret) {
		addInvalid("JWT_SECRET")
	}
	if password := GetConfigByKey("admin.password"); isUnsafeOrEmpty(password) || password == "admin" {
		addInvalid("ADMIN_PASSWORD")
	}

	if env == "prod" {
		if secret := GetConfigByKey("auth.jwt.secret"); len(secret) < 32 {
			addInvalid("JWT_SECRET")
		}
		if password := GetConfigByKey("admin.password"); len(password) < 12 {
			addInvalid("ADMIN_PASSWORD")
		}
		if !validProductionOrigins(GetConfigByKey("cors.origins")) {
			addInvalid("CORS_ORIGINS")
		}
		if GetConfigByKey("metrics.enabled") == "true" {
			if token := GetConfigByKey("metrics.bearer_token"); len(token) < 32 || isUnsafeOrEmpty(token) {
				addInvalid("METRICS_BEARER_TOKEN")
			}
		}
	}

	if len(invalid) > 0 {
		return fmt.Errorf("invalid or unsafe runtime configuration keys: %s", strings.Join(invalid, ", "))
	}
	return nil
}

func isBoolean(value string) bool {
	return value == "true" || value == "false"
}

func isPositiveInteger(value string) bool {
	parsed, err := strconv.ParseInt(value, 10, 64)
	return err == nil && parsed > 0
}

func isSupportedLogLevel(value string) bool {
	switch value {
	case "debug", "info", "warn", "error", "dpanic", "panic", "fatal":
		return true
	default:
		return false
	}
}

func isUnsafeOrEmpty(value string) bool {
	normalized := strings.ToLower(strings.TrimSpace(value))
	return normalized == "" || strings.Contains(normalized, "change_me") ||
		strings.Contains(normalized, "change-in-production")
}

func validProductionOrigins(value string) bool {
	origins := strings.Split(value, ",")
	if len(origins) == 0 {
		return false
	}
	for _, rawOrigin := range origins {
		origin := strings.TrimSpace(rawOrigin)
		if origin == "" || strings.Contains(origin, "*") {
			return false
		}
		parsed, err := url.Parse(origin)
		if err != nil || parsed.Scheme != "https" || parsed.Host == "" || parsed.User != nil ||
			parsed.Path != "" || parsed.RawQuery != "" || parsed.Fragment != "" {
			return false
		}
		host := parsed.Hostname()
		if strings.EqualFold(host, "localhost") {
			return false
		}
		if ip := net.ParseIP(host); ip != nil && ip.IsLoopback() {
			return false
		}
		if port := parsed.Port(); port != "" {
			parsedPort, err := strconv.Atoi(port)
			if err != nil || parsedPort < 1 || parsedPort > 65535 {
				return false
			}
		}
	}
	return true
}
