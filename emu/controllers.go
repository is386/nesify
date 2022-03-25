package emu

import (
	"os"

	"github.com/veandco/go-sdl2/sdl"
)

type Button uint8

const (
	A Button = iota
	B
	Select
	Start
	Up
	Down
	Left
	Right
)

var (
	buttonMap = map[sdl.Keycode]Button{
		sdl.K_RETURN: Start,
		sdl.K_RSHIFT: Select,
		sdl.K_w:      Up,
		sdl.K_s:      Down,
		sdl.K_a:      Left,
		sdl.K_d:      Right,
		sdl.K_j:      A,
		sdl.K_k:      B,
	}
)

type Controllers struct {
	buttons   [8]Button
	pollInput int
}

func NewControllers() *Controllers {
	return &Controllers{}
}

func (c *Controllers) update() {
	for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
		switch e := event.(type) {
		case *sdl.QuitEvent:
			os.Exit(0)
		case *sdl.KeyboardEvent:
			switch e.Type {
			case sdl.KEYDOWN:
				c.keyDown(e.Keysym.Sym)
			case sdl.KEYUP:
				c.keyUp(e.Keysym.Sym)
			}
		}
	}
}

func (c *Controllers) keyDown(key sdl.Keycode) {
	c.buttons[buttonMap[key]] = 1
}

func (c *Controllers) keyUp(key sdl.Keycode) {
	c.buttons[buttonMap[key]] = 0
}

func (c *Controllers) readController1() uint8 {
	if c.pollInput >= 0 {
		val := uint8(c.buttons[c.pollInput])
		c.pollInput++
		if c.pollInput > 7 {
			c.pollInput = -1
		}
		return val | 0x40
	}
	return 0x40
}

func (c *Controllers) enablePolling(val uint8) {
	if val != 0 {
		c.pollInput = 0
	}
}
