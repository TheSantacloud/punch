package models

import (
	"bytes"
	"fmt"

	"gopkg.in/yaml.v3"
)

type Company struct {
	Name     string `yaml:"name"`
	PPH      uint16 `yaml:"pph"`
	Currency string `yaml:"currency"`
}

func (c Company) String() string {
	return fmt.Sprintf("%s\t%d %s", c.Name, c.PPH, c.Currency)
}

func (c *Company) Serialize() (*bytes.Buffer, error) {
	var buf bytes.Buffer

	data, err := yaml.Marshal(c)
	if err != nil {
		return nil, err
	}

	buf.Write(data)

	return &buf, nil
}

func DeserializeCompanyFromYAML(buf *bytes.Buffer, company *Company) error {
	decoder := yaml.NewDecoder(buf)
	err := decoder.Decode(&company)
	if err != nil {
		return err
	}
	return nil
}
