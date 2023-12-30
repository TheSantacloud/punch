package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-playground/validator"
	"github.com/spf13/viper"
)

type Config struct {
	Settings Settings
	Database Database
	Remotes  map[string]Remote
}

type Settings struct {
	Editor        string
	Currency      string   `mapstructure:"default_currency"`
	DefaultRemote string   `mapstructure:"default_remote"`
	DefaultClient string   `mapstructure:"default_client"` // TODO: this might be better as a databased setting
	AutoSync      []string `mapstructure:"autosync" validate:"omitempty,dive,oneof=start end edit delete"`
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
	ID                     string   `mapstructure:"spreadsheet_id" validate:"required"`
	SheetName              string   `mapstructure:"sheet_name" validate:"required"`
	ServiceAccountJsonPath string   `mapstructure:"service_account_json_path"`
	Columns                struct { // TODO: this is duplicated in sheet.go, find a better way
		ID        string `validate:"required"`
		Client    string `validate:"required"`
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

func InitConfig(configPaths ...string) (*Config, error) {
	var configPath string
	if len(configPaths) > 0 {
		configPath = configPaths[0]
	} else {
		configPath = ""
	}
	if err := setupDefaultConfig(configPath); err != nil {
		return nil, err
	}

	conf, err := loadConfig()
	if err != nil {
		return nil, err
	}

	return conf, nil
}

func setupDefaultConfig(optionalConfigPath string) error {
	configPath := determineConfigPath(optionalConfigPath)

	if err := os.MkdirAll(configPath, os.ModePerm); err != nil {
		return err
	}

	viper.AddConfigPath(configPath)
	viper.SetConfigName("config")
	viper.SetConfigType("toml")

	viper.SetDefault("database.engine", "sqlite3")
	viper.SetDefault("database.path", filepath.Join(configPath, "punch.db"))

	viper.SetDefault("settings.default_currency", "USD")
	viper.SetDefault("settings.editor", "vi")

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

func determineConfigPath(optionalConfigPath string) string {
	if optionalConfigPath != "" {
		return optionalConfigPath
	}

	if envConfigPath, ok := os.LookupEnv("PUNCH_CONFIG_PATH"); ok {
		return envConfigPath
	}

	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	return filepath.Join(home, ".punch")
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

	validate := validator.New()
	validate.RegisterValidation("autosync_requires_default_remote",
		validateAutoSync)
	validate.RegisterValidation("default_remote must have a coresponding remote set",
		validateDefaultRemoteExistsWithinRemotes)

	err := validate.Struct(conf)
	if err != nil {
		fmt.Printf("Validation errors: %v\n", err)
	}

	return conf, nil
}

func validateAutoSync(fl validator.FieldLevel) bool {
	settings := fl.Parent().Interface().(Settings)
	autoSync := settings.AutoSync
	defaultRemote := settings.DefaultRemote

	return len(autoSync) == 0 || (len(autoSync) > 0 && defaultRemote != "")
}

func validateDefaultRemoteExistsWithinRemotes(fl validator.FieldLevel) bool {
	config := fl.Parent().Interface().(Config)
	settings := config.Settings
	defaultRemote := settings.DefaultRemote

	if defaultRemote == "" {
		return true
	}

	for _, remote := range config.Remotes {
		if remote.String() == defaultRemote {
			return true
		}
	}
	return false
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

			// set default values
			if remote.ServiceAccountJsonPath == "" {
				remote.ServiceAccountJsonPath = filepath.Join(determineConfigPath(""), "service-account.json")
			}

			conf.Remotes[key] = &remote
		default:
			return fmt.Errorf("unknown remote type '%s'", remoteType)
		}
	}

	return nil
}
