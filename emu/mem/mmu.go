package mem

import "fmt"

type MMU struct {
	ram  [0x800]uint8
	cart Cart
}

func NewMMU(rom []uint8) (*MMU, []uint8) {
	c := newCart(rom)
	return &MMU{cart: c}, c.getChr()
}

func (mmu *MMU) Read(addr uint16) uint8 {
	switch {

	case addr < 0x2000:
		return mmu.ram[addr%0x800]

	case addr < 0x4000 || addr == 0x4014:
		fmt.Printf("reading PPU register at: %X\n", addr)
		return 0

	case addr < 0x4018:
		fmt.Printf("reading APU/IO register at: %X\n", addr)
		return 0

	case addr < 0x4020:
		fmt.Printf("reading disabled APU/IO register at: %X\n", addr)
		return 0

	case addr < 0xFFFF:
		return mmu.cart.read(addr)

	default:
		return 0
	}
}

func (mmu *MMU) Write(addr uint16, val uint8) {
	switch {

	case addr < 0x2000:
		mmu.ram[addr%0x800] = val

	case addr < 0x4000 || addr == 0x4014:
		fmt.Printf("writing PPU register at: %X\n", addr)

	case addr < 0x4018:
		fmt.Printf("writing APU/IO register at: %X\n", addr)

	case addr < 0x4020:
		fmt.Printf("writing disabled APU/IO register at: %X\n", addr)

	case addr < 0xFFFF:
		mmu.cart.write(addr, val)
	}
}
