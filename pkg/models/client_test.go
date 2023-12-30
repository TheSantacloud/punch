package models

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClient_String(t *testing.T) {
	client := Client{Name: "Test Client", PPH: 100, Currency: "USD"}
	expectedString := "Test Client\t100 USD"
	assert.Equal(t, expectedString, client.String(), "Client String representation should match expected format")
}

func TestClient_Serialize(t *testing.T) {
	client := Client{Name: "Test Client", PPH: 100, Currency: "USD"}
	buf, err := client.Serialize()

	assert.NoError(t, err, "Serialize should not return an error")
	assert.NotNil(t, buf, "Serialized buffer should not be nil")
	assert.Contains(t, buf.String(), "Test Client", "Serialized data should contain the client's name")
}

func TestDeserializeClientFromYAML_ValidYAML(t *testing.T) {
	client := Client{Name: "Test Client", PPH: 100, Currency: "USD"}
	buf, _ := client.Serialize()

	var deserializedClient Client
	err := DeserializeClientFromYAML(buf, &deserializedClient)

	assert.NoError(t, err, "Deserialize should not return an error")
	assert.Equal(t, client, deserializedClient, "Deserialized client should match the original")
}

func TestDeserializeClientFromYAML_InvalidYAML(t *testing.T) {
	buf := bytes.NewBufferString("invalid yaml content")
	var client Client
	err := DeserializeClientFromYAML(buf, &client)

	assert.Error(t, err, "Deserialize should return an error for invalid YAML")
}
