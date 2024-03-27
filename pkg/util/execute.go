package util

import (
	"encoding/json"
	"fmt"
	"os/exec"

	"github.com/pkg/errors"
)

func LocalExecutefJsonMarshal(obj interface{},
	fs string, a ...interface{}) error {
	out, err := LocalExecutef(fs, a...)
	if err != nil {
		return errors.Wrapf(err, "%s", fmt.Sprintf(fs, a...))
	}
	return json.Unmarshal([]byte(out), obj)
}

func LocalExecute(cmd string) (string, error) {
	out, err := exec.Command("sh", "-c", cmd).Output()
	if err != nil {
		return "", errors.Wrapf(err, "%s", cmd)
	}
	return string(out), nil
}

func LocalExecutef(fs string, a ...interface{}) (string, error) {
	return LocalExecute(fmt.Sprintf(fs, a...))
}
