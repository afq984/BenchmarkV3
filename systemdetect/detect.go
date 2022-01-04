package systemdetect

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"strconv"
	"strings"
)

func run(name string, arg ...string) (string, error) {
	cmd := exec.Command(name, arg...)
	b, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(bytes.TrimRight(b, "\n")), nil
}

func keyValueGet(r io.Reader, key string) (string, error) {
	s := bufio.NewScanner(r)
	for s.Scan() {
		parts := strings.SplitN(s.Text(), "=", 2)
		if len(parts) != 2 {
			continue
		}
		if parts[0] == key {
			value := strings.TrimSpace(parts[1])
			if value[0] == '"' {
				unquoted, err := strconv.Unquote(value)
				if err == nil {
					return unquoted, nil
				}
				return value, nil
			}
		}
	}
	if s.Err() != nil {
		return "", s.Err()
	}
	return "", fmt.Errorf("key %q not found", key)
}
