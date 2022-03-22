package emu

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"
)

const (
	CLOCK_SPEED = 1789773
	FPS         = 60
	FRAMETIME   = time.Second / time.Duration(FPS)
	CPS         = CLOCK_SPEED / FPS
)

type NES struct {
	cpu            *CPU
	ppu            *PPU
	cyc            int
	running, debug bool
}

func NewNES(romFileName string, debug bool) *NES {
	nes := &NES{debug: debug}
	rom := nes.loadRom(romFileName)
	cart := NewCart(rom)
	nes.ppu = NewPPU(NewPpuBus(cart))
	nes.cpu = NewCPU(NewCpuBus(cart, nes.ppu), debug)
	nes.ppu.cpu = nes.cpu
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
		cpuCyc := nes.cpu.update()
		nes.cyc += cpuCyc
		for i := 0; i < cpuCyc; i++ {
			nes.ppu.update()
		}
	}
	nes.cyc -= CPS
}

func (nes *NES) shutdown() {
	nes.running = false
}
