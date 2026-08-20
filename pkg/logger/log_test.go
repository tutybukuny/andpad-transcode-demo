package logger

import "testing"

func TestLog(t *testing.T) {
	logger := New(LogConfig{
		Env:          "dev",
		Level:        "info",
		EnableStdout: true,
		Encoding:     "console",
	})
	logger.Debug("Test log message should not be logged")
	logger.Info("Test log message should be seen")
	logger.Error("Test log message should be logged")
}
