package systemdetect

import (
	"strings"
	"testing"
)

// blocks helper: repeat a processor block n times with a rising processor index,
// keeping the remaining fields identical.
func repeatBlock(n int, body string) string {
	var b strings.Builder
	for i := 0; i < n; i++ {
		b.WriteString("processor\t: ")
		b.WriteString(itoa(i))
		b.WriteByte('\n')
		b.WriteString(body)
		b.WriteByte('\n')
	}
	return b.String()
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	var digits []byte
	for i > 0 {
		digits = append([]byte{byte('0' + i%10)}, digits...)
		i /= 10
	}
	return string(digits)
}

func TestParseCpuinfoX86SingleSocket(t *testing.T) {
	// 2 physical cores, 4 threads, one socket.
	in := repeatBlock(4, `vendor_id	: GenuineIntel
model name	: Intel(R) Core(TM) i5-8250U CPU @ 1.60GHz
physical id	: 0
siblings	: 4
cpu cores	: 2
`)
	got, err := parseCpuinfo(strings.NewReader(in))
	if err != nil {
		t.Fatal(err)
	}
	want := `2C4T "Intel(R) Core(TM) i5-8250U CPU @ 1.60GHz"`
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestParseCpuinfoX86DualSocket(t *testing.T) {
	// Two sockets, each 2 cores / 4 threads -> 2S4C8T.
	body := func(id string) string {
		return `model name	: Intel(R) Xeon(R) CPU
physical id	: ` + id + `
siblings	: 4
cpu cores	: 2
`
	}
	in := "processor\t: 0\n" + body("0") + "\n" +
		"processor\t: 1\n" + body("0") + "\n" +
		"processor\t: 2\n" + body("0") + "\n" +
		"processor\t: 3\n" + body("0") + "\n" +
		"processor\t: 4\n" + body("1") + "\n" +
		"processor\t: 5\n" + body("1") + "\n" +
		"processor\t: 6\n" + body("1") + "\n" +
		"processor\t: 7\n" + body("1") + "\n"
	got, err := parseCpuinfo(strings.NewReader(in))
	if err != nil {
		t.Fatal(err)
	}
	want := `2S4C8T "Intel(R) Xeon(R) CPU"`
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

// snapdragon888Cpuinfo mirrors the arm64 /proc/cpuinfo layout: no model name,
// only implementer/part ids, one block per logical core.
const snapdragon888Cpuinfo = `processor	: 0
BogoMIPS	: 49.15
CPU implementer	: 0x41
CPU architecture: 8
CPU variant	: 0x1
CPU part	: 0xd44
CPU revision	: 0

processor	: 1
CPU implementer	: 0x41
CPU part	: 0xd41

processor	: 2
CPU implementer	: 0x41
CPU part	: 0xd05

processor	: 3
CPU implementer	: 0x41
CPU part	: 0xd05

processor	: 4
CPU implementer	: 0x41
CPU part	: 0xd05

processor	: 5
CPU implementer	: 0x41
CPU part	: 0xd05

processor	: 6
CPU implementer	: 0x41
CPU part	: 0xd41

processor	: 7
CPU implementer	: 0x41
CPU part	: 0xd41
`

func TestParseCpuinfoARMBigLittle(t *testing.T) {
	got, err := parseCpuinfo(strings.NewReader(snapdragon888Cpuinfo))
	if err != nil {
		t.Fatal(err)
	}
	// Clusters ordered by descending part id: X1 (0xd44), A78 (0xd41), A55 (0xd05).
	want := `8C8T "1x Cortex-X1 + 3x Cortex-A78 + 4x Cortex-A55"`
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestParseCpuinfoARMHomogeneous(t *testing.T) {
	// A homogeneous server part collapses to a single name (no "Nx" prefix).
	in := repeatBlock(4, `CPU implementer	: 0x41
CPU part	: 0xd0c
`)
	got, err := parseCpuinfo(strings.NewReader(in))
	if err != nil {
		t.Fatal(err)
	}
	want := `4C4T "Neoverse-N1"`
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestParseCpuinfoARMUnknownPart(t *testing.T) {
	// An unrecognized part falls back to implementer + raw id, never empty.
	in := repeatBlock(2, `CPU implementer	: 0x41
CPU part	: 0xfff
`)
	got, err := parseCpuinfo(strings.NewReader(in))
	if err != nil {
		t.Fatal(err)
	}
	want := `2C2T "ARM part 0xfff"`
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestParseCpuinfoQualcommImplementer(t *testing.T) {
	in := repeatBlock(2, `CPU implementer	: 0x51
CPU part	: 0x804
`)
	got, err := parseCpuinfo(strings.NewReader(in))
	if err != nil {
		t.Fatal(err)
	}
	want := `2C2T "Kryo-4XX-Gold"`
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestParseCpuinfoEmpty(t *testing.T) {
	if _, err := parseCpuinfo(strings.NewReader("")); err == nil {
		t.Error("expected error for empty /proc/cpuinfo")
	}
}
