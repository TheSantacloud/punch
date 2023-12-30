package editor

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"

	"crypto/sha256"
	"encoding/hex"

	"github.com/spf13/viper"
)

func InteractiveEdit(buffer *bytes.Buffer, filetype string) error {
	initialChecksum := sha256.Sum256(buffer.Bytes())
	initialChecksumStr := hex.EncodeToString(initialChecksum[:])

	tmpfile, err := os.CreateTemp("", fmt.Sprintf("editbuffer.*.%s", filetype))
	if err != nil {
		return err
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write(buffer.Bytes()); err != nil {
		return err
	}
	if err := tmpfile.Sync(); err != nil {
		return err
	}
	if err := tmpfile.Close(); err != nil {
		return err
	}

	editor := viper.GetString("settings.editor")

	if _, err := exec.LookPath(editor); err != nil {
		return fmt.Errorf("%s executable not found. "+
			"Change the `editor` field in your config %s",
			editor, viper.GetViper().ConfigFileUsed())
	}

	cmd := exec.Command(editor, tmpfile.Name())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	err = cmd.Run()
	if err != nil {
		return err
	}

	updatedData, err := os.ReadFile(tmpfile.Name())
	if err != nil {
		return err
	}

	postEditChecksum := sha256.Sum256(updatedData)
	postEditChecksumStr := hex.EncodeToString(postEditChecksum[:])

	if initialChecksumStr == postEditChecksumStr {
		return fmt.Errorf("No changes made")
	}

	buffer.Reset()
	buffer.Write(updatedData)
	return nil
}
