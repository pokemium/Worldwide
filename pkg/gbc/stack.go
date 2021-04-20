package gbc

func (cpu *CPU) push(b byte) {
	cpu.SetMemory8(cpu.Reg.SP-1, b)
	cpu.Reg.SP--
}

func (cpu *CPU) pop() byte {
	value := cpu.FetchMemory8(cpu.Reg.SP)
	cpu.Reg.SP++
	return value
}

func (cpu *CPU) pushAF() {
	cpu.push(cpu.Reg.R[A])
	cpu.timer(1)
	cpu.push(cpu.Reg.R[F] & 0xf0)
}

func (cpu *CPU) popAF() {
	cpu.Reg.R[F] = cpu.pop() & 0xf0
	cpu.timer(1)
	cpu.Reg.R[A] = cpu.pop()
}

func (cpu *CPU) pushBC() {
	cpu.push(cpu.Reg.R[B])
	cpu.timer(1)
	cpu.push(cpu.Reg.R[C])
}

func (cpu *CPU) popBC() {
	cpu.Reg.R[C] = cpu.pop()
	cpu.timer(1)
	cpu.Reg.R[B] = cpu.pop()
}

func (cpu *CPU) pushDE() {
	cpu.push(cpu.Reg.R[D])
	cpu.timer(1)
	cpu.push(cpu.Reg.R[E])
}

func (cpu *CPU) popDE() {
	cpu.Reg.R[E] = cpu.pop()
	cpu.timer(1)
	cpu.Reg.R[D] = cpu.pop()
}

func (cpu *CPU) pushHL() {
	cpu.push(cpu.Reg.R[H])
	cpu.timer(1)
	cpu.push(cpu.Reg.R[L])
}

func (cpu *CPU) popHL() {
	cpu.Reg.R[L] = cpu.pop()
	cpu.timer(1)
	cpu.Reg.R[H] = cpu.pop()
}

func (cpu *CPU) pushPC() {
	upper, lower := byte(cpu.Reg.PC>>8), byte(cpu.Reg.PC)
	cpu.push(upper)
	cpu.push(lower)
}

func (cpu *CPU) pushPCCALL() {
	upper := byte(cpu.Reg.PC >> 8)
	cpu.push(upper)
	cpu.timer(1) // M = 4: PC push: memory access for high byte
	lower := byte(cpu.Reg.PC & 0x00ff)
	cpu.push(lower)
	cpu.timer(1) // M = 5: PC push: memory access for low byte
}

func (cpu *CPU) popPC() {
	lower := uint16(cpu.pop())
	upper := uint16(cpu.pop())
	cpu.Reg.PC = (upper << 8) | lower
}
