package emu

import (
	"github.com/is386/NESify/emu/bits"
)

// TODO:
// - Sprite overlap priority
// - 8x16 sprites
// - Scrolling
// - NameTable mirroring
// - CPU stalling after DMA

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
	COLORS = []uint32{
		0x666666, 0x002A88, 0x1412A7, 0x3B00A4, 0x5C007E, 0x6E0040, 0x6C0600, 0x561D00,
		0x333500, 0x0B4800, 0x005200, 0x004F08, 0x00404D, 0x000000, 0x000000, 0x000000,
		0xADADAD, 0x155FD9, 0x4240FF, 0x7527FE, 0xA01ACC, 0xB71E7B, 0xB53120, 0x994E00,
		0x6B6D00, 0x388700, 0x0C9300, 0x008F32, 0x007C8D, 0x000000, 0x000000, 0x000000,
		0xFFFEFF, 0x64B0FF, 0x9290FF, 0xC676FF, 0xF36AFF, 0xFE6ECC, 0xFE8170, 0xEA9E22,
		0xBCBE00, 0x88D800, 0x5CE430, 0x45E082, 0x48CDDE, 0x4F4F4F, 0x000000, 0x000000,
		0xFFFEFF, 0xC0DFFF, 0xD3D2FF, 0xE8C8FF, 0xFBC2FF, 0xFEC4EA, 0xFECCC5, 0xF7D8A5,
		0xE4E594, 0xCFEF96, 0xBDF4AB, 0xB3F3CC, 0xB5EBF2, 0xB8B8B8, 0x000000, 0x000000,
	}
)

type PPU struct {
	cpu                                                         *CPU
	bus                                                         *PpuBus
	screen                                                      *Screen
	bgPixels                                                    [NES_WIDTH][NES_HEIGHT]uint8
	scanline, cyc                                               int
	addr                                                        uint16
	ppuCtrl, ppuMask, ppuStatus, oamAddr, ppuScroll, dataBuffer uint8
	nmiOccurred, nmiOutput, ppuAddrLoaded                       bool
}

func NewPPU(b *PpuBus) *PPU {
	p := &PPU{bus: b}
	p.screen = NewScreen(NES_WIDTH, NES_HEIGHT, SCALE)
	p.screen.win.SetTitle("NESify")
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
		p.renderBackground()
		p.renderSprites()
	} else if p.scanline == 261 && p.cyc == 1 {
		p.resetVblank()
		p.scanline = 0
		p.bgPixels = [NES_WIDTH][NES_HEIGHT]uint8{}
	}
}

func (p *PPU) renderBackground() {
	ntAddr := p.getNameTableAddr()
	baseX := 0
	for ntByte := uint16(0); ntByte < 960; ntByte++ {
		ptIdx := p.bus.read(ntAddr + ntByte)
		ptBaseAddr := p.getBgPatternTableAddr() + (uint16(ptIdx) * 16)

		for ptAddr := ptBaseAddr; ptAddr < ptBaseAddr+8; ptAddr++ {
			ptByte1 := p.bus.read(ptAddr)
			ptByte2 := p.bus.read(ptAddr + 8)
			y := int(ptAddr-ptBaseAddr) + int(ntByte/32)*8

			for x := baseX; x < baseX+8; x++ {
				pixel := 7 - (x % 8)
				colorBit0 := (ptByte1 >> pixel) & 1
				colorBit1 := (ptByte2 >> pixel) & 1
				colorNum := (colorBit1 << 1) | colorBit0
				p.bgPixels[x][y] = colorNum

				blockX := x / 32
				blockY := y / 32
				blockAddr := uint16(8*blockY) + uint16(blockX) + ntAddr + 0x3C0
				blockByte := p.bus.read(blockAddr)

				quadX := x / 16
				quadY := y / 16
				quad := uint8(((quadY % 2) << 1) | (quadX % 2))

				paletteBit0 := bits.Value(blockByte, (quad * 2))
				paletteBit1 := bits.Value(blockByte, (quad*2)+1)
				paletteNum := (paletteBit1 << 1) | paletteBit0
				color := p.getPalette(int(paletteNum))[colorNum]
				p.screen.drawPixel(int32(x), int32(y), color)
			}
		}
		baseX = (baseX + 8) % NES_WIDTH
	}
}

func (p *PPU) renderSprites() {
	for oamAddr := 0; oamAddr < 256; oamAddr += 4 {
		spriteY := int(p.bus.readOam(uint8(oamAddr)))
		spriteX := int(p.bus.readOam(uint8(oamAddr) + 3))

		attrs := p.bus.readOam(uint8(oamAddr) + 2)
		paletteNum := (attrs & 3) + 4
		bgPriority := bits.Test(attrs, 5)
		xFlip := bits.Test(attrs, 6)
		yFlip := bits.Test(attrs, 7)

		tileIdx := p.bus.readOam(uint8(oamAddr) + 1)
		ptBaseAddr := p.getSpritePatternTableAddr() + (uint16(tileIdx) * 16)

		for ptAddr := ptBaseAddr; ptAddr < ptBaseAddr+8; ptAddr++ {
			if spriteY >= NES_HEIGHT {
				break
			}

			y := spriteY
			if yFlip {
				y = 8 - spriteY - 1
			}

			ptByte1 := p.bus.read(ptAddr)
			ptByte2 := p.bus.read(ptAddr + 8)
			diff := (8 - (spriteX % 8))

			for x := spriteX; x < spriteX+8; x++ {
				if x >= NES_WIDTH {
					break
				}

				pixel := 7 - ((x + diff) % 8)
				if xFlip {
					pixel = 7 - pixel
				}
				colorBit0 := (ptByte1 >> pixel) & 1
				colorBit1 := (ptByte2 >> pixel) & 1
				colorNum := (colorBit1 << 1) | colorBit0

				if colorNum == 0 || (bgPriority && p.bgPixels[x][spriteY] != 0) {
					continue
				}

				color := p.getPalette(int(paletteNum))[colorNum]
				p.screen.drawPixel(int32(x), int32(y), color)
			}
			spriteY++
		}
	}
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
			p.screen.drawPixel(int32(x)+NES_WIDTH, int32(y), COLORS[colorNum*3])
		}
	}
	p.screen.Update()
}

func (p *PPU) readRegister(addr uint16) uint8 {
	switch addr {

	case PPUSTATUS:
		return p.ppuStatus

	case OAMDATA:
		data := p.bus.readOam(p.oamAddr)
		if (p.oamAddr & 3) == 2 {
			data &= 0xE3
		}
		return data

	case PPUDATA:
		data := p.bus.read(p.addr)
		if p.addr < 0x3F00 {
			data, p.dataBuffer = p.dataBuffer, data
		}
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
		p.bus.writeOam(p.oamAddr, val)
		p.oamAddr++

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
		cpuAddr := uint16(val) << 8
		for i := uint16(0); i < 256; i++ {
			p.bus.writeOam(p.oamAddr, p.cpu.read(cpuAddr+i))
			p.oamAddr++
		}
	}
}

func (p *PPU) getNameTableAddr() uint16 {
	return 0x2000 + (0x400 * uint16(p.ppuCtrl&3))
}

func (p *PPU) getBgPatternTableAddr() uint16 {
	return 0x1000 * uint16(bits.Value(p.ppuCtrl, 4))
}

func (p *PPU) getSpritePatternTableAddr() uint16 {
	return 0x1000 * uint16(bits.Value(p.ppuCtrl, 3))
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

func (p *PPU) getPalette(num int) [4]uint32 {
	palette := [4]uint32{COLORS[p.bus.read(uint16(0x3F00))]}
	addr := 0x3F00 + (num * 4)
	for i := addr + 1; i < addr+4; i++ {
		paletteByte := p.bus.read(uint16(i))
		palette[i-addr] = COLORS[paletteByte]
	}
	return palette
}
