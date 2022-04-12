package chip8

import (
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"time"
	"unicode"
)

const BASE = 0x200

type VirtualMachine struct {
	ROM [0x1000]byte

	Memory [0x1000]byte
	Video  [32][64]byte

	Stack [16]uint

	SP uint

	PC uint

	Base uint

	Size int

	I uint

	V [16]byte

	R [8]byte

	DT byte

	ST byte

	Clock int64

	Cycles int64

	Speed int64

	W *byte

	Keys [16]bool

	Pitch int
}

func LoadROM(program []byte, eti bool) (*VirtualMachine, error) {
	base := BASE

	if len(program) > 0x1000-base {
		return nil, errors.New("Program too large to fit int memory!")
	}

	vm := &VirtualMachine{
		Size:  len(program),
		Base:  uint(BASE),
		Speed: 500,
	}

	copy(vm.ROM[:base], EmulatorROM[:])
	copy(vm.ROM[base:], program[:])

	vm.Reset()

	return vm, nil
}

func LoadFromFile(filePath string) (*VirtualMachine, error) {
	file, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	for _, c := range string(file) {
		if unicode.IsSpace(c) || unicode.IsGraphic(c) {
			continue
		}
	}

	return LoadROM(file, false)
}

func (vm *VirtualMachine) Reset() {
	copy(vm.Memory[:], vm.ROM[:])
	vm.Video = [len(vm.Video)][len(vm.Video[0])]byte{}

	vm.Keys = [16]bool{}
	vm.PC = vm.Base
	vm.SP = 0

	vm.I = 0

	vm.V = [16]byte{}
	vm.R = [8]byte{}

	vm.DT = 0
	vm.ST = 0

	vm.Clock = time.Now().UnixNano()
	vm.Cycles = 0

	vm.W = nil

	vm.Pitch = 8
}

func (vm *VirtualMachine) Step() error {
	if vm.W != nil {
		return nil
	}

	instruction := vm.fetch()

	a := instruction & 0xFFF

	b := byte(instruction & 0xFF)
	n := byte(instruction & 0xF)

	x := instruction >> 8 & 0xF
	y := instruction >> 4 & 0xF

	switch {
	case instruction == 0x00E0:
		vm.cls()
	case instruction == 0x00EE:
		vm.ret()
	case instruction&0xF000 == 0x1000:
		vm.jump(a)
	case instruction&0xF000 == 0x2000:
		vm.call(a)
	case instruction&0xF000 == 0x3000:
		vm.skipIf(x, b)
	case instruction&0xF000 == 0x4000:
		vm.skipIfNot(x, b)
	case instruction&0xF000 == 0x5000:
		vm.skipIfXY(x, y)
	case instruction&0xF000 == 0x6000:
		vm.loadX(x, b)
	case instruction&0xF000 == 0x7000:
		vm.addX(x, b)
	case instruction&0xF00F == 0x8000:
		vm.loadXY(x, y)
	case instruction&0xF00F == 0x8001:
		vm.or(x, y)
	case instruction&0xF00F == 0x8002:
		vm.and(x, y)
	case instruction&0xF00F == 0x8003:
		vm.xor(x, y)
	case instruction&0xF00F == 0x8004:
		vm.addXY(x, y)
	case instruction&0xF00F == 0x8005:
		vm.subXY(x, y)
	case instruction&0xF00F == 0x8006:
		vm.shr(x)
	case instruction&0xF00F == 0x8007:
		vm.subnXY(x, y)
	case instruction&0xF00F == 0x800E:
		vm.shl(x)
	case instruction&0xF00F == 0x9000:
		vm.skipIfNotXY(x, y)
	case instruction&0xF000 == 0xA000:
		vm.loadI(a)
	case instruction&0xF000 == 0xB000:
		vm.jumpV0(a)
	case instruction&0xF000 == 0xC000:
		vm.random(x, b)
	case instruction&0xF000 == 0xD000:
		vm.drawSprite(x, y, n)
	case instruction&0xF0FF == 0xE09E:
		vm.skipIfPressed(x)
	case instruction&0xF0FF == 0xE0A1:
		vm.skipIfNotPressed(x)
	case instruction&0xF0FF == 0xF007:
		vm.loadXDT(x)
	case instruction&0xF0FF == 0xF00A:
		vm.loadXK(x)
	case instruction&0xF0FF == 0xF015:
		vm.loadDTX(x)
	case instruction&0xF0FF == 0xF018:
		vm.loadSTX(x)
	case instruction&0xF0FF == 0xF01E:
		vm.addIX(x)
	case instruction&0xF0FF == 0xF029:
		vm.loadF(x)
	case instruction&0xF0FF == 0xF033:
		vm.bcd(x, y)
	case instruction&0xF0FF == 0xF055:
		vm.saveRegs(x)
	case instruction&0xF0FF == 0xF065:
		vm.loadRegs(x)
	default:
		// return fmt.Errorf("Invalid opcode: %04X", instruction)
		panic(fmt.Sprintf("Invalid opcode: %04X", instruction))
	}

	vm.Cycles += 1

	return nil
}

func (vm *VirtualMachine) fetch() uint {
	i := vm.PC

	vm.PC += 2

	return uint(vm.Memory[i])<<8 | uint(vm.Memory[i+1])
}

func (vm *VirtualMachine) cls() {
	for i := 0; i < len(vm.Video); i++ {
		for j := 0; j < len(vm.Video[i]); j++ {
			vm.Video[i][j] = 0x0
		}
	}
}

func (vm *VirtualMachine) ret() {
	if vm.SP == 0 {
		panic("Stack underflow!")
	}

	vm.SP--
	vm.PC = vm.Stack[vm.SP]
}

func (vm *VirtualMachine) jump(address uint) {
	vm.PC = address
}

func (vm *VirtualMachine) call(address uint) {
	if int(vm.SP) >= len(vm.Stack) {
		panic("Stack overflow!")
	}

	vm.Stack[vm.SP] = vm.PC
	vm.SP++

	vm.PC = address
}

func (vm *VirtualMachine) skipIf(x uint, b byte) {
	if vm.V[x] == b {
		vm.PC += 2
	}
}

func (vm *VirtualMachine) skipIfNot(x uint, b byte) {
	if vm.V[x] != b {
		vm.PC += 2
	}
}

func (vm *VirtualMachine) skipIfXY(x, y uint) {
	if vm.V[x] == vm.V[y] {
		vm.PC += 2
	}
}

func (vm *VirtualMachine) loadX(x uint, b byte) {
	vm.V[x] = b
}

func (vm *VirtualMachine) addX(x uint, b byte) {
	vm.V[x] += b
}

func (vm *VirtualMachine) loadXY(x, y uint) {
	vm.V[x] = vm.V[y]
}

func (vm *VirtualMachine) or(x, y uint) {
	vm.V[x] |= vm.V[y]
}

func (vm *VirtualMachine) and(x, y uint) {
	vm.V[x] &= vm.V[y]
}

func (vm *VirtualMachine) xor(x, y uint) {
	vm.V[x] ^= vm.V[y]
}

func (vm *VirtualMachine) addXY(x, y uint) {
	vm.V[x] += vm.V[y]

	if vm.V[x] < vm.V[y] {
		vm.V[0xF] = 1
	} else {
		vm.V[0xF] = 0
	}
}

func (vm *VirtualMachine) subXY(x, y uint) {
	if vm.V[x] >= vm.V[y] {
		vm.V[0xF] = 1
	} else {
		vm.V[0xF] = 0
	}

	vm.V[x] -= vm.V[y]
}

func (vm *VirtualMachine) shr(x uint) {
	vm.V[0xF] = vm.V[x] & 0x1

	vm.V[x] >>= 1
}

func (vm *VirtualMachine) subnXY(x, y uint) {
	if vm.V[y] >= vm.V[x] {
		vm.V[0xF] = 1
	} else {
		vm.V[0xF] = 0
	}

	vm.V[x] = vm.V[y] - vm.V[x]
}

func (vm *VirtualMachine) shl(x uint) {
	vm.V[0xF] = vm.V[x] >> 7
	vm.V[x] <<= 1
}

func (vm *VirtualMachine) skipIfNotXY(x, y uint) {
	if vm.V[x] != vm.V[y] {
		vm.PC += 2
	}
}

func (vm *VirtualMachine) loadI(address uint) {
	vm.I = address
}

func (vm *VirtualMachine) jumpV0(address uint) {
	vm.PC = address + uint(vm.V[0x0])
}

func (vm *VirtualMachine) random(x uint, b byte) {
	vm.V[x] = byte(rand.Intn(256) & int(b))
}

func (vm *VirtualMachine) drawSprite(x, y uint, n byte) {
	vm.V[0xF] = 0

	var i, j byte
	maxY := byte(len(vm.Video))
	maxX := byte(len(vm.Video[0]))

	for j = 0; j < n; j++ {
		pixel := vm.Memory[vm.I+uint(j)]

		for i = 0; i < 8; i++ {
			if (pixel & (0x80 >> i)) != 0 {
				wrapY := vm.V[y] + j
				if wrapY >= maxY {
					wrapY %= byte(maxY)
				}
				wrapX := vm.V[x] + i
				if wrapX >= maxX {
					wrapX %= byte(maxX)
				}

				if vm.Video[wrapY][wrapX] == 1 {
					vm.V[0xF] = 1
				}
				vm.Video[wrapY][wrapX] ^= 1
			}
		}
	}
}

func (vm *VirtualMachine) skipIfPressed(x uint) {
	if vm.Keys[vm.V[x]] {
		vm.PC += 2
	}
}

func (vm *VirtualMachine) skipIfNotPressed(x uint) {
	if !vm.Keys[vm.V[x]] {
		vm.PC += 2
	}
}

func (vm *VirtualMachine) loadXDT(x uint) {
	vm.V[x] = vm.DT
}

func (vm *VirtualMachine) loadXK(x uint) {
	vm.W = &vm.V[x]
}

func (vm *VirtualMachine) loadDTX(x uint) {
	vm.DT = vm.V[x]
}

func (vm *VirtualMachine) loadSTX(x uint) {
	vm.ST = vm.V[x]
}

func (vm *VirtualMachine) addIX(x uint) {
	vm.I += uint(vm.V[x])

	if vm.I >= 0x1000 {
		vm.V[0xF] = 1
	} else {
		vm.V[0xF] = 0
	}
}

func (vm *VirtualMachine) loadF(x uint) {
	vm.I = uint(vm.V[x]) * 5
}

func (vm *VirtualMachine) bcd(x, y uint) {
	n := uint(vm.V[x])
	b := uint(0)

	for i := uint(0); i < 8; i++ {
		if (b>>0)&0xF >= 5 {
			b += 3
		}
		if (b>>4)&0xF >= 5 {
			b += 3 << 4
		}
		if (b>>8)&0xF >= 5 {
			b += 3 << 8
		}

		b = (b << 1) | (n >> (7 - i) & 1)
	}

	vm.Memory[vm.I+0] = byte(b>>8) & 0xF
	vm.Memory[vm.I+1] = byte(b>>4) & 0xF
	vm.Memory[vm.I+2] = byte(b>>0) & 0xF
}

func (vm *VirtualMachine) saveRegs(x uint) {
	for i := uint(0); i <= x; i++ {
		if vm.I+i < 0x1000 {
			vm.Memory[vm.I+i] = vm.V[i]
		}
	}
}

func (vm *VirtualMachine) loadRegs(x uint) {
	for i := uint(0); i <= x; i++ {
		if vm.I+i < 0x1000 {
			vm.V[i] = vm.Memory[vm.I+i]
		} else {
			vm.V[i] = 0
		}
	}
}

func (vm *VirtualMachine) PressKey(key uint) {
	if key < 16 {
		vm.Keys[key] = true

		if vm.W != nil {
			*vm.W = byte(key)

			vm.W = nil
		}
	}
}

func (vm *VirtualMachine) ReleasedKey(key uint) {
	if key < 16 {
		vm.Keys[key] = false
	}
}
