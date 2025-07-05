package editor

import (
	"bytes"
	"log"
	"os"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestInteractiveEdit_InvalidEditorExecutable(t *testing.T) {
	buffer := bytes.NewBufferString("test content")
	viper.Set("settings.editor", "invalid_editor")

	err := InteractiveEdit(buffer, "txt")
	assert.Error(t, err, "should error out due to invalid editor executable")
}

func TestInteractiveEdit_SameChecksum(t *testing.T) {
	buffer := bytes.NewBufferString("test content")
	viper.Set("settings.editor", "echo")

	err := InteractiveEdit(buffer, "txt")
	assert.Error(t, err, "should error due to unchanged content")
}

func TestInteractiveEdit_ReturnsErrorWhenNoChangesAreMade(t *testing.T) {
	buffer := bytes.NewBufferString("test content")
	viper.Set("settings.editor", "echo")

	err := InteractiveEdit(buffer, "txt")
	assert.Error(t, err, "should error due to unchanged content")
}

func TestInteractiveEdit_InvalidFile(t *testing.T) {
	buffer := bytes.NewBufferString("test content")
	viper.Set("settings.editor", "echo")

	err := InteractiveEdit(buffer, "txt")
	assert.Error(t, err, "should error due to invalid file path")
}

func TestInteractiveEdit_ChangedChecksum(t *testing.T) {
	editorScript := []byte(`#!/bin/bash
echo "modified content" > $1`)
	tmpScriptFile, err := os.CreateTemp("", "test_editor_*.sh")
	assert.NoError(t, err, "Failed to create temporary script file")
	defer func() {
		if err := os.Remove(tmpScriptFile.Name()); err != nil {
			log.Printf("failed to remove temp file %q: %v", tmpScriptFile.Name(), err)
		}
	}()

	_, err = tmpScriptFile.Write(editorScript)
	assert.NoError(t, err, "Failed to write to temporary script file")
	assert.NoError(t, tmpScriptFile.Close(), "Failed to close temporary script file")

	assert.NoError(t, os.Chmod(tmpScriptFile.Name(), 0700), "Failed to make script executable")

	viper.Set("settings.editor", tmpScriptFile.Name())

	buffer := bytes.NewBufferString("test content")

	err = InteractiveEdit(buffer, "txt")
	assert.NoError(t, err, "InteractiveEdit should not error for changed content")

	modifiedContent := buffer.String()
	assert.Equal(t, "modified content\n", modifiedContent, "Buffer content should be modified by the script")
}
