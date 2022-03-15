package mem

type MMU struct {
	ram [0xFFFF]uint8
}

func NewMMU() *MMU {
	return &MMU{}
}

func (mmu *MMU) Read(addr uint16) uint8 {
	return mmu.ram[addr]
}

func (mmu *MMU) Write(addr uint16, val uint8) {
	mmu.ram[addr] = val
}
