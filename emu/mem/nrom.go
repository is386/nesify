package mem

type NROM struct {
	rom [0x4000]uint8
	chr [0x2000]uint8
}

func newNROM() Mapper {
	return &NROM{}
}

func (n *NROM) loadRom(rom []uint8) {
	for i := uint16(0); i < 0x4000; i++ {
		n.rom[i] = rom[i+0x10]
	}
	for i := uint16(0x4010); i < 0x6010; i++ {
		n.chr[i-0x4010] = rom[i]
	}
}

func (n *NROM) read(addr uint16) uint8 {
	if addr >= 0x8000 && addr <= 0xBFFF {
		return n.rom[addr-0x8000]
	} else if addr >= 0xC000 && addr <= 0xFFFF {
		return n.rom[addr-0xC000]
	}
	return 0
}

func (n *NROM) write(addr uint16, val uint8) {
	if addr >= 0x8000 && addr <= 0xBFFF {
		n.rom[addr-0x8000] = val
	} else if addr >= 0xC000 && addr <= 0xFFFF {
		n.rom[addr-0xC000] = val
	}
}

func (n *NROM) getChr() []uint8 {
	return n.chr[:]
}
