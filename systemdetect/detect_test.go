package systemdetect

import (
	"bytes"
	"log"
	"strings"
	"testing"
)

func TestCpu(t *testing.T) {
	cpu, err := Cpu()
	if err != nil {
		t.Fatal("cannot get CPU information")
	}
	log.Println("detected CPU:", cpu)
	// Guard against degenerate detections such as `0C0T ""`, which is what the
	// arm64 path used to produce because /proc/cpuinfo has no x86 topology
	// fields. A real CPU has at least one core and a non-empty name.
	if strings.HasPrefix(cpu, "0C") || strings.HasSuffix(cpu, `""`) {
		t.Errorf("degenerate CPU detection: %q", cpu)
	}
}

func TestRam(t *testing.T) {
	memory, err := Memory()
	if err != nil {
		t.Fatal("cannt get RAM information")
	}
	log.Println("detected RAM:", memory)
}

func TestMisc(t *testing.T) {
	misc, err := Misc()
	if err != nil {
		t.Fatal("cannot get misc information")
	}
	log.Println("detected misc:", misc)
}

func TestKeyValueGet(t *testing.T) {
	const etcOsRelease = `NAME="foo bar"
PRETTY_NAME="hello world"
ID=hi
`

	b := bytes.NewBufferString(etcOsRelease)
	value, err := keyValueGet(b, "PRETTY_NAME")
	if err != nil {
		t.Errorf("keyValueGet failed: %v", err)
	}
	if value != "hello world" {
		t.Errorf("unexpected value %q", value)
	}
}
