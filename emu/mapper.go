package emu

type Mapper interface {
	loadRom(rom []uint8)
	read(addr uint16) uint8
	write(addr uint16, val uint8)
}
