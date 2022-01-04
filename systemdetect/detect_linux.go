package systemdetect

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"syscall"
)

type socket struct {
	name string
	id   int
}

type cpuInfo struct {
	sockets int
	cores   int
	threads int
}

func Cpu() (string, error) {
	f, err := os.Open("/proc/cpuinfo")
	if err != nil {
		return "", err
	}
	defer f.Close()

	sockets := make(map[socket]cpuInfo)
	s := bufio.NewScanner(f)
	var (
		name    string
		id      int
		cores   int
		threads int
	)
	for s.Scan() {
		line := s.Text()
		if line == "" {
			sockets[socket{name, id}] = cpuInfo{1, cores, threads}
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			return "", fmt.Errorf("invalid line: %q", line)
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		valueNum, valueErr := strconv.Atoi(value)
		bad := false
		switch key {
		case "model name":
			name = value
		case "physical id":
			if valueErr != nil {
				bad = true
			}
			id = valueNum
		case "siblings":
			if valueErr != nil {
				bad = true
			}
			threads = valueNum
		case "cpu cores":
			if valueErr != nil {
				bad = true
			}
			cores = valueNum
		}
		if bad {
			return "", fmt.Errorf("bad line: %q: %v", line, err)
		}
	}
	if s.Err() != nil {
		return "", fmt.Errorf("failed reading /proc/cpuinfo: %w", s.Err())
	}
	if len(sockets) == 0 {
		return "", fmt.Errorf("failed to parse /proc/cpuinfo")
	}

	models := make(map[string]cpuInfo)
	for s, i := range sockets {
		c := models[s.name]
		models[s.name] = cpuInfo{
			c.sockets + i.sockets,
			c.cores + i.cores,
			c.threads + i.threads,
		}
	}
	var parts []string
	for name, info := range models {
		var counts string
		if info.sockets > 1 {
			counts = fmt.Sprintf("%dS%dC%dT", info.sockets, info.cores, info.threads)
		} else {
			counts = fmt.Sprintf("%dC%dT", info.cores, info.threads)
		}
		parts = append(parts, fmt.Sprintf("%s %q", counts, name))
	}
	return strings.Join(parts, ";"), nil
}

func Memory() (int64, error) {
	var info syscall.Sysinfo_t
	err := syscall.Sysinfo(&info)
	if err != nil {
		return -1, err
	}
	return int64(info.Totalram), nil
}

func osReleaseName() (name string, err error) {
	f, err := os.Open("/etc/os-release")
	if err != nil {
		return
	}
	defer f.Close()

	name, err = keyValueGet(f, "PRETTY_NAME")
	if err == nil {
		return
	}

	f.Seek(0, io.SeekStart)
	return keyValueGet(f, "NAME")
}

func stringOrNone(s string, err error) string {
	if err != nil {
		return "none"
	}
	return s
}

func Misc() (string, error) {
	return fmt.Sprintf("%s;%s;%s",
		stringOrNone(osReleaseName()),
		stringOrNone(run("systemd-detect-virt", "--vm")),
		stringOrNone(run("systemd-detect-virt", "--container")),
	), nil
}
