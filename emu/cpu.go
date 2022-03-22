package emu

import (
	"fmt"

	"github.com/is386/NESify/emu/bits"
)

type Interrupt int

const (
	Nmi Interrupt = iota
	NoInterrupt
)

var instructionNames = [256]string{
	"BRK", "ORA", "KIL", "SLO", "NOP", "ORA", "ASL", "SLO",
	"PHP", "ORA", "ASL", "ANC", "NOP", "ORA", "ASL", "SLO",
	"BPL", "ORA", "KIL", "SLO", "NOP", "ORA", "ASL", "SLO",
	"CLC", "ORA", "NOP", "SLO", "NOP", "ORA", "ASL", "SLO",
	"JSR", "AND", "KIL", "RLA", "BIT", "AND", "ROL", "RLA",
	"PLP", "AND", "ROL", "ANC", "BIT", "AND", "ROL", "RLA",
	"BMI", "AND", "KIL", "RLA", "NOP", "AND", "ROL", "RLA",
	"SEC", "AND", "NOP", "RLA", "NOP", "AND", "ROL", "RLA",
	"RTI", "EOR", "KIL", "SRE", "NOP", "EOR", "LSR", "SRE",
	"PHA", "EOR", "LSR", "ALR", "JMP", "EOR", "LSR", "SRE",
	"BVC", "EOR", "KIL", "SRE", "NOP", "EOR", "LSR", "SRE",
	"CLI", "EOR", "NOP", "SRE", "NOP", "EOR", "LSR", "SRE",
	"RTS", "ADC", "KIL", "RRA", "NOP", "ADC", "ROR", "RRA",
	"PLA", "ADC", "ROR", "ARR", "JMP", "ADC", "ROR", "RRA",
	"BVS", "ADC", "KIL", "RRA", "NOP", "ADC", "ROR", "RRA",
	"SEI", "ADC", "NOP", "RRA", "NOP", "ADC", "ROR", "RRA",
	"NOP", "STA", "NOP", "SAX", "STY", "STA", "STX", "SAX",
	"DEY", "NOP", "TXA", "XAA", "STY", "STA", "STX", "SAX",
	"BCC", "STA", "KIL", "AHX", "STY", "STA", "STX", "SAX",
	"TYA", "STA", "TXS", "TAS", "SHY", "STA", "SHX", "AHX",
	"LDY", "LDA", "LDX", "LAX", "LDY", "LDA", "LDX", "LAX",
	"TAY", "LDA", "TAX", "LAX", "LDY", "LDA", "LDX", "LAX",
	"BCS", "LDA", "KIL", "LAX", "LDY", "LDA", "LDX", "LAX",
	"CLV", "LDA", "TSX", "LAS", "LDY", "LDA", "LDX", "LAX",
	"CPY", "CMP", "NOP", "DCP", "CPY", "CMP", "DEC", "DCP",
	"INY", "CMP", "DEX", "AXS", "CPY", "CMP", "DEC", "DCP",
	"BNE", "CMP", "KIL", "DCP", "NOP", "CMP", "DEC", "DCP",
	"CLD", "CMP", "NOP", "DCP", "NOP", "CMP", "DEC", "DCP",
	"CPX", "SBC", "NOP", "ISC", "CPX", "SBC", "INC", "ISC",
	"INX", "SBC", "NOP", "SBC", "CPX", "SBC", "INC", "ISC",
	"BEQ", "SBC", "KIL", "ISC", "NOP", "SBC", "INC", "ISC",
	"SED", "SBC", "NOP", "ISC", "NOP", "SBC", "INC", "ISC",
}

type CPU struct {
	cyc        int
	a, x, y, s uint8
	pc         uint16
	p          *Status
	instr      Instruction
	bus        *CpuBus
	interrupt  Interrupt
	debug      bool
}

func NewCPU(bus *CpuBus, debug bool) *CPU {
	return &CPU{
		pc:        (uint16(bus.read(0xFFFC+1)) << 8) | uint16(bus.read(0xFFFC)),
		s:         0xFD,
		p:         NewStatus(),
		bus:       bus,
		debug:     debug,
		interrupt: NoInterrupt,
	}
}

func (c *CPU) update() int {
	c.print()
	c.cyc = 0
	c.checkInterrupts()
	opcode := c.fetch()
	c.instr = c.decode(opcode)
	operand := c.getOperand(c.instr.addrMode)
	c.instr.function(c, operand)
	c.cyc += c.instr.cyc
	return c.cyc
}

func (c *CPU) print() {
	if c.debug {
		fmt.Printf("%04X %s A:%02X X:%02X Y:%02X P:%02X SP:%02X\n",
			c.pc, instructionNames[c.read(c.pc)], c.a, c.x, c.y, c.p.getStatus(), c.s)
	}
}

func (c *CPU) read(addr uint16) uint8 {
	return c.bus.read(addr)
}

func (c *CPU) write(addr uint16, val uint8) {
	c.bus.write(addr, val)
}

func (c *CPU) nextByte() uint8 {
	val := c.read(c.pc)
	c.pc++
	return val
}

func (c *CPU) nextTwoBytes() uint16 {
	a := c.nextByte()
	b := c.nextByte()
	return (uint16(b) << 8) | uint16(a)
}

func (c *CPU) readTwoBytes() uint16 {
	a := c.nextByte()
	b := c.nextByte()
	addr := (uint16(b) << 8) | uint16(a)
	a = c.read(addr)
	b = c.read(addr + 1)
	if (addr & 0xFF) == 0xFF {
		b = c.read(addr & 0xF00)
	}
	addr = (uint16(b) << 8) | uint16(a)
	return addr
}

func (c *CPU) readTwoBytesIndexed(addr uint16) uint16 {
	lo := c.read((addr + 1) & 0xFF)
	hi := c.read(addr & 0xFF)
	return uint16(hi) | uint16(lo)<<8
}

func (c *CPU) checkPage(a, b uint16) {
	if (a & 0xFF00) != (b & 0xFF00) {
		c.cyc += c.instr.pageCyc
	}
}

func (c *CPU) getOperand(addrMode AddrMode) uint16 {
	switch addrMode {

	case Acc:
		return 0

	case Abs:
		return c.nextTwoBytes()

	case Abx:
		addr := c.nextTwoBytes() + uint16(c.x)
		c.checkPage(addr-uint16(c.x), addr)
		return addr

	case Aby:
		addr := c.nextTwoBytes() + uint16(c.y)
		c.checkPage(addr-uint16(c.y), addr)
		return addr

	case Imm:
		addr := c.pc
		c.pc++
		return uint16(addr)

	case Imp:
		return 0

	case Ind:
		return c.readTwoBytes()

	case Inx:
		return c.readTwoBytesIndexed(uint16(c.nextByte()) + uint16(c.x))

	case Iny:
		addr := c.readTwoBytesIndexed(uint16(c.nextByte())) + uint16(c.y)
		c.checkPage(addr-uint16(c.y), addr)
		return addr

	case Rel:
		offset := uint16(c.nextByte())
		var addr uint16
		if offset < 0x80 {
			addr = c.pc + offset
		} else {
			addr = c.pc + offset - 0x100
		}
		return addr

	case Zp:
		return uint16(c.nextByte()) % 256

	case Zpx:
		return (uint16(c.nextByte()) + uint16(c.x)) % 256

	case Zpy:
		return (uint16(c.nextByte()) + uint16(c.y)) % 256

	default:
		return 0
	}
}

func (c *CPU) fetch() uint8 {
	return c.nextByte()
}

func (c *CPU) decode(opcode uint8) Instruction {
	return INSTRUCTIONS[opcode]
}

func (c *CPU) push8(val uint8) {
	c.write(0x100|uint16(c.s), val)
	c.s--
}

func (c *CPU) push16(val uint16) {
	hi := uint8(val >> 8)
	lo := uint8(val & 0xFF)
	c.push8(hi)
	c.push8(lo)
}

func (c *CPU) pop8() uint8 {
	c.s++
	return c.read(0x100 | uint16(c.s))
}

func (c *CPU) pop16() uint16 {
	lo := uint16(c.pop8())
	hi := uint16(c.pop8())
	return (hi << 8) | lo
}

func (c *CPU) triggerInterrupt(i Interrupt) {
	c.interrupt = i
}

func (c *CPU) checkInterrupts() {
	switch c.interrupt {
	case Nmi:
		c.push16(c.pc)
		php(c, 0)
		c.pc = (uint16(c.read(0xFFFB)) << 8) | uint16(c.read(0xFFFA))
		c.p.setInterrupt()
		c.cyc += 7
	}
	c.interrupt = NoInterrupt
}

func illegal(c *CPU, operand uint16) {
	if c.debug {
		for {
		}
	}
}

func adc(c *CPU, operand uint16) {
	cy := c.p.getCarry()
	val := uint16(c.read(operand))
	ans := uint16(c.a) + val + uint16(cy)
	c.p.checkNegative(uint8(ans))
	c.p.checkZero(uint8(ans))
	c.p.checkCarry(ans, 0x100)
	c.p.checkOverflow(uint16(c.a), val, ans)
	c.a = uint8(ans)
}

func and(c *CPU, operand uint16) {
	c.a &= c.read(operand)
	c.p.checkNegative(c.a)
	c.p.checkZero(c.a)
}

func asla(c *CPU, operand uint16) {
	if ((c.a >> 7) & 1) == 1 {
		c.p.setCarry()
	} else {
		c.p.resetCarry()
	}
	c.a <<= 1
	c.p.checkNegative(c.a)
	c.p.checkZero(c.a)
}

func asl(c *CPU, operand uint16) {
	val := c.read(operand)
	if ((val >> 7) & 1) == 1 {
		c.p.setCarry()
	} else {
		c.p.resetCarry()
	}
	val <<= 1
	c.write(operand, val)
	c.p.checkNegative(val)
	c.p.checkZero(val)
}

func bcc(c *CPU, operand uint16) {
	if c.p.getCarry() == 0 {
		c.pc = operand
		c.cyc++
	}
}

func bcs(c *CPU, operand uint16) {
	if c.p.getCarry() == 1 {
		c.pc = operand
		c.cyc++
	}
}

func beq(c *CPU, operand uint16) {
	if c.p.getZero() == 1 {
		c.pc = operand
		c.cyc++
	}
}

func bit(c *CPU, operand uint16) {
	val := c.read(operand)
	if bits.Test(val, 7) {
		c.p.setNegative()
	} else {
		c.p.resetNegative()
	}

	if bits.Test(val, 6) {
		c.p.setOverflow()
	} else {
		c.p.resetOverflow()
	}

	c.p.checkZero(c.a & val)
}

func bne(c *CPU, operand uint16) {
	if c.p.getZero() == 0 {
		c.pc = operand
		c.cyc++
	}
}

func bmi(c *CPU, operand uint16) {
	if c.p.getNegative() == 1 {
		c.pc = operand
		c.cyc++
	}
}

func bpl(c *CPU, operand uint16) {
	if c.p.getNegative() == 0 {
		c.pc = operand
		c.cyc++
	}
}

func brk(c *CPU, operand uint16) {
	c.push16(c.pc + 1)
	c.p.setInterrupt()
	c.push8(c.p.getStatus())
}

func bvc(c *CPU, operand uint16) {
	if c.p.getOverflow() == 0 {
		c.pc = operand
		c.cyc++
	}
}

func bvs(c *CPU, operand uint16) {
	if c.p.getOverflow() == 1 {
		c.pc = operand
		c.cyc++
	}
}

func clc(c *CPU, operand uint16) {
	c.p.resetCarry()
}

func cld(c *CPU, operand uint16) {
	c.p.resetDecimal()
}

func cli(c *CPU, operand uint16) {
	c.p.resetInterrupt()
}

func clv(c *CPU, operand uint16) {
	c.p.resetOverflow()
}

func cmp(c *CPU, operand uint16) {
	val := c.a - c.read(operand)
	c.p.checkNegative(val)
	c.p.checkZero(val)
	c.p.checkCarry(uint16(c.a), uint16(c.read(operand)))
}

func cpx(c *CPU, operand uint16) {
	val := c.x - c.read(operand)
	c.p.checkNegative(val)
	c.p.checkZero(val)
	c.p.checkCarry(uint16(c.x), uint16(c.read(operand)))
}

func cpy(c *CPU, operand uint16) {
	val := c.y - c.read(operand)
	c.p.checkNegative(val)
	c.p.checkZero(val)
	c.p.checkCarry(uint16(c.y), uint16(c.read(operand)))
}

func dec(c *CPU, operand uint16) {
	val := c.read(operand) - 1
	c.write(operand, val)
	c.p.checkNegative(val)
	c.p.checkZero(val)
}

func dex(c *CPU, operand uint16) {
	c.x--
	c.p.checkNegative(c.x)
	c.p.checkZero(c.x)
}

func dey(c *CPU, operand uint16) {
	c.y--
	c.p.checkNegative(c.y)
	c.p.checkZero(c.y)
}

func eor(c *CPU, operand uint16) {
	c.a ^= c.read(operand)
	c.p.checkNegative(c.a)
	c.p.checkZero(c.a)
}

func inc(c *CPU, operand uint16) {
	val := c.read(operand) + 1
	c.write(operand, val)
	c.p.checkNegative(val)
	c.p.checkZero(val)
}

func inx(c *CPU, operand uint16) {
	c.x++
	c.p.checkNegative(c.x)
	c.p.checkZero(c.x)
}

func iny(c *CPU, operand uint16) {
	c.y++
	c.p.checkNegative(c.y)
	c.p.checkZero(c.y)
}

func jmp(c *CPU, operand uint16) {
	c.pc = operand
}

func jsr(c *CPU, operand uint16) {
	c.push16(c.pc - 1)
	c.pc = operand
}

func lda(c *CPU, operand uint16) {
	c.a = c.read(operand)
	c.p.checkNegative(c.a)
	c.p.checkZero(c.a)
}

func ldx(c *CPU, operand uint16) {
	c.x = c.read(operand)
	c.p.checkNegative(c.x)
	c.p.checkZero(c.x)
}

func ldy(c *CPU, operand uint16) {
	c.y = c.read(operand)
	c.p.checkNegative(c.y)
	c.p.checkZero(c.y)
}

func lsra(c *CPU, operand uint16) {
	if (c.a & 1) == 1 {
		c.p.setCarry()
	} else {
		c.p.resetCarry()
	}
	c.a >>= 1
	c.p.resetNegative()
	c.p.checkZero(c.a)
}

func lsr(c *CPU, operand uint16) {
	val := c.read(operand)
	if (val & 1) == 1 {
		c.p.setCarry()
	} else {
		c.p.resetCarry()
	}
	val >>= 1
	c.write(operand, val)
	c.p.resetNegative()
	c.p.checkZero(val)
}

func nop(c *CPU, operand uint16) {}

func ora(c *CPU, operand uint16) {
	c.a |= c.read(operand)
	c.p.checkNegative(c.a)
	c.p.checkZero(c.a)
}

func pha(c *CPU, operand uint16) {
	c.push8(c.a)
}

func php(c *CPU, operand uint16) {
	c.push8(c.p.getStatus() | 0x10)
}

func pla(c *CPU, operand uint16) {
	c.a = c.pop8()
	c.p.checkNegative(c.a)
	c.p.checkZero(c.a)
}

func plp(c *CPU, operand uint16) {
	bit4 := bits.Value(c.p.getStatus(), 4)
	bit5 := bits.Value(c.p.getStatus(), 5)
	c.p.setStatus(c.pop8())

	if bit4 == 1 {
		c.p.setBreak()
	} else {
		c.p.resetBreak()
	}

	if bit5 == 1 {
		c.p.setBit5()
	} else {
		c.p.resetBit5()
	}
}

func rola(c *CPU, operand uint16) {
	cy := c.p.getCarry()
	if ((c.a >> 7) & 1) == 1 {
		c.p.setCarry()
	} else {
		c.p.resetCarry()
	}
	c.a = (c.a << 1) | cy
	c.p.checkNegative(c.a)
	c.p.checkZero(c.a)
}

func rol(c *CPU, operand uint16) {
	val := c.read(operand)
	cy := c.p.getCarry()
	if ((val >> 7) & 1) == 1 {
		c.p.setCarry()
	} else {
		c.p.resetCarry()
	}
	val = (val << 1) | cy
	c.write(operand, val)
	c.p.checkNegative(val)
	c.p.checkZero(val)
}

func rora(c *CPU, operand uint16) {
	cy := c.p.getCarry()
	if (c.a & 1) == 1 {
		c.p.setCarry()
	} else {
		c.p.resetCarry()
	}
	c.a = (c.a >> 1) | (cy << 7)
	c.p.checkNegative(c.a)
	c.p.checkZero(c.a)
}

func ror(c *CPU, operand uint16) {
	val := c.read(operand)
	cy := c.p.getCarry()
	if (val & 1) == 1 {
		c.p.setCarry()
	} else {
		c.p.resetCarry()
	}
	val = (val >> 1) | (cy << 7)
	c.write(operand, val)
	c.p.checkNegative(val)
	c.p.checkZero(val)
}

func rti(c *CPU, operand uint16) {
	c.p.setStatus(c.pop8()&0xEF | 0x20)
	c.pc = c.pop16()
}

func rts(c *CPU, operand uint16) {
	c.pc = c.pop16() + 1
}

func sbc(c *CPU, operand uint16) {
	cy := c.p.getCarry()
	val := uint16(c.read(operand))
	ans := uint16(c.a) - val - uint16(1-cy)
	c.p.checkNegative(uint8(ans))
	c.p.checkZero(uint8(ans))
	c.p.checkBorrow(int(c.a)-int(val)-int(1-cy), 0)
	c.p.checkUnderflow(uint16(c.a), val, ans)
	c.a = uint8(ans)
}

func sec(c *CPU, operand uint16) {
	c.p.setCarry()
}

func sed(c *CPU, operand uint16) {
	c.p.setDecimal()
}

func sei(c *CPU, operand uint16) {
	c.p.setInterrupt()
}

func sta(c *CPU, operand uint16) {
	c.write(operand, c.a)
}

func stx(c *CPU, operand uint16) {
	c.write(operand, c.x)
}

func sty(c *CPU, operand uint16) {
	c.write(operand, c.y)
}

func tax(c *CPU, operand uint16) {
	c.x = c.a
	c.p.checkNegative(c.x)
	c.p.checkZero(c.x)
}

func tsx(c *CPU, operand uint16) {
	c.x = c.s
	c.p.checkNegative(c.x)
	c.p.checkZero(c.x)
}

func txa(c *CPU, operand uint16) {
	c.a = c.x
	c.p.checkNegative(c.a)
	c.p.checkZero(c.a)
}

func txs(c *CPU, operand uint16) {
	c.s = c.x
}

func tay(c *CPU, operand uint16) {
	c.y = c.a
	c.p.checkNegative(c.y)
	c.p.checkZero(c.y)
}

func tya(c *CPU, operand uint16) {
	c.a = c.y
	c.p.checkNegative(c.a)
	c.p.checkZero(c.a)
}
