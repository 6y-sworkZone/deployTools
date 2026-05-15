package utils

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"time"
)

type LogLevel int

const (
	LevelDebug LogLevel = iota
	LevelInfo
	LevelWarn
	LevelError
)

const (
	DefaultMaxLogSize    int64 = 10 * 1024 * 1024
	DefaultMaxLogFiles   int   = 10
	DefaultLogBufferSize int   = 1024
)

type Logger struct {
	logger      *log.Logger
	logFile     *os.File
	level       LogLevel
	logDir      string
	maxFileSize int64
	maxFiles    int
	currentSize int64
}

var DefaultLogger *Logger

func init() {
	DefaultLogger, _ = NewLogger("", LevelInfo, DefaultMaxLogSize, DefaultMaxLogFiles)
}

func NewLogger(logDir string, level LogLevel, maxFileSize int64, maxFiles int) (*Logger, error) {
	if logDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		logDir = filepath.Join(homeDir, ".deploytools", "logs")
	}

	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, err
	}

	logFileName := fmt.Sprintf("deploy_%s.log", time.Now().Format("20060102_150405"))
	logFilePath := filepath.Join(logDir, logFileName)

	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}

	fileInfo, err := logFile.Stat()
	if err != nil {
		logFile.Close()
		return nil, err
	}

	multiWriter := io.MultiWriter(os.Stdout, logFile)
	logger := log.New(multiWriter, "", log.LstdFlags)

	l := &Logger{
		logger:      logger,
		logFile:     logFile,
		level:       level,
		logDir:      logDir,
		maxFileSize: maxFileSize,
		maxFiles:    maxFiles,
		currentSize: fileInfo.Size(),
	}

	if err := l.rotateLogs(); err != nil {
		Warn("Failed to rotate logs: %v", err)
	}

	return l, nil
}

func (l *Logger) rotateLogs() error {
	files, err := filepath.Glob(filepath.Join(l.logDir, "deploy_*.log"))
	if err != nil {
		return err
	}

	if len(files) <= l.maxFiles {
		return nil
	}

	sort.Strings(files)

	for i := 0; i < len(files)-l.maxFiles; i++ {
		if err := os.Remove(files[i]); err != nil {
			Warn("Failed to remove old log file %s: %v", files[i], err)
		}
	}

	return nil
}

func (l *Logger) checkSizeAndRotate() error {
	if l.currentSize < l.maxFileSize {
		return nil
	}

	if err := l.logFile.Close(); err != nil {
		return err
	}

	newLogFileName := fmt.Sprintf("deploy_%s.log", time.Now().Format("20060102_150405"))
	newLogFilePath := filepath.Join(l.logDir, newLogFileName)

	newLogFile, err := os.OpenFile(newLogFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}

	l.logFile = newLogFile
	l.currentSize = 0

	multiWriter := io.MultiWriter(os.Stdout, newLogFile)
	l.logger.SetOutput(multiWriter)

	return l.rotateLogs()
}

func (l *Logger) write(level string, format string, v ...interface{}) {
	msg := fmt.Sprintf(level+" "+format, v...)
	bytesWritten := int64(len(msg) + 20)

	l.currentSize += bytesWritten
	l.logger.Println(msg)

	if err := l.checkSizeAndRotate(); err != nil {
		_ = fmt.Sprintf("Log rotation failed: %v", err)
	}
}

func (l *Logger) SetLevel(level LogLevel) {
	l.level = level
}

func (l *Logger) Debug(format string, v ...interface{}) {
	if l.level <= LevelDebug {
		l.write("[DEBUG]", format, v...)
	}
}

func (l *Logger) Info(format string, v ...interface{}) {
	if l.level <= LevelInfo {
		l.write("[INFO] ", format, v...)
	}
}

func (l *Logger) Warn(format string, v ...interface{}) {
	if l.level <= LevelWarn {
		l.write("[WARN] ", format, v...)
	}
}

func (l *Logger) Error(format string, v ...interface{}) {
	if l.level <= LevelError {
		l.write("[ERROR]", format, v...)
	}
}

func (l *Logger) Close() error {
	if l.logFile != nil {
		return l.logFile.Close()
	}
	return nil
}

func (l *Logger) GetLogFilePath() string {
	if l.logFile != nil {
		return l.logFile.Name()
	}
	return ""
}

func (l *Logger) GetLogDir() string {
	return l.logDir
}

func (l *Logger) CleanupOldLogs(keepDays int) (int, error) {
	files, err := filepath.Glob(filepath.Join(l.logDir, "deploy_*.log"))
	if err != nil {
		return 0, err
	}

	cutoff := time.Now().AddDate(0, 0, -keepDays)
	deletedCount := 0

	for _, file := range files {
		info, err := os.Stat(file)
		if err != nil {
			continue
		}
		if info.ModTime().Before(cutoff) {
			if err := os.Remove(file); err == nil {
				deletedCount++
			}
		}
	}

	return deletedCount, nil
}

func (l *Logger) ListLogFiles() ([]string, error) {
	files, err := filepath.Glob(filepath.Join(l.logDir, "deploy_*.log"))
	if err != nil {
		return nil, err
	}
	sort.Strings(files)
	return files, nil
}

func Debug(format string, v ...interface{}) {
	DefaultLogger.Debug(format, v...)
}

func Info(format string, v ...interface{}) {
	DefaultLogger.Info(format, v...)
}

func Warn(format string, v ...interface{}) {
	DefaultLogger.Warn(format, v...)
}

func Error(format string, v ...interface{}) {
	DefaultLogger.Error(format, v...)
}
