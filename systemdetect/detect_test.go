package systemdetect

import (
	"bytes"
	"log"
	"testing"
)

func TestCpu(t *testing.T) {
	cpu, err := Cpu()
	if err != nil {
		t.Fatal("cannot get CPU information")
	}
	log.Println("detected CPU:", cpu)
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
