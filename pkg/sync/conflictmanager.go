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
	if strings.Contains(diffString, "#ifndef HEAD") {
		onlyDiffsString := keepOnlyDiffObjects(&diffString)
		diffString = onlyDiffsString
	}
	replaceToGitDiffStandard(&diffString)
	return bytes.NewBufferString(diffString), nil
}

func replaceToGitDiffStandard(diffString *string) {
	if strings.Contains(*diffString, "ifdef") {
		*diffString = strings.Replace(*diffString, "#ifdef HEAD", "<<<<<<< HEAD\n=======", -1)
		*diffString = strings.Replace(*diffString, "#endif /* HEAD */", ">>>>>>> REMOTE", -1)
	} else {
		*diffString = strings.Replace(*diffString, "#ifndef HEAD", "<<<<<<< HEAD", -1)
		*diffString = strings.Replace(*diffString, "#else /* HEAD */", "=======", -1)
		*diffString = strings.Replace(*diffString, "#endif /* HEAD */", ">>>>>>> REMOTE", -1)
	}
}

func keepOnlyDiffObjects(diffString *string) string {
	// separate commented header from content
	diffComponents := strings.Split(*diffString, "\n\n")
	header := diffComponents[0]
	content := strings.Join(diffComponents[1:], "\n\n")

	objects := strings.Split(content, models.YAML_SERIALIZATION_SEPARATOR)
	relevantObjects := ""
	for _, object := range objects {
		if strings.HasPrefix(object, "#ifndef") {
			relevantObjects += object + models.YAML_SERIALIZATION_SEPARATOR
		}
	}
	if strings.HasSuffix(relevantObjects, models.YAML_SERIALIZATION_SEPARATOR) {
		relevantObjects = relevantObjects[:len(relevantObjects)-
			len(models.YAML_SERIALIZATION_SEPARATOR)]
	}

	return header + "\n\n" + relevantObjects
}

func DetectDeletedSessions(sessions *[]models.Session, editedSessions *[]models.Session) []models.Session {
	deletedSessions := make([]models.Session, 0)
	for _, session := range *sessions {
		found := false
		for _, editedSession := range *editedSessions {
			if *session.ID == *editedSession.ID {
				found = true
				break
			}
		}
		if !found {
			deletedSessions = append(deletedSessions, session)
		}
	}
	return deletedSessions
}
