package database

import (
	"fmt"

	_ "github.com/BurntSushi/toml"
	"github.com/spf13/viper"
)

type Company struct {
	Name string `toml:"name"`
	PPH  int32  `toml:"pph"`
}

func (c Company) String() string {
	cur := viper.GetString("settings.currency")
	return fmt.Sprintf("%s\t%s%d", c.Name, cur, c.PPH)
}
