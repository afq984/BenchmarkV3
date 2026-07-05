package systemdetect

import "fmt"

// cpuPartName decodes an arm64 "CPU implementer" + "CPU part" pair into a human
// readable core name (e.g. 0x41/0xd44 -> "Cortex-X1"). Unknown parts fall back
// to the implementer name plus the raw part id, so detection degrades
// gracefully on newer silicon instead of producing an empty string.
//
// The tables are ported from util-linux's lscpu (lib/cpu-arm.c), the de-facto
// reference for /proc/cpuinfo id decoding.
func cpuPartName(impl, part int) string {
	if parts, ok := cpuParts[impl]; ok {
		if name, ok := parts[part]; ok {
			return name
		}
	}
	return fmt.Sprintf("%s part 0x%03x", implementerName(impl), part)
}

func implementerName(impl int) string {
	if name, ok := cpuImplementers[impl]; ok {
		return name
	}
	return fmt.Sprintf("impl 0x%02x", impl)
}

var cpuImplementers = map[int]string{
	0x41: "ARM",
	0x42: "Broadcom",
	0x43: "Cavium",
	0x44: "DEC",
	0x46: "Fujitsu",
	0x48: "HiSilicon",
	0x49: "Infineon",
	0x4d: "Motorola/Freescale",
	0x4e: "NVIDIA",
	0x50: "APM",
	0x51: "Qualcomm",
	0x53: "Samsung",
	0x56: "Marvell",
	0x61: "Apple",
	0x66: "Faraday",
	0x69: "Intel",
	0x6d: "Microsoft",
	0x70: "Phytium",
	0xc0: "Ampere",
}

// cpuParts maps implementer -> part id -> core name.
var cpuParts = map[int]map[int]string{
	// 0x41 ARM Ltd.
	0x41: {
		0x810: "ARM810",
		0x920: "ARM920",
		0x922: "ARM922",
		0x926: "ARM926",
		0x940: "ARM940",
		0x946: "ARM946",
		0x966: "ARM966",
		0xa20: "ARM1020",
		0xa22: "ARM1022",
		0xa26: "ARM1026",
		0xb02: "ARM11 MPCore",
		0xb36: "ARM1136",
		0xb56: "ARM1156",
		0xb76: "ARM1176",
		0xc05: "Cortex-A5",
		0xc07: "Cortex-A7",
		0xc08: "Cortex-A8",
		0xc09: "Cortex-A9",
		0xc0d: "Cortex-A17", // originally A12
		0xc0f: "Cortex-A15",
		0xc0e: "Cortex-A17",
		0xc14: "Cortex-R4",
		0xc15: "Cortex-R5",
		0xc17: "Cortex-R7",
		0xc18: "Cortex-R8",
		0xc20: "Cortex-M0",
		0xc21: "Cortex-M1",
		0xc23: "Cortex-M3",
		0xc24: "Cortex-M4",
		0xc27: "Cortex-M7",
		0xc60: "Cortex-M0+",
		0xd01: "Cortex-A32",
		0xd02: "Cortex-A34",
		0xd03: "Cortex-A53",
		0xd04: "Cortex-A35",
		0xd05: "Cortex-A55",
		0xd06: "Cortex-A65",
		0xd07: "Cortex-A57",
		0xd08: "Cortex-A72",
		0xd09: "Cortex-A73",
		0xd0a: "Cortex-A75",
		0xd0b: "Cortex-A76",
		0xd0c: "Neoverse-N1",
		0xd0d: "Cortex-A77",
		0xd0e: "Cortex-A76AE",
		0xd13: "Cortex-R52",
		0xd15: "Cortex-R82",
		0xd16: "Cortex-R52+",
		0xd20: "Cortex-M23",
		0xd21: "Cortex-M33",
		0xd40: "Neoverse-V1",
		0xd41: "Cortex-A78",
		0xd42: "Cortex-A78AE",
		0xd43: "Cortex-A65AE",
		0xd44: "Cortex-X1",
		0xd46: "Cortex-A510",
		0xd47: "Cortex-A710",
		0xd48: "Cortex-X2",
		0xd49: "Neoverse-N2",
		0xd4a: "Neoverse-E1",
		0xd4b: "Cortex-A78C",
		0xd4c: "Cortex-X1C",
		0xd4d: "Cortex-A715",
		0xd4e: "Cortex-X3",
		0xd4f: "Neoverse-V2",
		0xd80: "Cortex-A520",
		0xd81: "Cortex-A720",
		0xd82: "Cortex-X4",
		0xd83: "Neoverse-V3AE",
		0xd84: "Neoverse-V3",
		0xd85: "Cortex-X925",
		0xd87: "Cortex-A725",
		0xd88: "Cortex-A520AE",
		0xd89: "Cortex-A720AE",
		0xd8e: "Neoverse-N3",
	},
	// 0x42 Broadcom
	0x42: {
		0x00f: "Brahma-B15",
		0x100: "Brahma-B53",
		0x516: "ThunderX2",
	},
	// 0x43 Cavium
	0x43: {
		0x0a0: "ThunderX",
		0x0a1: "ThunderX-88XX",
		0x0a2: "ThunderX-81XX",
		0x0a3: "ThunderX-83XX",
		0x0af: "ThunderX2-99xx",
		0x0b0: "OcteonTX2",
		0x0b8: "ThunderX3-T110",
	},
	// 0x46 Fujitsu
	0x46: {
		0x001: "A64FX",
	},
	// 0x48 HiSilicon
	0x48: {
		0xd01: "TaiShan-v110", // Kunpeng-920
		0xd02: "TaiShan-v120",
		0xd40: "Cortex-A76", // used in Kirin SoCs
		0xd41: "Cortex-A77",
	},
	// 0x4e NVIDIA
	0x4e: {
		0x000: "Denver",
		0x003: "Denver 2",
		0x004: "Carmel",
	},
	// 0x50 APM (Applied Micro)
	0x50: {
		0x000: "X-Gene",
	},
	// 0x51 Qualcomm
	0x51: {
		0x001: "Oryon",
		0x00f: "Scorpion",
		0x02d: "Scorpion",
		0x04d: "Krait",
		0x06f: "Krait",
		0x201: "Kryo",
		0x205: "Kryo",
		0x211: "Kryo",
		0x800: "Kryo-2XX-Gold", // Falkor V1
		0x801: "Kryo-2XX-Silver",
		0x802: "Kryo-3XX-Gold",
		0x803: "Kryo-3XX-Silver",
		0x804: "Kryo-4XX-Gold",
		0x805: "Kryo-4XX-Silver",
		0xc00: "Falkor",
		0xc01: "Saphira",
	},
	// 0x53 Samsung
	0x53: {
		0x001: "exynos-m1",
		0x002: "exynos-m3",
		0x003: "exynos-m4",
		0x004: "exynos-m5",
	},
	// 0x61 Apple
	0x61: {
		0x022: "Icestorm-M1",
		0x023: "Firestorm-M1",
		0x024: "Icestorm-M1-Pro",
		0x025: "Firestorm-M1-Pro",
		0x028: "Icestorm-M1-Max",
		0x029: "Firestorm-M1-Max",
		0x032: "Blizzard-M2",
		0x033: "Avalanche-M2",
	},
	// 0xc0 Ampere
	0xc0: {
		0xac3: "Ampere-1",
		0xac4: "Ampere-1a",
	},
}
