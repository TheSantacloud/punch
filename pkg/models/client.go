package models

import (
	"bytes"
	"fmt"

	"gopkg.in/yaml.v3"
)

type Client struct {
	Name     string `yaml:"name"`
	PPH      uint16 `yaml:"pph"`
	Currency string `yaml:"currency"`
}

func (c Client) String() string {
	return fmt.Sprintf("%s\t%d %s", c.Name, c.PPH, c.Currency)
}

func (c *Client) Serialize() (*bytes.Buffer, error) {
	var buf bytes.Buffer

	data, err := yaml.Marshal(c)
	if err != nil {
		return nil, err
	}

	buf.Write(data)

	return &buf, nil
}

func DeserializeClientFromYAML(buf *bytes.Buffer, client *Client) error {
	decoder := yaml.NewDecoder(buf)
	err := decoder.Decode(&client)
	if err != nil {
		return err
	}
	return nil
}
