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
		log.Fatal(err)
		return err
	}

	if err := tmpfile.Sync(); err != nil {
		log.Fatal(err)
		return err
	}

	if err := tmpfile.Close(); err != nil {
		log.Fatal(err)
		return err
	}

	editor := viper.GetString("settings.editor")
	cmd := exec.Command(editor, tmpfile.Name())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	err = cmd.Run()
	if err != nil {
		log.Fatal(err)
		return err
	}

	updatedData, err := os.ReadFile(tmpfile.Name())

	if err != nil {
		log.Fatal(err)
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
