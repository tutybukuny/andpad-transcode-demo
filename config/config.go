package config

import (
	"database/sql"
	"errors"
	"strconv"
	"strings"
	"testing"

	_ "github.com/lib/pq"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

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
	DB         db.DBConfig      `mapstructure:"db"`
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
