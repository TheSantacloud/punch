package database

import (
	"fmt"

	_ "github.com/BurntSushi/toml"
)

type Company struct {
	Name     string `toml:"name"`
	PPH      int32  `toml:"pph"`
	Currency string `toml:"currency"`
}

func (c Company) String() string {
	return fmt.Sprintf("%s\t%d %s", c.Name, c.PPH, c.Currency)
}
