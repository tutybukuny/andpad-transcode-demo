package config

import (
	"database/sql"
	"errors"
	"strconv"
	"strings"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"transcode-demo/pkg/aws"
	"transcode-demo/pkg/db"
	"transcode-demo/pkg/logger"
	"transcode-demo/pkg/utils"
)

type HttpServerConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

func (c *HttpServerConfig) SetDefault(parentKey string) {
	viper.SetDefault(utils.JoinConfigKeys(parentKey, "host"), "0.0.0.0")
	viper.SetDefault(utils.JoinConfigKeys(parentKey, "port"), "8000")
}

func (c *HttpServerConfig) GetAddr() string {
	return c.Host + ":" + strconv.Itoa(c.Port)
}

// Config is struct for config
type Config struct {
	Log        logger.LogConfig `mapstructure:"log"`
	HttpConfig HttpServerConfig `mapstructure:"http"`

	DB  db.DBConfig `mapstructure:"db"`
	AWS aws.Config  `mapstructure:"aws"`

	// watcher configurations
	WatcherSleepInterval   time.Duration `mapstructure:"watcher_sleep_interval"`
	StalledRequestDuration time.Duration `mapstructure:"stalled_request_duration"`
	RetriedTimesThreshold  int           `mapstructure:"retried_times_threshold"`

	// worker configurations
	GetTranscodeRequestInterval time.Duration `mapstructure:"get_transcode_request_interval"`

	// transcode configurations
	TranscodeStoragePrefix         string        `mapstructure:"transcode_storage_prefix"`
	TranscodeStreamingPrefix       string        `mapstructure:"transcode_streaming_prefix"`
	TranscodeTimeLimit             time.Duration `mapstructure:"transcode_time_limit"`
	WatchTransReqInterval          time.Duration `mapstructure:"watch_trans_req_interval"`
	UpdateLastProcessingAtInterval time.Duration `mapstructure:"update_last_processing_at_interval"`
}

// New creates a new config for the application.
func New() *Config {
	cfg, err := newCfg()
	if err != nil {
		panic(err)
	}
	return cfg
}

func (c *Config) SetDefault() {
	c.Log.SetDefault("log")
	c.HttpConfig.SetDefault("http")
	c.DB.SetDefault("db")
	c.AWS.SetDefault("aws")

	viper.SetDefault("watcher_sleep_interval", 10*time.Second)
	viper.SetDefault("stalled_request_duration", 10*time.Second)
	viper.SetDefault("retried_times_threshold", 3)

	viper.SetDefault("get_transcode_request_interval", 5*time.Second)

	viper.SetDefault("transcode_storage_prefix", "s3://local-bucket/transcode-output")
	viper.SetDefault("transcode_streaming_prefix", "http://localhost:4566/local-bucket/transcode-output")
	viper.SetDefault("transcode_time_limit", 10*time.Minute)
	viper.SetDefault("watch_trans_req_interval", 5*time.Second)
	viper.SetDefault("update_last_processing_at_interval", 5*time.Second)
}

func (c *Config) GetDB() (*sql.DB, error) {
	psqlInfo := c.DB.ConnectionString()
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func newCfg() (*Config, error) {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./samples")
	viper.SetConfigType("yaml")
	viper.SetEnvPrefix("tdapp")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
	cfg := Config{}
	cfg.SetDefault()
	err := viper.ReadInConfig()
	if err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if !errors.As(err, &configFileNotFoundError) {
			return nil, err
		}
	}
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// MustNewDevConfig creates a new config for the unit test
func MustNewDevConfig(t testing.TB) *Config {
	cfg, err := newCfg()
	require.NoError(t, err)
	return cfg
}

func MustNewDB(t testing.TB) *sql.DB {
	db, err := MustNewDevConfig(t).GetDB()
	require.NoError(t, err)
	t.Cleanup(func() {
		err = db.Close()
		require.NoError(t, err)
	})
	return db
}

func MustNewGorm(t testing.TB) *gorm.DB {
	dbConn := MustNewDB(t)
	gormDB, err := db.NewGormDB(dbConn)
	require.NoError(t, err)
	return gormDB
}
