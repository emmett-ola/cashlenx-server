package util

import "testing"

func TestConfiguredLogFolderFallsBackToEnvironment(t *testing.T) {
	original := configurationMap["logger.file"]
	t.Cleanup(func() {
		configurationMap["logger.file"] = original
	})

	configurationMap["logger.file"] = ""
	t.Setenv("LOG_FOLDER", "configured-test-logs")

	if got := configuredLogFolder(); got != "configured-test-logs" {
		t.Fatalf("configuredLogFolder() = %q, want %q", got, "configured-test-logs")
	}
}
