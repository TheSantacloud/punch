package sync

import (
	"bytes"
	"os"
	"os/exec"
	"sort"
	"strings"
	"syscall"

	"github.com/dormunis/punch/pkg/models"
)

func GetConflicts(localSessions, remoteSessions []models.Session) (*bytes.Buffer, error) {
	sort.SliceStable(localSessions, func(i, j int) bool {
		return localSessions[i].Start.Before(*localSessions[j].Start)
	})
	localBuffer, err := models.SerializeSessionsToYAML(localSessions)
	if err != nil {
		return nil, err
	}

	sort.SliceStable(remoteSessions, func(i, j int) bool {
		return remoteSessions[i].Start.Before(*remoteSessions[j].Start)
	})
	remoteBuffer, err := models.SerializeSessionsToYAML(remoteSessions)
	if err != nil {
		return nil, err
	}
	return generateDiffBuffer(localBuffer, remoteBuffer)
}

func generateDiffBuffer(localBuffer, remoteBuffer *bytes.Buffer) (*bytes.Buffer, error) {
	localFile, err := os.CreateTemp("", "punch-local-*.yaml")
	if err != nil {
		return nil, err
	}
	localFile.Write(localBuffer.Bytes())
	defer os.Remove(localFile.Name())

	remoteFile, err := os.CreateTemp("", "punch-remote-*.yaml")
	if err != nil {
		return nil, err
	}
	remoteFile.Write(remoteBuffer.Bytes())
	defer os.Remove(remoteFile.Name())

	cmd := exec.Command("diff", "-D", "HEAD", localFile.Name(), remoteFile.Name())

	var out bytes.Buffer
	cmd.Stdout = &out
	err = cmd.Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if ws, ok := exitErr.Sys().(syscall.WaitStatus); ok {
				exitCode := ws.ExitStatus()
				if exitCode != 1 {
					return nil, err
				}
			}
		} else {
			return nil, err
		}
	}

	diffString := out.String()
	replaceToGitDiffStandard(&diffString)
	onlyDiffsString := keepOnlyDiffs(&diffString)

	return bytes.NewBufferString(onlyDiffsString), nil
}

func replaceToGitDiffStandard(diffString *string) {
	*diffString = strings.Replace(*diffString, "#ifndef HEAD", "<<<<<<< HEAD", -1)
	*diffString = strings.Replace(*diffString, "#else /* HEAD */", "=======", -1)
	*diffString = strings.Replace(*diffString, "#endif /* HEAD */", ">>>>>>> REMOTE", -1)
}

// this is kinda tied to the interactive edit being a yaml, but i dont care for now
func keepOnlyDiffs(diffString *string) string {
	lines := strings.Split(*diffString, "\n")
	var foundDiff bool

	relevantLines := ""
	currentObjectBuffer := ""
	foundDiff = false

	for _, line := range lines {
		currentObjectBuffer += line + "\n"
		if foundDiff {
			relevantLines += line + "\n"
		}
		if strings.HasPrefix(line, "<<<<<<<") {
			foundDiff = true
			relevantLines += currentObjectBuffer
		}
		if strings.HasPrefix(line, "---") {
			foundDiff = false
			currentObjectBuffer = ""
		}
	}

	if strings.HasSuffix(relevantLines, "---\n") {
		relevantLines = relevantLines[:len(relevantLines)-len("---\n")]
	}

	return relevantLines
}
