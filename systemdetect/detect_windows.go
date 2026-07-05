package systemdetect

import (
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"unsafe"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

// Cpu reports the processor as "<physical>C<logical>T \"<name>\"", matching the
// format the Linux and darwin detectors use.
func Cpu() (string, error) {
	physical := physicalCores()
	logical := runtime.NumCPU()
	if physical == 0 {
		physical = logical
	}
	return fmt.Sprintf("%dC%dT %q", physical, logical, cpuName()), nil
}

func cpuName() string {
	k, err := registry.OpenKey(registry.LOCAL_MACHINE,
		`HARDWARE\DESCRIPTION\System\CentralProcessor\0`, registry.QUERY_VALUE)
	if err != nil {
		return "unknown"
	}
	defer k.Close()
	name, _, err := k.GetStringValue("ProcessorNameString")
	if err != nil {
		return "unknown"
	}
	if name = strings.TrimSpace(name); name == "" {
		return "unknown"
	}
	return name
}

// physicalCores counts RelationProcessorCore entries from
// GetLogicalProcessorInformation, or returns 0 if it cannot be determined.
func physicalCores() int {
	proc := windows.NewLazySystemDLL("kernel32.dll").NewProc("GetLogicalProcessorInformation")

	var size uint32
	proc.Call(0, uintptr(unsafe.Pointer(&size))) // fails, fills in the required size
	if size == 0 {
		return 0
	}
	buf := make([]byte, size)
	if r, _, _ := proc.Call(uintptr(unsafe.Pointer(&buf[0])), uintptr(unsafe.Pointer(&size))); r == 0 {
		return 0
	}

	// SYSTEM_LOGICAL_PROCESSOR_INFORMATION on 64-bit Windows is 32 bytes:
	// ProcessorMask (8) + Relationship (4) + padding (4) + union (16). The
	// Relationship enum sits at offset 8; RelationProcessorCore == 0.
	const entrySize, relOffset, relationProcessorCore = 32, 8, 0
	count := 0
	for off := 0; off+entrySize <= int(size); off += entrySize {
		if *(*uint32)(unsafe.Pointer(&buf[off+relOffset])) == relationProcessorCore {
			count++
		}
	}
	return count
}

// memoryStatusEx mirrors MEMORYSTATUSEX (kernel32).
type memoryStatusEx struct {
	Length               uint32
	MemoryLoad           uint32
	TotalPhys            uint64
	AvailPhys            uint64
	TotalPageFile        uint64
	AvailPageFile        uint64
	TotalVirtual         uint64
	AvailVirtual         uint64
	AvailExtendedVirtual uint64
}

func Memory() (int64, error) {
	proc := windows.NewLazySystemDLL("kernel32.dll").NewProc("GlobalMemoryStatusEx")
	var m memoryStatusEx
	m.Length = uint32(unsafe.Sizeof(m))
	if r, _, err := proc.Call(uintptr(unsafe.Pointer(&m))); r == 0 {
		return -1, err
	}
	return int64(m.TotalPhys), nil
}

func Misc() (string, error) {
	return fmt.Sprintf("%s;none;none", osName()), nil
}

func osName() string {
	k, err := registry.OpenKey(registry.LOCAL_MACHINE,
		`SOFTWARE\Microsoft\Windows NT\CurrentVersion`, registry.QUERY_VALUE)
	if err != nil {
		return "Windows"
	}
	defer k.Close()
	product, _, _ := k.GetStringValue("ProductName")
	display, _, _ := k.GetStringValue("DisplayVersion")
	build, _, _ := k.GetStringValue("CurrentBuild")

	// ProductName still reports "Windows 10" on Windows 11; correct it by build.
	if n, err := strconv.Atoi(build); err == nil && n >= 22000 {
		product = strings.Replace(product, "Windows 10", "Windows 11", 1)
	}
	name := strings.TrimSpace(product)
	if display != "" {
		name += " " + display
	}
	if build != "" {
		name += " (build " + build + ")"
	}
	if name == "" {
		name = "Windows"
	}
	return name
}
