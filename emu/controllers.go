package emu

import (
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

type Controllers struct {
	buttons1  [8]Button
	buttons2  [8]Button
	pollInput int
}

func NewControllers() *Controllers {
	return &Controllers{}
}

func (c *Controllers) update() {
	for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
		switch e := event.(type) {
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
	switch key {
	case sdl.K_RETURN:
		c.buttons1[Start] = 1
	case sdl.K_RSHIFT:
		c.buttons1[Select] = 1
	case sdl.K_w:
		c.buttons1[Up] = 1
	case sdl.K_s:
		c.buttons1[Down] = 1
	case sdl.K_a:
		c.buttons1[Left] = 1
	case sdl.K_d:
		c.buttons1[Right] = 1
	case sdl.K_j:
		c.buttons1[A] = 1
	case sdl.K_k:
		c.buttons1[B] = 1
	}
}

func (c *Controllers) keyUp(key sdl.Keycode) {
	switch key {
	case sdl.K_RETURN:
		c.buttons1[Start] = 0
	case sdl.K_RSHIFT:
		c.buttons1[Select] = 0
	case sdl.K_w:
		c.buttons1[Up] = 0
	case sdl.K_s:
		c.buttons1[Down] = 0
	case sdl.K_a:
		c.buttons1[Left] = 0
	case sdl.K_d:
		c.buttons1[Right] = 0
	case sdl.K_j:
		c.buttons1[A] = 0
	case sdl.K_k:
		c.buttons1[B] = 0
	}
}

func (c *Controllers) readController1() uint8 {
	if c.pollInput >= 0 {
		val := uint8(c.buttons1[c.pollInput])
		c.pollInput++
		if c.pollInput > 7 {
			c.pollInput = -1
		}
		return val | 0x40
	}
	return 0x40
}

func (c *Controllers) enablePolling(val uint8) {
	c.pollInput = int(val)
}
