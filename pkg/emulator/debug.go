package emulator

import (
	"fmt"
	"gbc/pkg/gpu"
	"gbc/pkg/util"
	"image/jpeg"
	"os"
)

// Debug - Info used in debug mode
type Debug struct {
	on          bool
	breakpoints []BreakPoint
	history     History
}

// BreakPoint - A Breakpoint info used in debug mode
type BreakPoint struct {
	PC   uint16
	Cond string
}

// History - CPU instruction log
type History struct {
	ptr    uint
	buffer [10]string
}

func (cpu *CPU) parseBreakpoints() {
	for _, s := range cpu.Config.Debug.BreakPoints {
		if bk, ok := newBreakPoint(s); ok {
			cpu.debug.breakpoints = append(cpu.debug.breakpoints, bk)
		}
	}
}

func newBreakPoint(s string) (bk BreakPoint, ok bool) {
	bk = BreakPoint{}
	ok = false
	return bk, ok
}

func (cpu *CPU) pushHistory(opcode byte) {
	PC := fmt.Sprintf("%04x: ", cpu.Reg.PC)
	instruction, operand1, operand2 := opcodeToString[opcode][0], opcodeToString[opcode][1], opcodeToString[opcode][2]
	history := &cpu.debug.history
	switch {
	case operand1 == "*" && operand2 == "*":
		history.buffer[history.ptr] = PC + instruction
	case operand2 == "*":
		history.buffer[history.ptr] = PC + instruction + " " + operand1
	default:
		history.buffer[history.ptr] = PC + instruction + " " + operand1 + ", " + operand2
	}
	history.ptr = (history.ptr + 1) % 10
}

func (cpu *CPU) debugHistory() string {
	result := "History\n"
	history := &cpu.debug.history
	for i := -9; i <= 0; i++ {
		index := (history.ptr + uint(i)) % 10
		log := history.buffer[index]
		if i < 0 {
			result += fmt.Sprintf("%d:    %0s\n", i, log)
		} else if i == 0 {
			result += fmt.Sprintf(" %d:    %0s\n", i, log)
		}
	}
	return result
}

func (cpu *CPU) debugRegister() string {
	A, F := byte(cpu.Reg.AF>>8), byte(cpu.Reg.AF)
	B, C := byte(cpu.Reg.BC>>8), byte(cpu.Reg.BC)
	D, E := byte(cpu.Reg.DE>>8), byte(cpu.Reg.DE)
	H, L := byte(cpu.Reg.HL>>8), byte(cpu.Reg.HL)

	return fmt.Sprintf(`Register
A: %02x       F: %02x
B: %02x       C: %02x
D: %02x       E: %02x
H: %02x       L: %02x
PC: 0x%04x  SP: 0x%04x`, A, F, B, C, D, E, H, L, cpu.Reg.PC, cpu.Reg.SP)
}

func (cpu *CPU) debugIOMap() string {
	LCDC := cpu.FetchMemory8(LCDCIO)
	STAT := cpu.FetchMemory8(LCDSTATIO)
	LY, LYC := cpu.FetchMemory8(LYIO), cpu.FetchMemory8(LYCIO)
	IE, IF, IME := cpu.FetchMemory8(IEIO), cpu.FetchMemory8(IFIO), util.Bool2Int(cpu.Reg.IME)
	spd := cpu.boost / 2
	rom := cpu.ROMBankPtr
	return fmt.Sprintf(`IO
LCDC: %02x   STAT: %02x
LY: %02x     LYC: %02x
IE: %02x     IF: %02x    IME: %02x
SPD: %02x    ROM: %02x`, LCDC, STAT, LY, LYC, IE, IF, IME, spd, rom)
}

// DebugExec - used in test
func (cpu *CPU) DebugExec(frame int, output string) error {
	const (
		WX, WY, scrollX, scrollY, scrollPixelX = 0, 0, 0, 0, 0
	)

	for i := 0; i < frame; i++ {
		for y := 0; y < 144; y++ {
			cpu.execScanline()
		}
		cpu.execVBlank()
	}

	// 最後の1frameは背景データを生成する
	for y := 0; y < 144; y++ {
		cpu.execScanline()

		LCDC := cpu.FetchMemory8(LCDCIO)
		for x := 0; x < 160; x += 8 {
			blockX := x / 8
			blockY := y / 8

			var tileX, tileY uint
			var useWindow bool
			var entryX int

			lineNumber := y % 8 // タイルの何行目を描画するか
			entryY := gpu.EntryY{}
			if util.Bit(LCDC, 5) == 1 && (WY <= uint(y)) && (WX <= uint(x)) {
				tileX = ((uint(x) - WX) / 8) % 32
				tileY = ((uint(y) - WY) / 8) % 32
				useWindow = true

				entryX = blockX * 8
				entryY.Block = blockY * 8
				entryY.Offset = y % 8
			} else {
				tileX = (scrollX + uint(x)) / 8 % 32
				tileY = (scrollY + uint(y)) / 8 % 32
				useWindow = false

				entryX = blockX*8 - int(scrollPixelX)
				entryY.Block = blockY * 8
				entryY.Offset = y % 8
				lineNumber = (int(scrollY) + y) % 8
			}

			if util.Bit(LCDC, 7) == 1 {
				if !cpu.GPU.SetBGLine(entryX, entryY, tileX, tileY, useWindow, cpu.Cartridge.IsCGB, lineNumber) {
					break
				}
			}
		}
	}
	cpu.execVBlank()
	screen := cpu.GPU.GetOriginal()

	file, err := os.Create(output)
	if err != nil {
		return err
	}
	defer file.Close()

	opt := jpeg.Options{
		Quality: 100,
	}
	if err = jpeg.Encode(file, screen, &opt); err != nil {
		return err
	}
	return nil
}
