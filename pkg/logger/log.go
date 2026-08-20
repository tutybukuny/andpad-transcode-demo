package logger

import (
	"fmt"

	"github.com/spf13/viper"
	z "go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"transcode-demo/pkg/utils"
)

// LogConfig is struct for log config
type LogConfig struct {
	Env            string `mapstructure:"env"`
	Level          string `mapstructure:"level"`
	EnableStdout   bool   `mapstructure:"stdout"`
	OutputFile     string `mapstructure:"output_file"`
	Encoding       string `mapstructure:"encoding"`
	ShowStackTrace bool   `mapstructure:"show_stack_trace"`
}

// SetDefault sets default values for log config
func (c *LogConfig) SetDefault(parentKey string) {
	viper.SetDefault(utils.JoinConfigKeys(parentKey, "env"), "dev")
	viper.SetDefault(utils.JoinConfigKeys(parentKey, "level"), "debug")
	viper.SetDefault(utils.JoinConfigKeys(parentKey, "stdout"), true)
	viper.SetDefault(utils.JoinConfigKeys(parentKey, "output_file"), "")
	viper.SetDefault(utils.JoinConfigKeys(parentKey, "encoding"), "console")
	viper.SetDefault(utils.JoinConfigKeys(parentKey, "show_stack_trace"), false)
}

// New creates a new logger for the application.
func New(cfg LogConfig) *z.Logger {
	encoderConfig := z.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeDuration = zapcore.StringDurationEncoder
	if cfg.Encoding == "console" {
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}
	if !cfg.ShowStackTrace {
		encoderConfig.StacktraceKey = ""
	}
	level := z.DebugLevel
	if cfg.Level != "" {
		err := level.UnmarshalText([]byte(cfg.Level))
		if err != nil {
			panic(fmt.Sprintf("invalid log level: %s", cfg.Level))
		}
	}
	zCfg := z.Config{
		Level:            z.NewAtomicLevelAt(level),
		Encoding:         cfg.Encoding,
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
		EncoderConfig:    encoderConfig,
	}

	switch cfg.Env {
	case "prod":
		zCfg.Development = false
	default:
		// default will be dev env
		zCfg.Development = true
	}

	logger, err := zCfg.Build()
	if err != nil {
		panic(fmt.Sprintf("failed to create logger: %v", err))
	}
	return logger.With(z.String("env", cfg.Env))
}
