package config

import (
	"os"
	"path/filepath"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

type intermidiateRemoteConfig struct {
	Type string `validate:"required"`
}

type IntermediateConfig struct {
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
		AutoSync    []string             `mapstructure:"autosync" validate:"omitempty,dive,oneof=start end"`
		SpreadSheet *SpreadsheetSettings `validate:"omitempty"`
	}
	Remotes map[string]intermidiateRemoteConfig `validate:"omitempty,dive"`
}

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
		AutoSync    []string             `mapstructure:"autosync" validate:"omitempty,dive,oneof=start end"`
		SpreadSheet *SpreadsheetSettings `validate:"omitempty"`
	}
	Remotes map[string]Remote `validate:"omitempty,dive"`
}

type Remote interface {
	Validate() error
	Type() string
}

type SpreadsheetSettings struct {
	ID      string `mapstructure:"spreadsheet_id" validate:"required"`
	Sheet   string `mapstructure:"sheet_name" validate:"required"`
	Columns struct {
		Company   string `validate:"required"`
		Date      string `validate:"required"`
		StartTime string `mapstructure:"start_time" validate:"required"`
		EndTime   string `mapstructure:"end_time" validate:"required"`
		TotalTime string `mapstructure:"total_time" validate:"required"`
		Note      string `validate:"required"` // TODO: make this optional
	} `validate:"required"`
}

func (s *SpreadsheetSettings) Validate() error {
	validate := validator.New()
	return validate.Struct(s)
}

func (s *SpreadsheetSettings) Type() string {
	return "spreadsheet"
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
	var ic IntermediateConfig
	if err := viper.Unmarshal(&ic); err != nil {
		return nil, err
	}

	conf := Config{
		Database: ic.Database,
		Settings: ic.Settings,
		Sync:     ic.Sync,
		Remotes:  make(map[string]Remote),
	}

	for key, value := range ic.Remotes {
		switch value.Type {
		case "spreadsheet":
			var spreadsheetRemote SpreadsheetSettings
			if err := viper.UnmarshalKey("remotes."+key, &spreadsheetRemote); err != nil {
				return nil, err
			}
			conf.Remotes[key] = &spreadsheetRemote
		}
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
