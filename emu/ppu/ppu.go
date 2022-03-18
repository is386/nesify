package ppu

// CHR - each entry is a tile
// Tiles are 8x8, each pixel represents a color 0-3 in the palette
// Each row of a tile is represented by two bytes. The two bytes have an 8 byte offset

var COLORS = []uint32{0x000000, 0x545454, 0xA9A9A9, 0xEFEFEF}

type PPU struct {
	vram [0x8000]uint8
}

func NewPPU(chr []uint8) *PPU {
	p := &PPU{}
	for i := range chr {
		p.vram[i] = chr[i]
	}
	return p
}

func (p *PPU) ShowCHR() {
	s := NewScreen(128, 256, 3)
	for r := 0; r < 256; r++ {
		for col := 0; col < 128; col++ {
			addr := (r / 8 * 0x100) + (r % 8) + (col/8)*0x10
			pixel := ((p.vram[addr] >> (7 - (col % 8))) & 1) + ((p.vram[addr+8]>>(7-(col%8)))&1)*2
			s.drawPixel(int32(col), int32(r), COLORS[pixel])
		}
	}
	s.Update()
}
