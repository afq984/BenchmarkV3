package main

import (
	"bytes"
	"debug/elf"
	"encoding/binary"
	"testing"
)

// TestLibxml2StubELF parses the generated stub back and checks it is a valid
// versioned ET_DYN exporting exactly the symbols (types and versions) lld needs.
// The stub has no section headers (the loader doesn't use them), so it is read
// the way the dynamic loader reads it: via PT_DYNAMIC. In this stub file offsets
// equal virtual addresses, so DT_* values index the raw bytes directly.
func TestLibxml2StubELF(t *testing.T) {
	for _, machine := range []elf.Machine{elf.EM_X86_64, elf.EM_AARCH64} {
		raw := libxml2StubELF(machine)

		f, err := elf.NewFile(bytes.NewReader(raw))
		if err != nil {
			t.Fatalf("%v: not a valid ELF: %v", machine, err)
		}
		if f.Type != elf.ET_DYN || f.Machine != machine {
			t.Fatalf("%v: Type=%v Machine=%v", machine, f.Type, f.Machine)
		}

		// Read the dynamic table from PT_DYNAMIC.
		dyn := map[elf.DynTag]uint64{}
		for _, p := range f.Progs {
			if p.Type != elf.PT_DYNAMIC {
				continue
			}
			data := make([]byte, p.Filesz)
			if _, err := p.ReadAt(data, 0); err != nil {
				t.Fatal(err)
			}
			for off := 0; off+16 <= len(data); off += 16 {
				var e elf.Dyn64
				binary.Read(bytes.NewReader(data[off:]), binary.LittleEndian, &e)
				if elf.DynTag(e.Tag) == elf.DT_NULL {
					break
				}
				dyn[elf.DynTag(e.Tag)] = e.Val
			}
		}

		// st_name / vda_name / DT_SONAME are offsets into the string table.
		strtab := dyn[elf.DT_STRTAB]
		str := func(rel uint64) string {
			s := raw[strtab+rel:]
			return string(s[:bytes.IndexByte(s, 0)])
		}
		if got := str(dyn[elf.DT_SONAME]); got != verBase {
			t.Errorf("%v: SONAME = %q, want %q", machine, got, verBase)
		}
		if n := dyn[elf.DT_VERDEFNUM]; n != uint64(len(libxml2Versions)) {
			t.Errorf("%v: VERDEFNUM = %d, want %d", machine, n, len(libxml2Versions))
		}

		// version index -> version name, from .gnu.version_d
		verName := map[uint16]string{}
		for i, vd := 0, dyn[elf.DT_VERDEF]; i < len(libxml2Versions); i++ {
			var d elfVerdef
			binary.Read(bytes.NewReader(raw[vd:]), binary.LittleEndian, &d)
			var a elfVerdaux
			binary.Read(bytes.NewReader(raw[vd+uint64(d.Aux):]), binary.LittleEndian, &a)
			verName[d.Ndx] = str(uint64(a.Name))
			vd += uint64(d.Next)
		}

		// walk .dynsym + .gnu.version, collecting name -> (type, version)
		type symInfo struct {
			typ elf.SymType
			ver string
		}
		got := map[string]symInfo{}
		symtab, versym := dyn[elf.DT_SYMTAB], dyn[elf.DT_VERSYM]
		for i := 1; i <= len(libxml2Symbols); i++ {
			var s elf.Sym64
			binary.Read(bytes.NewReader(raw[symtab+uint64(i)*24:]), binary.LittleEndian, &s)
			vi := binary.LittleEndian.Uint16(raw[versym+uint64(i)*2:])
			got[str(uint64(s.Name))] = symInfo{elf.ST_TYPE(s.Info), verName[vi]}
		}

		if len(got) != len(libxml2Symbols) {
			t.Errorf("%v: exported %d distinct symbols, want %d", machine, len(got), len(libxml2Symbols))
		}
		for _, want := range libxml2Symbols {
			wantType := elf.STT_FUNC
			if want.data {
				wantType = elf.STT_OBJECT
			}
			switch g, ok := got[want.name]; {
			case !ok:
				t.Errorf("%v: missing symbol %s", machine, want.name)
			case g.typ != wantType:
				t.Errorf("%v: %s: type %v, want %v", machine, want.name, g.typ, wantType)
			case g.ver != want.version:
				t.Errorf("%v: %s: version %q, want %q", machine, want.name, g.ver, want.version)
			}
		}
	}
}
