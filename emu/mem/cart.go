package mem

type Cart struct {
	mapper Mapper
}

func newCart(rom []uint8) Cart {
	c := Cart{mapper: newNROM()}
	c.mapper.loadRom(rom)
	return c
}

func (c *Cart) read(addr uint16) uint8 {
	return c.mapper.read(addr)
}

func (c *Cart) write(addr uint16, val uint8) {
	c.mapper.write(addr, val)
}

func (c *Cart) getChr() []uint8 {
	return c.mapper.getChr()
}
