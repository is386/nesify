package main

import (
	"github.com/is386/NESify/emu"
	"github.com/sqweek/dialog"
)

func main() {
	romFileName, err := dialog.File().Filter("NES Rom File", "nes").Load()
	if err != nil {
		panic(err)
	}
	n := emu.NewNES(romFileName)
	n.Run()
}
