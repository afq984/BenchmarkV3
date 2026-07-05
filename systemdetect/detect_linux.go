package systemdetect

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
	"syscall"
)

// cpuinfoBlocks parses /proc/cpuinfo into one map per logical processor. Blocks
// are separated by blank lines; each "key : value" pair becomes an entry with
// the key and value trimmed. The final block is flushed at EOF even when the
// file does not end with a blank line.
func cpuinfoBlocks(r io.Reader) ([]map[string]string, error) {
	var blocks []map[string]string
	cur := map[string]string{}
	flush := func() {
		if len(cur) > 0 {
			blocks = append(blocks, cur)
			cur = map[string]string{}
		}
	}

	s := bufio.NewScanner(r)
	for s.Scan() {
		line := s.Text()
		if strings.TrimSpace(line) == "" {
			flush()
			continue
		}
		key, value, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		cur[strings.TrimSpace(key)] = strings.TrimSpace(value)
	}
	if s.Err() != nil {
		return nil, fmt.Errorf("failed reading /proc/cpuinfo: %w", s.Err())
	}
	flush()
	return blocks, nil
}

func Cpu() (string, error) {
	f, err := os.Open("/proc/cpuinfo")
	if err != nil {
		return "", err
	}
	defer f.Close()

	return parseCpuinfo(f)
}

// parseCpuinfo turns the contents of /proc/cpuinfo into a human readable CPU
// description. Different architectures expose wildly different fields, so it
// dispatches on the keys that are actually present:
//
//   - x86 exposes "model name" plus "physical id"/"cpu cores"/"siblings", which
//     describe the socket/core/thread topology directly.
//   - arm64 exposes no model name at all, only per-core "CPU implementer" and
//     "CPU part" ids that must be decoded into core names (big.LITTLE systems
//     report several distinct parts).
func parseCpuinfo(r io.Reader) (string, error) {
	blocks, err := cpuinfoBlocks(r)
	if err != nil {
		return "", err
	}
	if len(blocks) == 0 {
		return "", fmt.Errorf("failed to parse /proc/cpuinfo")
	}

	switch {
	case anyHasKey(blocks, "model name"):
		return cpuX86(blocks)
	case anyHasKey(blocks, "CPU part"):
		return cpuARM(blocks)
	default:
		// Unknown architecture: at least report the logical CPU count.
		n := len(blocks)
		return fmt.Sprintf("%dC%dT %q", n, n, "unknown"), nil
	}
}

func anyHasKey(blocks []map[string]string, key string) bool {
	for _, b := range blocks {
		if _, ok := b[key]; ok {
			return true
		}
	}
	return false
}

// atoiField parses an integer field. A missing field is treated as 0 (not an
// error) so that CPUs which omit a field still parse; a present but malformed
// value is reported as an error.
func atoiField(b map[string]string, key string) (int, error) {
	v, ok := b[key]
	if !ok || v == "" {
		return 0, nil
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return 0, fmt.Errorf("bad %q: %q", key, v)
	}
	return n, nil
}

type socket struct {
	name string
	id   int
}

type cpuInfo struct {
	sockets int
	cores   int
	threads int
}

// cpuX86 aggregates the classic x86 topology: "siblings" is the thread count
// per socket, "cpu cores" the core count per socket, and "physical id" the
// socket number. Identical model names across sockets are summed together.
func cpuX86(blocks []map[string]string) (string, error) {
	sockets := make(map[socket]cpuInfo)
	for _, b := range blocks {
		name := b["model name"]
		id, err := atoiField(b, "physical id")
		if err != nil {
			return "", err
		}
		cores, err := atoiField(b, "cpu cores")
		if err != nil {
			return "", err
		}
		threads, err := atoiField(b, "siblings")
		if err != nil {
			return "", err
		}
		sockets[socket{name, id}] = cpuInfo{1, cores, threads}
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

	// Sort model names so the output is deterministic regardless of map order.
	names := make([]string, 0, len(models))
	for name := range models {
		names = append(names, name)
	}
	sort.Strings(names)

	var parts []string
	for _, name := range names {
		info := models[name]
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

// cpuCluster is a group of identical arm64 cores (same implementer + part).
type cpuCluster struct {
	impl int
	part int
}

// cpuARM decodes the arm64 /proc/cpuinfo, which has no model name — only the
// "CPU implementer" and "CPU part" ids of each logical core. arm64 has no SMT,
// so the logical core count is also the physical core count. big.LITTLE SoCs
// report several distinct parts; each becomes a "Nx <core name>" group.
func cpuARM(blocks []map[string]string) (string, error) {
	counts := make(map[cpuCluster]int)
	var order []cpuCluster
	total := 0
	for _, b := range blocks {
		partStr, ok := b["CPU part"]
		if !ok {
			continue
		}
		part, err := parseHexField(partStr)
		if err != nil {
			return "", fmt.Errorf("bad %q: %q", "CPU part", partStr)
		}
		impl, err := parseHexField(b["CPU implementer"])
		if err != nil {
			// implementer is optional for our purposes; default to 0 (unknown).
			impl = 0
		}
		c := cpuCluster{impl, part}
		if counts[c] == 0 {
			order = append(order, c)
		}
		counts[c]++
		total++
	}
	if total == 0 {
		return "", fmt.Errorf("no CPU part entries found in /proc/cpuinfo")
	}

	// Sort clusters by descending part id so the largest/newest cores (which
	// carry higher part numbers within a generation) are listed first, giving a
	// stable "prime -> big -> little" ordering independent of core enumeration.
	sort.Slice(order, func(i, j int) bool {
		if order[i].part != order[j].part {
			return order[i].part > order[j].part
		}
		return order[i].impl < order[j].impl
	})

	var name string
	if len(order) == 1 {
		name = cpuPartName(order[0].impl, order[0].part)
	} else {
		clusters := make([]string, len(order))
		for i, c := range order {
			clusters[i] = fmt.Sprintf("%dx %s", counts[c], cpuPartName(c.impl, c.part))
		}
		name = strings.Join(clusters, " + ")
	}

	return fmt.Sprintf("%dC%dT %q", total, total, name), nil
}

// parseHexField parses a hexadecimal (or 0x-prefixed) /proc/cpuinfo value.
func parseHexField(s string) (int, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("empty")
	}
	n, err := strconv.ParseInt(s, 0, 64)
	if err != nil {
		return 0, err
	}
	return int(n), nil
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
