package emu

type NROM struct {
	rom  [0x8000]uint8
	chr  [0x4000]uint8
	size int
}

func newNROM() Mapper {
	return &NROM{}
}

func (n *NROM) loadRom(rom []uint8) {
	n.size = len(rom)
	end := uint16(len(rom) - 0x2010)
	for i := uint16(0); i < end; i++ {
		n.rom[i] = rom[i+0x10]
	}
	start := uint16(len(rom) - 0x2000)
	end = uint16(len(rom))
	for i := start; i < end; i++ {
		n.chr[i-start] = rom[i]
	}
}

func (n *NROM) read(addr uint16) uint8 {
	if addr < 0x2000 {
		return n.chr[addr]
	} else if addr >= 0x8000 && addr <= 0xBFFF {
		return n.rom[addr-0x8000]
	} else if addr >= 0xC000 && addr <= 0xFFFF {
		if n.size == 0xA010 {
			return n.rom[addr-0x8000]
		} else {
			return n.rom[addr-0xC000]
		}
	}
	return 0
}

func (n *NROM) write(addr uint16, val uint8) {
	if addr < 0x2000 {
		n.chr[addr] = val
	} else if addr >= 0x8000 && addr <= 0xBFFF {
		n.rom[addr-0x8000] = val
	} else if addr >= 0xC000 && addr <= 0xFFFF {
		if n.size == 0xA010 {
			n.rom[addr-0x8000] = val
		} else {
			n.rom[addr-0xC000] = val
		}
	}
}
