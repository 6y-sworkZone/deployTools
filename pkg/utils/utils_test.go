package utils

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestFileExists(t *testing.T) {
	tempDir := t.TempDir()

	existingFile := filepath.Join(tempDir, "test.txt")
	os.WriteFile(existingFile, []byte("test"), 0644)

	if !FileExists(existingFile) {
		t.Error("Expected existing file to return true")
	}

	nonExistentFile := filepath.Join(tempDir, "nonexistent.txt")
	if FileExists(nonExistentFile) {
		t.Error("Expected non-existent file to return false")
	}
}

func TestIsDir(t *testing.T) {
	tempDir := t.TempDir()

	if !IsDir(tempDir) {
		t.Error("Expected directory to return true")
	}

	testFile := filepath.Join(tempDir, "test.txt")
	os.WriteFile(testFile, []byte("test"), 0644)

	if IsDir(testFile) {
		t.Error("Expected file to return false")
	}
}

func TestGetFileMD5(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")

	content := []byte("Hello, World!")
	os.WriteFile(testFile, content, 0644)

	md5Hash, err := GetFileMD5(testFile)
	if err != nil {
		t.Fatalf("Failed to get MD5: %v", err)
	}

	expected := "65a8e27d8879283cfdfdf1a9c0eab"
	if !strings.HasPrefix(md5Hash, expected[:10]) {
		t.Errorf("MD5 hash seems incorrect, got: %s", md5Hash)
	}

	if len(md5Hash) != 32 {
		t.Errorf("Expected MD5 hash length 32, got %d", len(md5Hash))
	}
}

func TestGetFileMD5NonExistent(t *testing.T) {
	_, err := GetFileMD5("/nonexistent/file.txt")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

func TestGetFileSize(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")

	content := []byte("12345")
	os.WriteFile(testFile, content, 0644)

	size, err := GetFileSize(testFile)
	if err != nil {
		t.Fatalf("Failed to get file size: %v", err)
	}

	if size != 5 {
		t.Errorf("Expected size 5, got %d", size)
	}
}

func TestFormatFileSize(t *testing.T) {
	tests := []struct {
		size     int64
		expected string
	}{
		{0, "0 B"},
		{512, "512 B"},
		{1024, "1.0 KB"},
		{1024 * 1024, "1.0 MB"},
		{1024 * 1024 * 1024, "1.0 GB"},
		{1536, "1.5 KB"},
	}

	for _, tt := range tests {
		result := FormatFileSize(tt.size)
		if result != tt.expected {
			t.Errorf("For size %d: expected %s, got %s", tt.size, tt.expected, result)
		}
	}
}

func TestNormalizePath(t *testing.T) {
	if runtime.GOOS == "windows" {
		path := "C:/test/path"
		normalized := NormalizePath(path)
		if !strings.Contains(normalized, "\\") {
			t.Error("Expected Windows path to use backslashes")
		}
	} else {
		path := "\\test\\path"
		normalized := NormalizePath(path)
		if !strings.Contains(normalized, "/") {
			t.Error("Expected Unix path to use forward slashes")
		}
	}
}

func TestValidateLocalPath(t *testing.T) {
	tempDir := t.TempDir()

	err := ValidateLocalPath(tempDir)
	if err != nil {
		t.Errorf("Expected no error for existing path, got: %v", err)
	}

	err = ValidateLocalPath("/nonexistent/path")
	if err == nil {
		t.Error("Expected error for non-existent path")
	}
}

func TestValidateRemotePath(t *testing.T) {
	err := ValidateRemotePath("")
	if err == nil {
		t.Error("Expected error for empty path")
	}

	err = ValidateRemotePath("relative/path")
	if err == nil {
		t.Error("Expected error for relative path")
	}

	err = ValidateRemotePath("/absolute/path")
	if err != nil {
		t.Errorf("Expected no error for absolute path, got: %v", err)
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		duration time.Duration
		expected string
	}{
		{0 * time.Second, "0s"},
		{45 * time.Second, "45s"},
		{2 * time.Minute, "2m 0s"},
		{2*time.Minute + 30*time.Second, "2m 30s"},
		{1 * time.Hour, "1h 0m 0s"},
		{1*time.Hour + 15*time.Minute + 30*time.Second, "1h 15m 30s"},
	}

	for _, tt := range tests {
		result := FormatDuration(tt.duration)
		if result != tt.expected {
			t.Errorf("For duration %v: expected %s, got %s", tt.duration, tt.expected, result)
		}
	}
}

func TestTruncateString(t *testing.T) {
	tests := []struct {
		input    string
		maxLen   int
		expected string
	}{
		{"short", 10, "short"},
		{"this is a very long string", 10, "this is..."},
		{"", 5, ""},
		{"exact", 5, "exact"},
		{"one", 2, "on"},
		{"test", 0, ""},
		{"testing", 3, "tes"},
	}

	for _, tt := range tests {
		result := TruncateString(tt.input, tt.maxLen)
		if result != tt.expected {
			t.Errorf("For '%s' (max %d): expected '%s', got '%s'",
				tt.input, tt.maxLen, tt.expected, result)
		}
	}
}

func TestGetExcludeMatcher(t *testing.T) {
	patterns := []string{"*.log", "node_modules", ".git"}
	matcher := GetExcludeMatcher(patterns)

	tests := []struct {
		path     string
		expected bool
	}{
		{"/test/file.log", true},
		{"/test/file.txt", false},
		{"/project/node_modules/package.json", true},
		{"/project/.git/config", true},
		{"/project/src/main.go", false},
	}

	for _, tt := range tests {
		result := matcher(tt.path)
		if result != tt.expected {
			t.Errorf("For path '%s': expected %v, got %v", tt.path, tt.expected, result)
		}
	}
}

func TestLogger(t *testing.T) {
	tempDir := t.TempDir()

	logger, err := NewLogger(tempDir, LevelDebug, 1024*1024, 10)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	logger.Debug("Debug message")
	logger.Info("Info message")
	logger.Warn("Warn message")
	logger.Error("Error message")

	logFile := logger.GetLogFilePath()
	if logFile == "" {
		t.Error("Expected log file path")
	}

	if !FileExists(logFile) {
		t.Error("Expected log file to exist")
	}

	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	if !strings.Contains(string(content), "Debug message") {
		t.Error("Log file should contain debug message")
	}
	if !strings.Contains(string(content), "Info message") {
		t.Error("Log file should contain info message")
	}
	if !strings.Contains(string(content), "Warn message") {
		t.Error("Log file should contain warn message")
	}
	if !strings.Contains(string(content), "Error message") {
		t.Error("Log file should contain error message")
	}

	err = logger.Close()
	if err != nil {
		t.Errorf("Failed to close logger: %v", err)
	}
}

func TestDefaultLogger(t *testing.T) {
	if DefaultLogger == nil {
		t.Error("DefaultLogger should not be nil")
	}

	Info("Test info message")
	Debug("Test debug message")
	Warn("Test warn message")
	Error("Test error message")
}
