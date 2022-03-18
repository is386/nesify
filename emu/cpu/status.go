package cpu

import "github.com/is386/NESify/emu/bits"

type Status struct {
	n, v, bit5, b, d, i, z, c uint8
}

func NewStatus() *Status {
	return &Status{i: 1, bit5: 1}
}

func (f *Status) getStatus() uint8 {
	status := uint8(0)
	if f.n == 1 {
		status = bits.Set(status, 7)
	}
	if f.v == 1 {
		status = bits.Set(status, 6)
	}
	if f.bit5 == 1 {
		status = bits.Set(status, 5)
	}
	if f.d == 1 {
		status = bits.Set(status, 3)
	}
	if f.i == 1 {
		status = bits.Set(status, 2)
	}
	if f.z == 1 {
		status = bits.Set(status, 1)
	}
	if f.c == 1 {
		status = bits.Set(status, 0)
	}
	return status
}

func (f *Status) setStatus(status uint8) {
	f.n = bits.Value(status, 7)
	f.v = bits.Value(status, 6)
	f.d = bits.Value(status, 3)
	f.i = bits.Value(status, 2)
	f.z = bits.Value(status, 1)
	f.c = bits.Value(status, 0)
}

func (f *Status) getNegative() uint8 {
	return f.n
}

func (f *Status) setNegative() {
	f.n = 1
}

func (f *Status) resetNegative() {
	f.n = 0
}

func (f *Status) checkNegative(val uint8) {
	bit7 := bits.Value(val, 7)
	if bit7 == 1 {
		f.setNegative()
	} else {
		f.resetNegative()
	}
}

func (f *Status) getOverflow() uint8 {
	return f.v
}

func (f *Status) setOverflow() {
	f.v = 1
}

func (f *Status) resetOverflow() {
	f.v = 0
}

func (f *Status) checkOverflow(a, b, total uint16) {
	if (((a ^ b) & 0x80) == 0) && (((a ^ total) & 0x80) != 0) {
		f.setOverflow()
	} else {
		f.resetOverflow()
	}
}

func (f *Status) checkUnderflow(a, b, total uint16) {
	if (((a ^ b) & 0x80) != 0) && (((a ^ total) & 0x80) != 0) {
		f.setOverflow()
	} else {
		f.resetOverflow()
	}
}

func (f *Status) setBit5() {
	f.bit5 = 1
}

func (f *Status) resetBit5() {
	f.bit5 = 0
}

func (f *Status) setBreak() {
	f.b = 1
}

func (f *Status) resetBreak() {
	f.b = 0
}

func (f *Status) setDecimal() {
	f.d = 1
}

func (f *Status) resetDecimal() {
	f.d = 0
}

func (f *Status) setInterrupt() {
	f.i = 1
}

func (f *Status) resetInterrupt() {
	f.i = 0
}

func (f *Status) getZero() uint8 {
	return f.z
}

func (f *Status) setZero() {
	f.z = 1
}

func (f *Status) resetZero() {
	f.z = 0
}

func (f *Status) checkZero(val uint8) {
	if val == 0 {
		f.setZero()
	} else {
		f.resetZero()
	}
}

func (f *Status) getCarry() uint8 {
	return f.c
}

func (f *Status) setCarry() {
	f.c = 1
}

func (f *Status) resetCarry() {
	f.c = 0
}

func (f *Status) checkCarry(val uint16, bound uint16) {
	if val >= bound {
		f.setCarry()
	} else {
		f.resetCarry()
	}
}

func (f *Status) checkBorrow(val int, bound int) {
	if val >= bound {
		f.setCarry()
	} else {
		f.resetCarry()
	}
}
