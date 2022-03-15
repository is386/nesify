package cpu

import (
	"fmt"
	"os"

	"github.com/is386/NESify/emu/mem"
)

type CPU struct {
	a   uint8
	x   uint8
	y   uint8
	pc  uint8
	s   uint8
	p   uint8
	mmu *mem.MMU
}

func NewCPU(mmu *mem.MMU) *CPU {
	cpu := &CPU{}
	cpu.mmu = mmu
	return cpu
}

func (c *CPU) Update() int {
	opcode := c.fetch()
	instr := c.decode(opcode)
	instr(c)
	return 0
}

func (c *CPU) readMMU(addr uint16) uint8 {
	return c.mmu.Read(addr)
}

func (c *CPU) nextByte() uint8 {
	val := c.readMMU(uint16(c.pc))
	c.pc++
	return val
}

func (c *CPU) fetch() uint8 {
	return c.nextByte()
}

func (c *CPU) decode(opcode uint8) func(*CPU) {
	return OPCODES[opcode]
}

func unimplemented(c *CPU) {
	fmt.Printf("unimplemented: %02X", c.readMMU(uint16(c.pc-1)))
	os.Exit(0)
}

func illegal(c *CPU) {
	fmt.Println("illegal opcode")
	os.Exit(0)
}
