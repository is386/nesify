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
	cpu                                                              *CPU
	bus                                                              *PpuBus
	screen                                                           *Screen
	scanline, cyc                                                    int
	addr                                                             uint16
	ppuCtrl, ppuMask, ppuStatus, oamAddr, oamData, ppuScroll, oamDma uint8
	nmiOccurred, nmiOutput, ppuAddrLoaded                            bool
}

func NewPPU(b *PpuBus) *PPU {
	p := &PPU{bus: b}
	p.screen = NewScreen(NES_WIDTH+CHR_WIDTH, CHR_HEIGHT, SCALE)
	p.screen.win.SetTitle("NESify")
	p.showCHR()
	return p
}

func (p *PPU) update() {
	p.cyc++
	if p.cyc > 340 {
		p.cyc -= 341
		p.scanline++
	}

	if p.scanline >= 0 && p.scanline <= 239 {
		// drawing
	} else if p.scanline == 241 && p.cyc == 1 {
		p.setVblank()
		if p.nmiOutput && p.nmiOccurred {
			p.cpu.triggerInterrupt(Nmi)
		}
		p.renderNameTables()
	} else if p.scanline == 261 && p.cyc == 1 {
		p.resetVblank()
		p.scanline = 0
	}
}

func (p *PPU) renderNameTables() {
	ntAddr := p.getNameTableAddr()
	baseX := 0
	for ntByte := uint16(0); ntByte < 960; ntByte++ {
		ptIdx := p.bus.read(ntAddr)
		ptBaseAddr := p.getBgPatternTableAddr() + (uint16(ptIdx) * 16)

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
			tileByte1 := p.bus.read(addr)
			tileByte2 := p.bus.read(addr + 8)
			pixel := 7 - (x % 8)
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
		return p.ppuStatus

	case OAMDATA:
		return p.oamData

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
		p.ppuCtrl = val

	case PPUMASK:
		p.ppuMask = val

	case PPUSTATUS:
		p.ppuStatus = val

	case OAMADDR:
		p.oamAddr = val

	case OAMDATA:
		p.oamData = val

	case PPUSCROLL:
		p.ppuScroll = val

	case PPUADDR:
		if p.ppuAddrLoaded {
			p.addr = (p.addr << 8) | uint16(val)
		} else {
			p.addr = uint16(val)
		}
		p.ppuAddrLoaded = !p.ppuAddrLoaded

	case PPUDATA:
		p.bus.write(p.addr, val)
		p.addr += p.getAddrIncrement()

	case OAMDMA:
		p.oamDma = val
	}
}

func (p *PPU) getNameTableAddr() uint16 {
	return 0x2000 + (0x400 * uint16(p.ppuCtrl&3))
}

func (p *PPU) getBgPatternTableAddr() uint16 {
	return 0x1000 * uint16(bits.Value(p.ppuCtrl, 4))
}

func (p *PPU) getAddrIncrement() uint16 {
	return uint16(bits.Value(p.ppuCtrl, 2)*31) + 1
}

func (p *PPU) setVblank() {
	p.ppuStatus = bits.Set(p.ppuStatus, 7)
	p.nmiOccurred = true
}

func (p *PPU) resetVblank() {
	p.ppuStatus = bits.Reset(p.ppuStatus, 7)
	p.nmiOccurred = false
}
