package emu

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/is386/NESify/emu/cpu"
	"github.com/is386/NESify/emu/mem"
	"github.com/is386/NESify/emu/ppu"
)

const (
	CLOCK_SPEED = 1789773
	FPS         = 60
	FRAMETIME   = time.Second / time.Duration(FPS)
	CPS         = CLOCK_SPEED / FPS
)

type NES struct {
	cpu            *cpu.CPU
	mmu            *mem.MMU
	ppu            *ppu.PPU
	cyc            int
	running, debug bool
}

func NewNES(romFileName string, debug bool) *NES {
	nes := &NES{debug: debug}
	rom := nes.loadRom(romFileName)
	mmu, chr := mem.NewMMU(rom)
	nes.mmu = mmu
	nes.cpu = cpu.NewCPU(mmu, debug)
	nes.ppu = ppu.NewPPU(chr)
	return nes
}

func (nes *NES) Run() {
	ticker := time.NewTicker(FRAMETIME)
	nes.running = true

	for range ticker.C {
		if !nes.running {
			nes.shutdown()
		}
		nes.update()
	}
}

func (nes *NES) loadRom(romFileName string) []uint8 {
	rom, err := ioutil.ReadFile(romFileName)
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}
	return rom
}

func (nes *NES) update() {
	for nes.cyc < CPS {
		nes.cyc += nes.cpu.Update()
	}
	nes.cyc -= CPS
}

func (nes *NES) shutdown() {
	nes.running = false
}
