package main

import (
	"debug/elf"
	"encoding/binary"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// The LLVM lld linker lists libxml2.so.2 as a NEEDED library and imports the
// symbols below, but it only calls them for COFF/Windows output. Our ELF
// cross-link never invokes them; lld is linked BIND_NOW, so the dynamic loader
// resolves their addresses at startup but never calls through.
//
// On hosts that lack libxml2.so.2 (minimal installs, or rolling-release distros
// that moved libxml2 to a newer soname such as libxml2.so.16) lld fails to
// start even though it never needs the library. Rather than ship or require a
// prebuilt binary, we generate a tiny stub libxml2.so.2 whose symbols point at
// a zeroed region. It only has to satisfy the loader, never run.
//
// lld imports these symbols with their libxml2 symbol versions, and glibc
// refuses to bind a versioned reference against an unversioned definition when
// the soname matches, so the stub reproduces those versions. If a future
// toolchain references different symbols/versions, update this list from
// `nm -D --undefined-only lld`.
const (
	verBase = "libxml2.so.2" // base version node; its name is the soname
	ver2430 = "LIBXML2_2.4.30"
	ver260  = "LIBXML2_2.6.0"
)

var libxml2Symbols = []struct {
	name    string
	data    bool // xmlFree is a data object (a function pointer); the rest are functions
	version string
}{
	{"xmlAddChild", false, ver2430},
	{"xmlCopyNamespace", false, ver2430},
	{"xmlDocDumpFormatMemoryEnc", false, ver2430},
	{"xmlDocGetRootElement", false, ver2430},
	{"xmlDocSetRootElement", false, ver2430},
	{"xmlFreeDoc", false, ver2430},
	{"xmlFreeNode", false, ver2430},
	{"xmlFreeNs", false, ver2430},
	{"xmlNewDoc", false, ver2430},
	{"xmlNewNs", false, ver2430},
	{"xmlNewProp", false, ver2430},
	{"xmlReadMemory", false, ver260},
	{"xmlSetGenericErrorFunc", false, ver2430},
	{"xmlStrdup", false, ver2430},
	{"xmlUnlinkNode", false, ver2430},
	{"xmlFree", true, ver2430},
}

// version definitions the stub provides: index 1 is the base (soname), the rest
// are the versions lld requires. Order matters: vd_ndx is the 1-based position.
var libxml2Versions = []struct {
	name string
	base bool
}{
	{verBase, true},
	{ver2430, false},
	{ver260, false},
}

// Elf64 symbol-version structures, which debug/elf does not export.
type elfVerdef struct {
	Version, Flags, Ndx, Cnt uint16
	Hash, Aux, Next          uint32
}

type elfVerdaux struct {
	Name, Next uint32
}

const (
	elfVerdefSize  = 20
	elfVerdauxSize = 8
)

// elfHash is the classic SysV/ELF symbol-name hash, also used for version names.
func elfHash(s string) uint32 {
	var h uint32
	for i := 0; i < len(s); i++ {
		h = (h << 4) + uint32(s[i])
		if g := h & 0xf0000000; g != 0 {
			h ^= g >> 24
		}
		h &^= 0xf0000000
	}
	return h
}

// libxml2StubELF returns the bytes of a minimal, versioned ELF64 shared object
// with soname libxml2.so.2 for the given machine, exporting libxml2Symbols with
// the versions lld requires. It is assembled from debug/elf's structures so it
// reads as an ELF description: an ET_DYN with one PT_LOAD covering the whole
// file, a PT_DYNAMIC, a SysV hash table, .dynsym, .dynstr, .gnu.version and
// .gnu.version_d. There are no section headers, no code and no relocations;
// every symbol resolves to one zeroed 8-byte slot, which is all the loader needs
// (the symbols are never called).
func libxml2StubELF(machine elf.Machine) []byte {
	le := binary.LittleEndian
	symEnt := uint64(binary.Size(elf.Sym64{})) // 24
	dynEnt := uint64(binary.Size(elf.Dyn64{})) // 16
	nsym := len(libxml2Symbols) + 1            // + the null symbol at index 0

	vidxOf := func(name string) uint16 {
		for i, v := range libxml2Versions {
			if v.name == name {
				return uint16(i + 1) // vd_ndx is 1-based
			}
		}
		return 1
	}

	// .dynstr: leading NUL, symbol names, soname, then version names.
	dynstr := []byte{0}
	addStr := func(s string) uint32 {
		off := uint32(len(dynstr))
		dynstr = append(dynstr, s...)
		dynstr = append(dynstr, 0)
		return off
	}
	nameOff := make([]uint32, len(libxml2Symbols))
	for i, s := range libxml2Symbols {
		nameOff[i] = addStr(s.name)
	}
	sonameOff := addStr(verBase)
	verNameOff := make([]uint32, len(libxml2Versions))
	verNameOff[0] = sonameOff // the base version's name is the soname
	for i := 1; i < len(libxml2Versions); i++ {
		verNameOff[i] = addStr(libxml2Versions[i].name)
	}

	// Lay the pieces out end to end and record their offsets.
	const ehdrSize, phdrSize, phnum = 64, 56, 2
	align := func(x, a uint64) uint64 { return (x + a - 1) &^ (a - 1) }

	hashOff := uint64(ehdrSize + phnum*phdrSize)
	hashSize := uint64(4 * (3 + nsym)) // nbucket, nchain, one bucket, nsym chain slots
	dynsymOff := hashOff + hashSize
	dynstrOff := dynsymOff + uint64(nsym)*symEnt
	dynstrSize := uint64(len(dynstr))
	versymOff := align(dynstrOff+dynstrSize, 2)
	verdefOff := align(versymOff+uint64(nsym)*2, 4)
	verdefSize := uint64(len(libxml2Versions) * (elfVerdefSize + elfVerdauxSize))
	dynOff := align(verdefOff+verdefSize, 8)
	const numDyn = 10
	symvalOff := dynOff + numDyn*dynEnt
	total := symvalOff + 8 // the shared zeroed slot every symbol points at

	b := make([]byte, total)
	put := func(off uint64, v any) {
		if _, err := binary.Encode(b[off:], le, v); err != nil {
			panic(err)
		}
	}

	// ELF header.
	var ident [elf.EI_NIDENT]byte
	copy(ident[:], elf.ELFMAG)
	ident[elf.EI_CLASS] = byte(elf.ELFCLASS64)
	ident[elf.EI_DATA] = byte(elf.ELFDATA2LSB)
	ident[elf.EI_VERSION] = byte(elf.EV_CURRENT)
	put(0, elf.Header64{
		Ident:     ident,
		Type:      uint16(elf.ET_DYN),
		Machine:   uint16(machine),
		Version:   uint32(elf.EV_CURRENT),
		Phoff:     ehdrSize,
		Ehsize:    ehdrSize,
		Phentsize: phdrSize,
		Phnum:     phnum,
	})

	// Program headers: one PT_LOAD covering the whole file, and PT_DYNAMIC.
	rw := uint32(elf.PF_R | elf.PF_W)
	put(ehdrSize, elf.Prog64{
		Type:   uint32(elf.PT_LOAD),
		Flags:  rw,
		Filesz: total,
		Memsz:  total,
		Align:  0x1000,
	})
	put(ehdrSize+phdrSize, elf.Prog64{
		Type:   uint32(elf.PT_DYNAMIC),
		Flags:  rw,
		Off:    dynOff,
		Vaddr:  dynOff,
		Paddr:  dynOff,
		Filesz: numDyn * dynEnt,
		Memsz:  numDyn * dynEnt,
		Align:  8,
	})

	// SysV hash table: a single bucket, so every symbol sits on one chain.
	hash := []uint32{1, uint32(nsym), 1} // nbucket, nchain, bucket[0]
	for i := range nsym {
		next := uint32(i + 1) // chain[i] -> next symbol
		if i == 0 || i == nsym-1 {
			next = 0 // chain[0] is unused; the last entry ends the chain
		}
		hash = append(hash, next)
	}
	put(hashOff, hash)

	// .dynsym (index 0 is the reserved null symbol).
	syms := make([]elf.Sym64, nsym)
	for i, s := range libxml2Symbols {
		typ := elf.STT_FUNC
		if s.data {
			typ = elf.STT_OBJECT
		}
		sym := elf.Sym64{
			Name:  nameOff[i],
			Info:  elf.ST_INFO(elf.STB_GLOBAL, typ),
			Shndx: 1,         // any defined section -> address = load base + Value
			Value: symvalOff, // the shared zeroed slot; never dereferenced
		}
		if s.data {
			sym.Size = 8
		}
		syms[i+1] = sym
	}
	put(dynsymOff, syms)

	copy(b[dynstrOff:], dynstr)

	// .gnu.version: a version index per .dynsym entry (0 for the null symbol).
	versym := make([]uint16, nsym)
	for i, s := range libxml2Symbols {
		versym[i+1] = vidxOf(s.version)
	}
	put(versymOff, versym)

	// .gnu.version_d: one Verdef (plus its single Verdaux) per version.
	for i, v := range libxml2Versions {
		vd := verdefOff + uint64(i*(elfVerdefSize+elfVerdauxSize))
		var flags uint16
		if v.base {
			flags = uint16(elf.VER_FLG_BASE)
		}
		next := uint32(elfVerdefSize + elfVerdauxSize)
		if i == len(libxml2Versions)-1 {
			next = 0 // the last definition ends the list
		}
		put(vd, elfVerdef{
			Version: 1,
			Flags:   flags,
			Ndx:     uint16(i + 1),
			Cnt:     1,
			Hash:    elfHash(v.name),
			Aux:     elfVerdefSize, // the Verdaux immediately follows
			Next:    next,
		})
		put(vd+elfVerdefSize, elfVerdaux{Name: verNameOff[i]})
	}

	// .dynamic
	put(dynOff, []elf.Dyn64{
		{Tag: int64(elf.DT_HASH), Val: hashOff},
		{Tag: int64(elf.DT_STRTAB), Val: dynstrOff},
		{Tag: int64(elf.DT_SYMTAB), Val: dynsymOff},
		{Tag: int64(elf.DT_STRSZ), Val: dynstrSize},
		{Tag: int64(elf.DT_SYMENT), Val: symEnt},
		{Tag: int64(elf.DT_SONAME), Val: uint64(sonameOff)},
		{Tag: int64(elf.DT_VERSYM), Val: versymOff},
		{Tag: int64(elf.DT_VERDEF), Val: verdefOff},
		{Tag: int64(elf.DT_VERDEFNUM), Val: uint64(len(libxml2Versions))},
		{Tag: int64(elf.DT_NULL), Val: 0},
	})

	return b
}

// hostEMachine returns the ELF machine for the current host, or false if this
// host is not a Linux architecture the stub supports.
func hostEMachine() (elf.Machine, bool) {
	if runtime.GOOS != "linux" {
		return 0, false
	}
	switch runtime.GOARCH {
	case "amd64":
		return elf.EM_X86_64, true
	case "arm64":
		return elf.EM_AARCH64, true
	default:
		return 0, false
	}
}

// lldRuns reports whether the linker at lldPath can start, i.e. whether its
// libxml2.so.2 dependency is satisfied by the host. It is only a probe, so its
// output is discarded.
func lldRuns(lldPath string) bool {
	cmd := exec.Command(lldPath, "-flavor", "gnu", "--version")
	return cmd.Run() == nil
}

// libxml2StubEnv returns environment additions (an LD_LIBRARY_PATH entry) that
// let the toolchain's lld run on this host, or nil if none are needed.
//
// The bundled lld lists libxml2.so.2 as NEEDED but never calls it for ELF
// linking. On hosts that already provide libxml2.so.2 we must NOT interpose our
// stub, so we probe lldPath first and only write a stub when lld cannot
// otherwise start (minimal installs, or distros that renamed libxml2 to a newer
// soname such as libxml2.so.16).
func libxml2StubEnv(buildDir, lldPath string) ([]string, error) {
	machine, ok := hostEMachine()
	if !ok || lldRuns(lldPath) {
		return nil, nil
	}

	// The directory must be absolute: it goes into LD_LIBRARY_PATH, which the
	// dynamic loader resolves relative to each linker process's working
	// directory (cmake runs the compiler checks in a temporary subdirectory).
	dir, err := filepath.Abs(filepath.Join(buildDir, "xmlstub"))
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	if err := os.WriteFile(filepath.Join(dir, "libxml2.so.2"), libxml2StubELF(machine), 0o644); err != nil {
		return nil, err
	}

	ldPath := dir
	if existing := os.Getenv("LD_LIBRARY_PATH"); existing != "" {
		ldPath += string(os.PathListSeparator) + existing
	}
	return []string{"LD_LIBRARY_PATH=" + ldPath}, nil
}
