package util

import (
	"strings"
	"testing"
)

func TestValidateRuntimeConfigurationAcceptsProductionBaseline(t *testing.T) {
	setRuntimeConfigForTest(t, map[string]string{
		"env":                                     "prod",
		"api.schema.validation":                   "false",
		"auth.registration.enabled":               "true",
		"smtp.enabled":                            "false",
		"metrics.enabled":                         "true",
		"metrics.bearer_token":                    strings.Repeat("m", 32),
		"api.rate_limit.requests_per_minute":      "600",
		"api.rate_limit.burst":                    "60",
		"logger.level":                            "info",
		"db.type":                                 "mongodb",
		"db.mongodb.url":                          "mongodb://database.example/cashlenx",
		"auth.jwt.secret":                         strings.Repeat("j", 32),
		"admin.password":                          "strong-admin-password",
		"cors.origins":                            "https://app.cashlenx.com,https://admin.cashlenx.com:8443",
		"auth.jwt.expiration_minutes":             "30",
		"auth.refresh_token.expiration_days":      "14",
		"verification.code.expire_minutes":        "30",
		"verification.code.send_interval_seconds": "60",
	})

	if err := ValidateRuntimeConfiguration(); err != nil {
		t.Fatalf("ValidateRuntimeConfiguration() error = %v", err)
	}
}

func TestValidateRuntimeConfigurationAcceptsMySQLAndIgnoresDisabledCapabilities(t *testing.T) {
	setRuntimeConfigForTest(t, map[string]string{
		"env":                                     "prod",
		"api.schema.validation":                   "false",
		"auth.registration.enabled":               "true",
		"smtp.enabled":                            "false",
		"smtp.host":                               "CHANGE_ME_SMTP_HOST",
		"smtp.password":                           "CHANGE_ME_SMTP_PASSWORD",
		"metrics.enabled":                         "false",
		"metrics.bearer_token":                    "CHANGE_ME_METRICS_TOKEN",
		"api.rate_limit.requests_per_minute":      "600",
		"api.rate_limit.burst":                    "60",
		"logger.level":                            "info",
		"db.type":                                 "mysql",
		"db.mongodb.url":                          "",
		"db.mysql.url":                            "cashlenx:strong-password@tcp(mysql:3306)",
		"auth.jwt.secret":                         strings.Repeat("j", 32),
		"admin.password":                          "strong-admin-password",
		"cors.origins":                            "https://app.cashlenx.com",
		"auth.jwt.expiration_minutes":             "30",
		"auth.refresh_token.expiration_days":      "14",
		"verification.code.expire_minutes":        "30",
		"verification.code.send_interval_seconds": "60",
	})

	if err := ValidateRuntimeConfiguration(); err != nil {
		t.Fatalf("ValidateRuntimeConfiguration() error = %v", err)
	}
}

func TestValidateRuntimeConfigurationRejectsUnsafeProductionValuesWithoutEchoingThem(t *testing.T) {
	secret := "short-secret-value"
	setRuntimeConfigForTest(t, map[string]string{
		"env":                                     "prod",
		"api.schema.validation":                   "false",
		"auth.registration.enabled":               "true",
		"smtp.enabled":                            "false",
		"metrics.enabled":                         "true",
		"metrics.bearer_token":                    secret,
		"api.rate_limit.requests_per_minute":      "0",
		"api.rate_limit.burst":                    "invalid",
		"logger.level":                            "verbose",
		"db.type":                                 "mongodb",
		"db.mongodb.url":                          "mongodb://CHANGE_ME_PASSWORD@database.example/cashlenx",
		"auth.jwt.secret":                         secret,
		"admin.password":                          "admin",
		"cors.origins":                            "http://localhost:*,https://app.cashlenx.com/path",
		"auth.jwt.expiration_minutes":             "30",
		"auth.refresh_token.expiration_days":      "14",
		"verification.code.expire_minutes":        "30",
		"verification.code.send_interval_seconds": "60",
	})

	err := ValidateRuntimeConfiguration()
	if err == nil {
		t.Fatal("expected unsafe production configuration to fail")
	}
	message := err.Error()
	for _, key := range []string{
		"JWT_SECRET", "ADMIN_PASSWORD", "CORS_ORIGINS", "METRICS_BEARER_TOKEN",
		"API_RATE_LIMIT_REQUESTS_PER_MINUTE", "API_RATE_LIMIT_BURST", "LOG_LEVEL", "MONGO_DB_URI",
	} {
		if !strings.Contains(message, key) {
			t.Fatalf("error %q does not identify %s", message, key)
		}
	}
	if strings.Contains(message, secret) {
		t.Fatal("validation error exposed a configured secret value")
	}
}

func setRuntimeConfigForTest(t *testing.T, values map[string]string) {
	t.Helper()
	original := make(map[string]string, len(values))
	for key, value := range values {
		original[key] = GetConfigByKey(key)
		SetConfigByKey(key, value)
	}
	t.Cleanup(func() {
		for key, value := range original {
			SetConfigByKey(key, value)
		}
	})
}
