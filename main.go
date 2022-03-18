package main

import (
	"fmt"
	"os"

	"github.com/akamensky/argparse"
	"github.com/is386/NESify/emu"
	"github.com/sqweek/dialog"
)

func parseArgs() bool {
	parser := argparse.NewParser("NESify", "A simple NES emulator written in Go.")

	debugFlag := parser.Flag("d", "debug",
		&argparse.Options{
			Required: false,
			Help:     "Turns on debugging mode",
			Default:  false,
		})

	err := parser.Parse(os.Args)
	if err != nil {
		fmt.Print(parser.Usage(err))
		os.Exit(0)
	}

	return *debugFlag
}

func main() {
	debug := parseArgs()
	romFileName, err := dialog.File().Filter("NES Rom File", "nes").Load()
	if err != nil {
		panic(err)
	}
	n := emu.NewNES(romFileName, debug)
	n.Run()

}
