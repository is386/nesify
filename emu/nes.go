package emu

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/is386/NESify/emu/cpu"
	"github.com/is386/NESify/emu/mem"
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
	cyc            int
	running, debug bool
}

func NewNES(romFileName string, debug bool) *NES {
	nes := &NES{debug: debug}
	mmu := mem.NewMMU()
	nes.mmu = mmu
	nes.cpu = cpu.NewCPU(mmu, debug)
	nes.loadRom(romFileName)
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

func (nes *NES) loadRom(romFileName string) {
	rom, err := ioutil.ReadFile(romFileName)
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}
	for i := 0x10; i < 0x400F; i++ {
		addr := uint16(0xC000 + i - 0x10)
		nes.mmu.Write(addr, rom[i])
	}
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
