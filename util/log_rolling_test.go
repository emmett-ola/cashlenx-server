package util

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func resetLogConfig(t *testing.T) {
	t.Helper()
	keys := []string{
		"logger.file", "logger.dir", "logger.pattern", "logger.rotation",
		"logger.max_size_mb", "logger.retention_days", "logger.console_format",
	}
	original := make(map[string]string, len(keys))
	for _, key := range keys {
		original[key] = configurationMap[key]
		configurationMap[key] = ""
	}
	t.Cleanup(func() {
		for key, value := range original {
			configurationMap[key] = value
		}
	})
	for _, key := range []string{
		"LOG_DIR", "LOG_FOLDER", "LOG_FILE_PATTERN", "LOG_ROTATION",
		"LOG_MAX_SIZE_MB", "LOG_RETENTION_DAYS", "LOG_CONSOLE_FORMAT",
	} {
		t.Setenv(key, "")
	}
}

func TestRenderLogPathExpandsTokens(t *testing.T) {
	now := time.Date(2026, 11, 22, 13, 45, 6, 0, time.UTC)

	cases := []struct {
		name    string
		pattern string
		shard   int
		want    string
	}{
		{"default first file", "cashlenx_{date}{shard}.log", 0, "cashlenx_20261122.log"},
		{"default second shard", "cashlenx_{date}{shard}.log", 2, "cashlenx_20261122_2.log"},
		{"hourly template", "cashlenx_{datetime}.log", 0, "cashlenx_20261122-134506.log"},
		{"time token", "app_{time}.log", 0, "app_134506.log"},
		// A template without {shard} still has to produce a distinct file.
		{"no shard token", "app.log", 3, "app_3.log"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			got := filepath.Base(renderLogPath(testCase.pattern, "/var/log/cashlenx", now, testCase.shard))
			if got != testCase.want {
				t.Fatalf("renderLogPath() = %q, want %q", got, testCase.want)
			}
		})
	}
}

func TestRenderLogPathJoinsConfiguredDirectory(t *testing.T) {
	now := time.Date(2026, 11, 22, 0, 0, 0, 0, time.UTC)
	got := renderLogPath(defaultLogPattern, "/var/log/cashlenx", now, 0)
	want := filepath.Join("/var/log/cashlenx", "cashlenx_20261122.log")
	if got != want {
		t.Fatalf("renderLogPath() = %q, want %q", got, want)
	}
}

func TestLogPeriodKey(t *testing.T) {
	now := time.Date(2026, 11, 22, 13, 45, 6, 0, time.UTC)
	if got := logPeriodKey("daily", now); got != "20261122" {
		t.Fatalf("daily key = %q", got)
	}
	if got := logPeriodKey("hourly", now); got != "2026112213" {
		t.Fatalf("hourly key = %q", got)
	}
	if got := logPeriodKey("size", now); got != "" {
		t.Fatalf("size key = %q, want empty", got)
	}
}

func TestLoadLogSettingsReadsEnvironment(t *testing.T) {
	resetLogConfig(t)

	t.Setenv("LOG_DIR", "/var/log/cashlenx")
	t.Setenv("LOG_FILE_PATTERN", "api_{date}_{time}{shard}.log")
	t.Setenv("LOG_ROTATION", "hourly")
	t.Setenv("LOG_MAX_SIZE_MB", "32")
	t.Setenv("LOG_RETENTION_DAYS", "14")
	t.Setenv("LOG_CONSOLE_FORMAT", "text")

	settings := loadLogSettings()

	if settings.Dir != "/var/log/cashlenx" {
		t.Fatalf("Dir = %q", settings.Dir)
	}
	if settings.FilePattern != "api_{date}_{time}{shard}.log" {
		t.Fatalf("FilePattern = %q", settings.FilePattern)
	}
	if settings.Rotation != "hourly" {
		t.Fatalf("Rotation = %q", settings.Rotation)
	}
	if settings.MaxSizeBytes != 32*1024*1024 {
		t.Fatalf("MaxSizeBytes = %d", settings.MaxSizeBytes)
	}
	if settings.RetentionDays != 14 {
		t.Fatalf("RetentionDays = %d", settings.RetentionDays)
	}
	if settings.ConsoleFormat != "text" {
		t.Fatalf("ConsoleFormat = %q", settings.ConsoleFormat)
	}
}

func TestLoadLogSettingsDefaultsAndFallbacks(t *testing.T) {
	resetLogConfig(t)

	t.Setenv("LOG_FOLDER", "./legacy-logs")
	t.Setenv("LOG_ROTATION", "weekly")
	t.Setenv("LOG_CONSOLE_FORMAT", "xml")
	t.Setenv("LOG_MAX_SIZE_MB", "not-a-number")

	settings := loadLogSettings()

	if settings.Dir != "./legacy-logs" {
		t.Fatalf("Dir = %q, want the LOG_FOLDER fallback", settings.Dir)
	}
	if settings.Rotation != defaultLogRotation {
		t.Fatalf("Rotation = %q, want %q", settings.Rotation, defaultLogRotation)
	}
	if settings.ConsoleFormat != defaultLogConsoleForm {
		t.Fatalf("ConsoleFormat = %q, want %q", settings.ConsoleFormat, defaultLogConsoleForm)
	}
	if settings.MaxSizeBytes != 0 {
		t.Fatalf("MaxSizeBytes = %d, want 0", settings.MaxSizeBytes)
	}
	if settings.FilePattern != defaultLogPattern {
		t.Fatalf("FilePattern = %q, want %q", settings.FilePattern, defaultLogPattern)
	}
}

func TestLoadLogSettingsPrefersLogDirOverLogFolder(t *testing.T) {
	resetLogConfig(t)

	t.Setenv("LOG_DIR", "./current-logs")
	t.Setenv("LOG_FOLDER", "./legacy-logs")

	if got := loadLogSettings().Dir; got != "./current-logs" {
		t.Fatalf("Dir = %q, want the LOG_DIR value", got)
	}
}

func TestRollingWriterRotatesWhenTheDayChanges(t *testing.T) {
	dir := t.TempDir()
	now := time.Date(2026, 11, 22, 23, 59, 0, 0, time.UTC)

	writer, err := newRollingWriterWithClock(logSettings{
		Dir:         dir,
		FilePattern: defaultLogPattern,
		Rotation:    "daily",
	}, func() time.Time { return now })
	if err != nil {
		t.Fatalf("newRollingWriter: %v", err)
	}

	if _, err := writer.Write([]byte("first\n")); err != nil {
		t.Fatalf("write: %v", err)
	}

	now = now.Add(2 * time.Minute)
	if _, err := writer.Write([]byte("second\n")); err != nil {
		t.Fatalf("write: %v", err)
	}

	first := filepath.Join(dir, "cashlenx_20261122.log")
	second := filepath.Join(dir, "cashlenx_20261123.log")

	if content, err := os.ReadFile(first); err != nil || string(content) != "first\n" {
		t.Fatalf("first file = %q, %v", content, err)
	}
	if content, err := os.ReadFile(second); err != nil || string(content) != "second\n" {
		t.Fatalf("second file = %q, %v", content, err)
	}
}

func TestRollingWriterShardsWhenTheSizeLimitIsHit(t *testing.T) {
	dir := t.TempDir()
	now := time.Date(2026, 11, 22, 12, 0, 0, 0, time.UTC)

	writer, err := newRollingWriterWithClock(logSettings{
		Dir:          dir,
		FilePattern:  defaultLogPattern,
		Rotation:     "size",
		MaxSizeBytes: 8,
	}, func() time.Time { return now })
	if err != nil {
		t.Fatalf("newRollingWriter: %v", err)
	}

	for _, entry := range []string{"aaaa\n", "bbbb\n", "cccc\n"} {
		if _, err := writer.Write([]byte(entry)); err != nil {
			t.Fatalf("write %q: %v", entry, err)
		}
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read dir: %v", err)
	}
	if len(entries) != 3 {
		names := make([]string, 0, len(entries))
		for _, entry := range entries {
			names = append(names, entry.Name())
		}
		t.Fatalf("expected 3 shards, got %d: %v", len(entries), names)
	}
	if _, err := os.Stat(filepath.Join(dir, "cashlenx_20261122.log")); err != nil {
		t.Fatalf("base shard missing: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "cashlenx_20261122_1.log")); err != nil {
		t.Fatalf("second shard missing: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "cashlenx_20261122_2.log")); err != nil {
		t.Fatalf("third shard missing: %v", err)
	}
}

func TestRollingWriterAppendsToAnExistingFileOnRestart(t *testing.T) {
	dir := t.TempDir()
	now := time.Date(2026, 11, 22, 9, 0, 0, 0, time.UTC)
	path := filepath.Join(dir, "cashlenx_20261122.log")
	if err := os.WriteFile(path, []byte("earlier\n"), 0o644); err != nil {
		t.Fatalf("seed file: %v", err)
	}

	writer, err := newRollingWriterWithClock(logSettings{
		Dir:         dir,
		FilePattern: defaultLogPattern,
		Rotation:    "daily",
	}, func() time.Time { return now })
	if err != nil {
		t.Fatalf("newRollingWriter: %v", err)
	}
	if _, err := writer.Write([]byte("restarted\n")); err != nil {
		t.Fatalf("write: %v", err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if got := string(content); got != "earlier\nrestarted\n" {
		t.Fatalf("file = %q, want the previous content plus the new line", got)
	}
}

func TestRollingWriterPrunesFilesOutsideRetention(t *testing.T) {
	dir := t.TempDir()
	now := time.Date(2026, 11, 22, 0, 0, 0, 0, time.UTC)

	stale := filepath.Join(dir, "cashlenx_20261001.log")
	fresh := filepath.Join(dir, "cashlenx_20261121.log")
	unrelated := filepath.Join(dir, "notes.txt")
	for _, path := range []string{stale, fresh, unrelated} {
		if err := os.WriteFile(path, []byte("x\n"), 0o644); err != nil {
			t.Fatalf("seed %s: %v", path, err)
		}
	}
	oldTime := now.AddDate(0, 0, -30)
	if err := os.Chtimes(stale, oldTime, oldTime); err != nil {
		t.Fatalf("age stale file: %v", err)
	}
	if err := os.Chtimes(unrelated, oldTime, oldTime); err != nil {
		t.Fatalf("age unrelated file: %v", err)
	}

	writer, err := newRollingWriterWithClock(logSettings{
		Dir:           dir,
		FilePattern:   defaultLogPattern,
		Rotation:      "daily",
		RetentionDays: 7,
	}, func() time.Time { return now })
	if err != nil {
		t.Fatalf("newRollingWriter: %v", err)
	}
	// A period change is what triggers the sweep.
	now = now.Add(24 * time.Hour)
	if _, err := writer.Write([]byte("next day\n")); err != nil {
		t.Fatalf("write: %v", err)
	}

	if _, err := os.Stat(stale); !os.IsNotExist(err) {
		t.Fatalf("stale log should have been pruned, stat error = %v", err)
	}
	if _, err := os.Stat(fresh); err != nil {
		t.Fatalf("fresh log should survive: %v", err)
	}
	if _, err := os.Stat(unrelated); err != nil {
		t.Fatalf("unrelated file must not be pruned: %v", err)
	}
}

func TestConsoleLogEncoderHonoursTheConfiguredFormat(t *testing.T) {
	entry := zapcore.Entry{Level: zapcore.InfoLevel, Message: "hello", Time: time.Now()}
	fields := []zapcore.Field{zap.String("request_id", "abc")}

	jsonOut, err := consoleLogEncoder("json").EncodeEntry(entry, fields)
	if err != nil {
		t.Fatalf("json encode: %v", err)
	}
	var decoded map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSpace(jsonOut.String())), &decoded); err != nil {
		t.Fatalf("json encoder output is not JSON: %v (%q)", err, jsonOut.String())
	}
	if decoded["msg"] != "hello" {
		t.Fatalf("json entry msg = %v", decoded["msg"])
	}
	if decoded["request_id"] != "abc" {
		t.Fatalf("json entry request_id = %v", decoded["request_id"])
	}

	textOut, err := consoleLogEncoder("text").EncodeEntry(entry, fields)
	if err != nil {
		t.Fatalf("text encode: %v", err)
	}
	if strings.HasPrefix(strings.TrimSpace(textOut.String()), "{") {
		t.Fatalf("text format should not emit JSON: %q", textOut.String())
	}
}
