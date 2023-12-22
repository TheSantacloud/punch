package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type Config struct {
	Settings Settings
	Database Database
	Remotes  map[string]Remote
}

type Settings struct {
	Editor         string
	Currency       string   `mapstructure:"default_currency"`
	DefaultCompany string   `mapstructure:"default_company"`
	DefaultRemote  string   `mapstructure:"default_remote"`
	AutoSync       []string `mapstructure:"autosync" validate:"omitempty,dive,oneof=start end edit"`
}

type Database struct {
	Engine string `validate:"required,oneof=sqlite3"`
	Path   string `validate:"required"`
}

type Remote interface {
	Type() string // TODO: change to specific preset type RemoteType
	String() string
}

type SpreadsheetRemote struct {
	ID        string `mapstructure:"spreadsheet_id" validate:"required"`
	SheetName string `mapstructure:"sheet_name" validate:"required"`
	Columns   struct {
		Company   string `validate:"required"`
		Date      string `validate:"required"`
		StartTime string `mapstructure:"start_time" validate:"required"`
		EndTime   string `mapstructure:"end_time" validate:"required"`
		TotalTime string `mapstructure:"total_time" validate:"required"`
		Note      string `validate:"required"` // TODO: make this optional
	} `validate:"required"`
}

func (s *SpreadsheetRemote) Type() string {
	return "spreadsheet"
}

func (s *SpreadsheetRemote) String() string {
	return fmt.Sprintf("[%s] (%s)", s.Type(), s.ID)
}

func InitConfig() (*Config, error) {
	if err := setupDefaultConfig(); err != nil {
		return nil, err
	}

	conf, err := loadConfig()
	if err != nil {
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
	var intermediateConfig struct {
		Settings Settings
		Database Database
		Remotes  map[string]interface{}
	}
	if err := viper.Unmarshal(&intermediateConfig); err != nil {
		return nil, err
	}

	conf := &Config{
		Settings: intermediateConfig.Settings,
		Database: intermediateConfig.Database,
		Remotes:  make(map[string]Remote),
	}

	if err := unmarshalRemotes(viper.GetStringMap("remotes"), conf); err != nil {
		return nil, err
	}

	return conf, nil
}

func unmarshalRemotes(remoteMap map[string]interface{}, conf *Config) error {
	conf.Remotes = make(map[string]Remote)

	for key, value := range remoteMap {
		remoteConfig := value.(map[string]interface{})

		remoteType, ok := remoteConfig["type"].(string)
		if !ok {
			return fmt.Errorf("remote '%s' missing type", key)
		}

		switch remoteType {
		case "spreadsheet":
			var remote SpreadsheetRemote
			if err := viper.UnmarshalKey(fmt.Sprintf("remotes.%s", key), &remote); err != nil {
				return err
			}
			conf.Remotes[key] = &remote
		default:
			return fmt.Errorf("unknown remote type '%s'", remoteType)
		}
	}

	return nil
}
