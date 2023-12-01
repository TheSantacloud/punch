package config

import (
	"os"
	"path/filepath"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

type Config struct {
	Database struct {
		Engine string `validate:"required,engine"`
		Path   string `validate:"required"`
	}
	Settings struct {
		Currency       string `validate:"required"`
		Editor         string `validate:"required"`
		DefaultCompany string `mapstructure:"default_company"`
	}
	Sync struct {
		Engine      string               `validate:"omitempty,oneof=spreadsheet"`
		SyncActions []string             `mapstructure:"sync_actions" validate:"omitempty,dive,oneof=start end"`
		SpreadSheet *SpreadsheetSettings `validate:"omitempty"`
	}
}

type SpreadsheetSettings struct {
	ID      string `validate:"required"`
	Sheet   string `validate:"required"`
	Columns struct {
		Company   string `validate:"required"`
		Date      string `validate:"required"`
		StartTime string `mapstructure:"start_time" validate:"required"`
		EndTime   string `mapstructure:"end_time" validate:"required"`
		TotalTime string `mapstructure:"total_time" validate:"required"`
	} `validate:"required"`
}

func InitConfig() (*Config, error) {
	if err := setupDefaultConfig(); err != nil {
		return nil, err
	}

	conf, err := loadConfig()
	if err != nil {
		return nil, err
	}

	if err := validateConfig(conf); err != nil {
		return nil, err
	}

	return conf, nil
}

func setupDefaultConfig() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	configPath := filepath.Join(home, ".punch")
	if err := os.MkdirAll(configPath, os.ModePerm); err != nil {
		return err
	}

	viper.AddConfigPath(configPath)
	viper.SetConfigName("config")
	viper.SetConfigType("toml")

	viper.SetDefault("database.engine", "sqlite3")
	viper.SetDefault("database.path", filepath.Join(configPath, "punch.db"))

	viper.SetDefault("settings.default_currency", "USD")
	viper.SetDefault("settings.editor", "vim")

	if viper.IsSet("sync.engine") {
		viper.SetDefault("sync.sync_actions", []string{"end"})
	}

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return viper.SafeWriteConfigAs(filepath.Join(configPath, "config.toml"))
		}
		return err
	}

	return nil
}

func loadConfig() (*Config, error) {
	var conf Config
	if err := viper.Unmarshal(&conf); err != nil {
		return nil, err
	}
	return &conf, nil
}

func validateConfig(conf *Config) error {
	validate := validator.New()
	validate.RegisterValidation("engine", validateEngine)
	return validate.Struct(conf)
}

func validateEngine(fl validator.FieldLevel) bool {
	engine := fl.Field().String()
	return engine == "sqlite3"
}
