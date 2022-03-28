package emu

import (
	"github.com/is386/NESify/emu/bits"
)

// TODO:
// - Sprite overlap priority
// - 8x16 sprites
// - NameTable mirroring
// - Register Sharing

const (
	NES_WIDTH  = 256
	NES_HEIGHT = 240
	CHR_WIDTH  = 128
	CHR_HEIGHT = 256
	SCALE      = 2
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
	cpu                                              *CPU
	bus                                              *PpuBus
	screen                                           *Screen
	bgPixels                                         [NES_WIDTH][NES_HEIGHT]uint8
	scanline, cyc, scrollX, scrollY                  int
	addr                                             uint16
	ppuCtrl, ppuMask, ppuStatus, oamAddr, dataBuffer uint8
	nmiOccurred, nmiOutput, isSecondWrite            bool
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
		if p.cyc == 230 {
			p.renderBackground()
			p.renderSprites()
		}
	} else if p.scanline == 241 && p.cyc == 1 {
		p.enterVblank()
	} else if p.scanline == 261 && p.cyc == 1 {
		p.exitVblank()
	}
}

func (p *PPU) renderBackground() {
	for x := 0; x < NES_WIDTH; x++ {
		scrolledX := (x + p.scrollX) % 256
		scrolledY := (p.scanline + p.scrollY) % 240

		ntX := (x + p.scrollX) / 8
		ntY := (p.scanline + p.scrollY) / 8
		ntBaseAddr := p.getNameTableAddr()
		ntAddr := uint16((ntY*32)+ntX) + ntBaseAddr

		if ntY > 29 {
			ntAddr = ((ntAddr + 0x40) ^ 0x800) ^ 0x400
			ntBaseAddr ^= 0x800
		}
		if ntX > 31 {
			ntAddr = (ntAddr ^ 0x400) - 0x20
			ntBaseAddr ^= 0x400
		}

		ptIdx := p.bus.read(ntAddr)
		ptAddr := p.getBgPatternTableAddr() + (uint16(ptIdx) * 16) + uint16(scrolledY%8)
		ptByte1 := p.bus.read(ptAddr)
		ptByte2 := p.bus.read(ptAddr + 8)

		pixel := 7 - (scrolledX % 8)
		colorBit0 := (ptByte1 >> pixel) & 1
		colorBit1 := (ptByte2 >> pixel) & 1
		colorNum := (colorBit1 << 1) | colorBit0
		p.bgPixels[x][p.scanline] = colorNum

		blockX := scrolledX / 32
		blockY := scrolledY / 32
		blockAddr := uint16(8*blockY) + uint16(blockX) + ntBaseAddr + 0x3C0
		blockByte := p.bus.read(blockAddr)

		quadX := scrolledX / 16
		quadY := scrolledY / 16
		quad := uint8(((quadY % 2) << 1) | (quadX % 2))

		paletteBit0 := bits.Value(blockByte, (quad * 2))
		paletteBit1 := bits.Value(blockByte, (quad*2)+1)
		paletteNum := (paletteBit1 << 1) | paletteBit0
		color := p.getPalette(int(paletteNum))[colorNum]
		p.screen.drawPixel(int32(x), int32(p.scanline), color)
	}
}

func (p *PPU) renderSprites() {
	for oamAddr := 0; oamAddr < 256; oamAddr += 4 {
		spriteHeight := 8
		spriteY := int(p.bus.readOam(uint8(oamAddr)))
		spriteX := int(p.bus.readOam(uint8(oamAddr) + 3))

		if p.scanline < spriteY || p.scanline >= (spriteY+spriteHeight) {
			continue
		}

		if oamAddr == 0 && spriteX < 255 {
			p.setZeroHit()
		}

		attrs := p.bus.readOam(uint8(oamAddr) + 2)
		paletteNum := (attrs & 3) + 4
		bgPriority := bits.Test(attrs, 5)
		xFlip := bits.Test(attrs, 6)
		yFlip := bits.Test(attrs, 7)

		y := p.scanline - spriteY
		if yFlip {
			y = spriteHeight - y - 1
		}

		tileIdx := p.bus.readOam(uint8(oamAddr) + 1)
		ptAddr := p.getSpritePatternTableAddr() + (uint16(tileIdx) * 16) + uint16(y%8)
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

			if colorNum == 0 || (bgPriority && p.bgPixels[x][p.scanline] != 0) {
				continue
			}

			color := p.getPalette(int(paletteNum))[colorNum]
			p.screen.drawPixel(int32(x), int32(p.scanline), color)
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
	p.screen.update()
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
		if p.isSecondWrite {
			p.scrollY = int(val)
		} else {
			p.scrollX = int(val)
		}
		p.isSecondWrite = !p.isSecondWrite

	case PPUADDR:
		if p.isSecondWrite {
			p.addr = (p.addr << 8) | uint16(val)
		} else {
			p.addr = uint16(val)
		}
		p.isSecondWrite = !p.isSecondWrite

	case PPUDATA:
		p.bus.write(p.addr, val)
		p.addr += p.getAddrIncrement()

	case OAMDMA:
		cpuAddr := uint16(val) << 8
		for i := uint16(0); i < 256; i++ {
			p.bus.writeOam(p.oamAddr, p.cpu.read(cpuAddr+i))
			p.oamAddr++
		}
		p.cpu.stallForDma()
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

func (p *PPU) enterVblank() {
	p.ppuStatus = bits.Set(p.ppuStatus, 7)
	p.nmiOccurred = true
	if p.nmiOutput && p.nmiOccurred {
		p.cpu.triggerInterrupt(Nmi)
	}
}

func (p *PPU) exitVblank() {
	p.ppuStatus = bits.Reset(p.ppuStatus, 7)
	p.resetZeroHit()
	p.bgPixels = [NES_WIDTH][NES_HEIGHT]uint8{}
	p.scanline = 0
	p.nmiOccurred = false
	p.screen.update()
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

func (p *PPU) setZeroHit() {
	p.ppuStatus = bits.Set(p.ppuStatus, 6)
}

func (p *PPU) resetZeroHit() {
	p.ppuStatus = bits.Reset(p.ppuStatus, 6)
}
