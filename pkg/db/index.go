package db

import (
	"database/sql"
	"fmt"

	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"transcode-demo/pkg/utils"
)

// DBConfig is config for DB
type DBConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"name"`
}

// SetDefault sets default values for log config
func (c *DBConfig) SetDefault(parentKey string) {
	viper.SetDefault(utils.JoinConfigKeys(parentKey, "host"), "127.0.0.1")
	viper.SetDefault(utils.JoinConfigKeys(parentKey, "port"), "5432")
	viper.SetDefault(utils.JoinConfigKeys(parentKey, "user"), "dev")
	viper.SetDefault(utils.JoinConfigKeys(parentKey, "password"), "dev")
	viper.SetDefault(utils.JoinConfigKeys(parentKey, "name"), "tdapp")
}

// ConnectionString returns connection string
func (c *DBConfig) ConnectionString() string {
	psqlInfo := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		c.Host, c.Port, c.User, c.Password, c.DBName,
	)
	return psqlInfo
}

func NewGormDB(db *sql.DB) (*gorm.DB, error) {
	gormDB, err := gorm.Open(
		postgres.New(postgres.Config{
			Conn: db,
		}), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to open gorm db: %w", err)
	}
	return gormDB, nil
}
