package systemdetect

import (
	"fmt"
	"strconv"
)

func Cpu() (string, error) {
	cpu, err := run("sysctl", "-n", "machdep.cpu.brand_string")
	if err != nil {
		return "", err
	}

	cores, err := run("sysctl", "-n", "machdep.cpu.core_count")
	if err != nil {
		return "", err
	}

	threads, err := run("sysctl", "-n", "machdep.cpu.thread_count")
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%sC%sT %q", cores, threads, cpu), nil
}

func Memory() (int64, error) {
	s, err := run("sysctl", "-n", "hw.memsize")
	if err != nil {
		return -1, err
	}

	return strconv.ParseInt(s, 10, 64)
}

func Misc() (string, error) {
	return run("sysctl", "-n", "hw.model")
}
