package emu

import (
	"github.com/is386/NESify/emu/bits"
)

const (
	NES_WIDTH  = 256
	NES_HEIGHT = 240
	CHR_WIDTH  = 128
	CHR_HEIGHT = 256
	SCALE      = 3
	PPUCTRL    = 0x2000
	PPUMASK    = 0x2001
	PPUSTATUS  = 0x2002
	OAMADDR    = 0x2003
	OAMDATA    = 0x2004
	PPUSCROLL  = 0x2005
	PPUADDR    = 0x2006
	PPUDATA    = 0x2007
	OAMDMA     = 0x4014
)

var (
	COLORS = []uint32{0x000000, 0x545454, 0xA9A9A9, 0xEFEFEF}
)

type PPU struct {
	cpu                                   *CPU
	bus                                   *PpuBus
	screen                                *Screen
	reg                                   [9]uint8
	scanline, cyc                         int
	addr                                  uint16
	nmiOccurred, nmiOutput, ppuAddrLoaded bool
}

func NewPPU(b *PpuBus) *PPU {
	p := &PPU{bus: b}
	p.screen = NewScreen(NES_WIDTH+CHR_WIDTH, CHR_HEIGHT, SCALE)
	p.screen.win.SetTitle("NESify")
	p.showCHR()
	return p
}

func (p *PPU) update() {
	if p.nmiOutput && p.nmiOccurred {
		p.cpu.interrupt = Nmi
	}

	p.cyc++
	if p.cyc > 340 {
		p.cyc -= 341
		p.scanline++
	}

	if p.scanline >= 0 && p.scanline <= 239 {
		// drawing
	} else if p.scanline == 241 && p.cyc == 1 {
		stat := p.readRegister(PPUSTATUS)
		stat = bits.Set(stat, 7)
		p.writeRegister(PPUSTATUS, stat)
		p.nmiOccurred = true
		p.renderNameTables()
	} else if p.scanline == 261 && p.cyc == 1 {
		stat := p.readRegister(PPUSTATUS)
		stat = bits.Reset(stat, 7)
		p.writeRegister(PPUSTATUS, stat)
		p.nmiOccurred = false
		p.scanline = 0
	}
}

func (p *PPU) renderNameTables() {
	ntAddr := uint16(0x2000)
	baseX := 0
	for ntByte := uint16(0); ntByte < 960; ntByte++ {
		ptIdx := p.bus.read(ntAddr)
		ptBaseAddr := 0x1000 + (uint16(ptIdx) * 16)

		y := 0
		for ptAddr := ptBaseAddr; ptAddr < ptBaseAddr+8; ptAddr++ {
			ptByte1 := p.bus.read(ptAddr)
			ptByte2 := p.bus.read(ptAddr + 8)

			for x := baseX; x < baseX+8; x++ {
				pixel := 7 - (x % 8)
				colorBit0 := (ptByte1 >> pixel) & 1
				colorBit1 := (ptByte2 >> pixel) & 1
				colorNum := (colorBit1 << 1) | colorBit0
				p.screen.drawPixel(int32(x), int32(y+int(ntByte/32)*8), COLORS[colorNum])
			}
			y++
		}

		baseX = (baseX + 8) % NES_WIDTH
		ntAddr++
	}
	p.screen.Update()
}

func (p *PPU) showCHR() {
	for y := 0; y < CHR_HEIGHT; y++ {
		for x := 0; x < CHR_WIDTH; x++ {
			addr := uint16((y / 8 * 0x100) + (y % 8) + (x/8)*0x10)

			// Each row has 2 bytes, byte 1 = bit 0 of color num, byte 2 = bit 1 of color num
			// Each bit pair represents the color for 1 pixel in the 8 pixel row
			tileByte1 := p.bus.read(addr)
			tileByte2 := p.bus.read(addr + 8)

			// The pixel in the current row we want the color of
			pixel := 7 - (x % 8)

			// Color number 0-3 corresponding to the 4 colors of the palette
			colorBit0 := (tileByte1 >> pixel) & 1
			colorBit1 := (tileByte2 >> pixel) & 1
			colorNum := (colorBit1 << 1) | colorBit0
			p.screen.drawPixel(int32(x)+NES_WIDTH, int32(y), COLORS[colorNum])
		}
	}
	p.screen.Update()
}

func (p *PPU) readRegister(addr uint16) uint8 {
	switch addr {

	case PPUSTATUS:
		return p.reg[2]

	case OAMDATA:
		return p.reg[4]

	case PPUDATA:
		data := p.bus.read(p.addr)
		p.addr += p.getAddrIncrement()
		return data

	default:
		return 0
	}
}

func (p *PPU) writeRegister(addr uint16, val uint8) {
	switch addr {

	case PPUCTRL:
		p.nmiOutput = bits.Test(val, 7)
		p.reg[0] = val

	case PPUMASK:
		p.reg[1] = val

	case PPUSTATUS:
		p.reg[2] = val

	case OAMADDR:
		p.reg[3] = val

	case OAMDATA:
		p.reg[4] = val

	case PPUSCROLL:
		p.reg[5] = val

	case PPUADDR:
		if p.ppuAddrLoaded {
			p.addr = (p.addr << 8) | uint16(val)
		} else {
			p.addr = uint16(val)
		}
		p.ppuAddrLoaded = !p.ppuAddrLoaded
		p.reg[6] = val

	case PPUDATA:
		p.bus.write(p.addr, val)
		p.addr += p.getAddrIncrement()
		p.reg[7] = val

	case OAMDMA:
		p.reg[8] = val
	}
}

func (p *PPU) getAddrIncrement() uint16 {
	bit2 := bits.Value(p.reg[0], 2)
	if bit2 == 0 {
		return 1
	} else {
		return 32
	}
}
