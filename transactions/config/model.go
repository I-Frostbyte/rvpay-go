package model

import (
	"errors"
	"fmt"
	"os"

	"github.com/ardanlabs/conf/v3"
	"github.com/joho/godotenv"
)

type Config struct {
	LogLevel      string `conf:"env:LOG_LEVEL,default:debug"`
	ListenPort    string `conf:"env:LISTEN_PORT,required"`
	MigrationPath string `conf:"env:MIGRATION_PATH,required"`
	RunMigrations bool   `conf:"env:RUN_MIGRATIONS,default:true"`

	DB DBConfig
}

type DBConfig struct {
	DBUser      string `conf:"env:DB_USER,required"`
	DBPassword  string `conf:"env:DB_PASSWORD,required"`
	DBHost      string `conf:"env:DB_HOST,required"`
	DBPort      uint   `conf:"env:DB_PORT,required"`
	DBName      string `conf:"env:DB_NAME,required"`
	TLSDisabled bool   `conf:"env:DB_TLS_DISABLED,default:false"`
}

func (c *Config) LoadConfig() error {
	if _, err := os.Stat(".env"); err == nil {
		if err := godotenv.Load(); err != nil {
			return fmt.Errorf("error loading .env file: %w", err)
		}
	}

	if _, err := conf.Parse("", c); err != nil {
		if errors.Is(err, conf.ErrHelpWanted) {
			return err
		}
		return fmt.Errorf("error parsing config: %w", err)
	}

	return nil
}
