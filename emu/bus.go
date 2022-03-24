package emu

type CpuBus struct {
	ram  [0x800]uint8
	cart *Cart
	ppu  *PPU
}

func NewCpuBus(c *Cart, ppu *PPU) *CpuBus {
	b := &CpuBus{cart: c, ppu: ppu}
	return b
}

func (bus *CpuBus) read(addr uint16) uint8 {
	switch {

	case addr < 0x2000:
		return bus.ram[addr%0x800]

	case addr < 0x4000 || addr == 0x4014:
		return bus.ppu.readRegister(addr)

	case addr < 0x4018:
		return 0

	case addr < 0x4020:
		return 0

	case addr < 0xFFFF:
		return bus.cart.read(addr)

	default:
		return 0
	}
}

func (bus *CpuBus) write(addr uint16, val uint8) {
	switch {

	case addr < 0x2000:
		bus.ram[addr%0x800] = val

	case addr < 0x4000 || addr == 0x4014:
		bus.ppu.writeRegister(addr, val)

	case addr < 0x4018:
		return

	case addr < 0x4020:
		return

	case addr < 0xFFFF:
		bus.cart.write(addr, val)
	}
}

type PpuBus struct {
	vram [0x8000]uint8
	oam  [0x0100]uint8
	cart *Cart
}

func NewPpuBus(c *Cart) *PpuBus {
	b := &PpuBus{cart: c}
	return b
}

func (bus *PpuBus) read(addr uint16) uint8 {
	switch {

	case addr < 0x2000:
		return bus.cart.read(addr)

	case addr < 0x4000:
		return bus.vram[addr]

	default:
		return 0
	}
}

func (bus *PpuBus) write(addr uint16, val uint8) {
	switch {

	case addr < 0x2000:
		bus.cart.write(addr, val)

	case addr < 0x4000:
		bus.vram[addr] = val
	}
}

func (bus *PpuBus) readOam(addr uint8) uint8 {
	return bus.oam[addr]
}

func (bus *PpuBus) writeOam(addr uint8, val uint8) {
	bus.oam[addr] = val
}
