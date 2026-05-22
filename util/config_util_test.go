package util

import "testing"

func TestConfigContainsRuntimeKeys(t *testing.T) {
	keys := []string{
		"env",
		"logger.file",
		"logger.level",
		"db.name",
		"db.type",
		"db.mongodb.url",
		"db.mysql.url",
		"api.schema.validation",
		"auth.jwt.secret",
		"auth.jwt.expiration_minutes",
		"auth.refresh_token.expiration_days",
		"auth.registration.enabled",
		"admin.username",
		"admin.password",
		"cors.origins",
		"server.port",
		"server.host",
		"timezone",
		"api.version",
		"snowflake.worker_id",
		"default_categories.path",
		"verification.code.expire_minutes",
		"smtp.host",
		"smtp.port",
		"smtp.username",
		"smtp.password",
		"smtp.from_address",
		"smtp.from_name",
		"smtp.max_retries",
		"smtp.retry_interval",
		"smtp.rate_limit.daily_per_ip",
		"smtp.rate_limit.daily_per_email",
	}

	for _, key := range keys {
		if _, ok := configurationMap[key]; !ok {
			t.Fatalf("configurationMap missing key %q", key)
		}
	}
}

func TestLegacyDatabaseURIAliasesAreNotRegistered(t *testing.T) {
	if _, ok := configurationMap["mongodb.uri"]; ok {
		t.Fatal("legacy key mongodb.uri should not be registered; use db.mongodb.url")
	}
	if _, ok := configurationMap["mysql.uri"]; ok {
		t.Fatal("legacy key mysql.uri should not be registered; use db.mysql.url")
	}
}
