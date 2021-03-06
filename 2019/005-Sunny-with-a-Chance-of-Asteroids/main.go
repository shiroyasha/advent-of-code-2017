package main

import (
	"fmt"
)

const (
	OpcodeAdd         = 1
	OpcodeMultiply    = 2
	OpcodeGetInput    = 3
	OpcodeWriteOutput = 4
	OpcodeJumpIfTrue  = 5
	OpcodeJumpIfFalse = 6
	OpcodeLessThan    = 7
	OpcodeEquals      = 8
	OpcodeHalt        = 99
)

const (
	InputModePosition  = 0
	InputModeImmidiate = 1
)

type Output struct {
	values []int
}

type Program struct {
	memory   []int
	position int
}

func NewProgram(code []int) *Program {
	memory := make([]int, len(code))

	copy(memory, code)

	return &Program{memory: memory, position: 0}
}

//
// Read & write to program memory
//
func (p *Program) Read(position int) (int, error) {
	if position >= len(p.memory) || position < 0 {
		return 0, fmt.Errorf("Index %d out of range", position)
	}

	return p.memory[position], nil
}

func (p *Program) Write(position int, value int) error {
	if position >= len(p.memory) || position < 0 {
		return fmt.Errorf("Index out of range")
	}

	p.memory[position] = value

	return nil
}

func (p *Program) DumpMemory() {
	for i, v := range p.memory {
		if i == p.position {
			fmt.Printf("[%d] ", v)
		} else {
			fmt.Printf("%d ", v)
		}
	}

	fmt.Print("\n")
}

func (p *Program) LoadParam(position int, mode int) (int, error) {
	switch mode {
	case InputModePosition:
		pointer, err := p.Read(position)
		if err != nil {
			return 0, err
		}

		value, err := p.Read(pointer)
		if err != nil {
			return 0, err
		}

		return value, nil
	case InputModeImmidiate:
		value, err := p.Read(position)
		if err != nil {
			return 0, err
		}

		return value, nil
	default:
		return 0, fmt.Errorf("Unknonwn input mode")
	}
}

//
// Run program until halt or error.
//
func (p *Program) Run(input int, output *Output) error {
	operation, err := p.Read(p.position)
	if err != nil {
		return err
	}

	fmt.Printf("%d: ", p.position)

	instruction := operation % 100

	param1Mode := (operation / 100) % 10
	param2Mode := (operation / 1000) % 10

	switch instruction {
	case OpcodeAdd:
		value1, err := p.LoadParam(p.position+1, param1Mode)
		if err != nil {
			return err
		}

		value2, err := p.LoadParam(p.position+2, param2Mode)
		if err != nil {
			return err
		}

		resultPointer, err := p.Read(p.position + 3)
		if err != nil {
			return err
		}

		p.Write(resultPointer, value1+value2)

		fmt.Printf("ADD %d %d %d ", value1, value2, resultPointer)

		p.position += 4
	case OpcodeMultiply:
		value1, err := p.LoadParam(p.position+1, param1Mode)
		if err != nil {
			return err
		}

		value2, err := p.LoadParam(p.position+2, param2Mode)
		if err != nil {
			return err
		}

		resultPointer, err := p.Read(p.position + 3)
		if err != nil {
			return err
		}

		p.Write(resultPointer, value1*value2)

		fmt.Printf("MUL %d %d %d ", value1, value2, resultPointer)

		p.position += 4
	case OpcodeGetInput:
		pointer, err := p.Read(p.position + 1)
		if err != nil {
			return err
		}

		p.Write(pointer, input)

		fmt.Printf("GET %d ", pointer)

		p.position += 2

	case OpcodeWriteOutput:
		value, err := p.LoadParam(p.position+1, param1Mode)
		if err != nil {
			return err
		}

		output.values = append(output.values, value)

		fmt.Printf("WRT %d ", value)

		p.position += 2

	case OpcodeJumpIfTrue:
		value1, err := p.LoadParam(p.position+1, param1Mode)
		if err != nil {
			return err
		}

		fmt.Printf("JMPT %d ", value1)

		if value1 != 0 {
			value2, err := p.LoadParam(p.position+2, param2Mode)
			if err != nil {
				return err
			}

			fmt.Printf("%d true", value2)

			p.position = value2
		} else {
			fmt.Printf("false")

			p.position += 3
		}

	case OpcodeJumpIfFalse:
		value1, err := p.LoadParam(p.position+1, param1Mode)
		if err != nil {
			return err
		}

		fmt.Printf("JMPF %d ", value1)

		if value1 == 0 {
			value2, err := p.LoadParam(p.position+2, param2Mode)
			if err != nil {
				return err
			}

			fmt.Printf("%d true", value2)

			p.position = value2
		} else {
			fmt.Printf("false ")

			p.position += 3
		}

	case OpcodeLessThan:
		value1, err := p.LoadParam(p.position+1, param1Mode)
		if err != nil {
			return err
		}

		value2, err := p.LoadParam(p.position+2, param2Mode)
		if err != nil {
			return err
		}

		pos, err := p.Read(p.position + 3)
		if err != nil {
			return err
		}

		fmt.Printf("LT %d %d %d ", value1, value2, pos)

		if value1 < value2 {
			fmt.Printf("true")

			p.Write(pos, 1)
		} else {
			fmt.Printf("false")

			p.Write(pos, 0)
		}

		p.position += 4

	case OpcodeEquals:
		value1, err := p.LoadParam(p.position+1, param1Mode)
		if err != nil {
			return err
		}

		value2, err := p.LoadParam(p.position+2, param2Mode)
		if err != nil {
			return err
		}

		pos, err := p.Read(p.position + 3)
		if err != nil {
			return err
		}

		fmt.Printf("EQ %d %d %d ", value1, value2, pos)

		if value1 == value2 {
			fmt.Printf("true")

			p.Write(pos, 1)
		} else {
			fmt.Printf("false")

			p.Write(pos, 0)
		}

		p.position += 4

	case OpcodeHalt:
		fmt.Printf("halt")

		return nil

	default:
		return fmt.Errorf("Unknonwn opcode %d", operation)
	}

	fmt.Printf("\n")

	return p.Run(input, output)
}

func main() {
	code := []int{
		3, 225, 1, 225, 6, 6, 1100, 1, 238, 225, 104, 0, 1101, 40, 27, 224, 101, -67, 224, 224, 4, 224, 1002, 223, 8, 223, 1001, 224, 2, 224, 1, 224, 223, 223, 1101, 33, 38, 225, 1102, 84, 60, 225, 1101, 65, 62, 225, 1002, 36, 13, 224, 1001, 224, -494, 224, 4, 224, 1002, 223, 8, 223, 1001, 224, 3, 224, 1, 223, 224, 223, 1102, 86, 5, 224, 101, -430, 224, 224, 4, 224, 1002, 223, 8, 223, 101, 6, 224, 224, 1, 223, 224, 223, 1102, 23, 50, 225, 1001, 44, 10, 224, 101, -72, 224, 224, 4, 224, 102, 8, 223, 223, 101, 1, 224, 224, 1, 224, 223, 223, 102, 47, 217, 224, 1001, 224, -2303, 224, 4, 224, 102, 8, 223, 223, 101, 2, 224, 224, 1, 223, 224, 223, 1102, 71, 84, 225, 101, 91, 40, 224, 1001, 224, -151, 224, 4, 224, 1002, 223, 8, 223, 1001, 224, 5, 224, 1, 223, 224, 223, 1101, 87, 91, 225, 1102, 71, 19, 225, 1, 92, 140, 224, 101, -134, 224, 224, 4, 224, 1002, 223, 8, 223, 101, 1, 224, 224, 1, 224, 223, 223, 2, 170, 165, 224, 1001, 224, -1653, 224, 4, 224, 1002, 223, 8, 223, 101, 5, 224, 224, 1, 223, 224, 223, 1101, 49, 32, 225, 4, 223, 99, 0, 0, 0, 677, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1105, 0, 99999, 1105, 227, 247, 1105, 1, 99999, 1005, 227, 99999, 1005, 0, 256, 1105, 1, 99999, 1106, 227, 99999, 1106, 0, 265, 1105, 1, 99999, 1006, 0, 99999, 1006, 227, 274, 1105, 1, 99999, 1105, 1, 280, 1105, 1, 99999, 1, 225, 225, 225, 1101, 294, 0, 0, 105, 1, 0, 1105, 1, 99999, 1106, 0, 300, 1105, 1, 99999, 1, 225, 225, 225, 1101, 314, 0, 0, 106, 0, 0, 1105, 1, 99999, 1107, 226, 677, 224, 1002, 223, 2, 223, 1006, 224, 329, 101, 1, 223, 223, 8, 226, 226, 224, 1002, 223, 2, 223, 1005, 224, 344, 101, 1, 223, 223, 1007, 677, 226, 224, 102, 2, 223, 223, 1005, 224, 359, 101, 1, 223, 223, 8, 226, 677, 224, 102, 2, 223, 223, 1005, 224, 374, 101, 1, 223, 223, 1107, 677, 677, 224, 1002, 223, 2, 223, 1005, 224, 389, 1001, 223, 1, 223, 108, 226, 677, 224, 102, 2, 223, 223, 1005, 224, 404, 1001, 223, 1, 223, 108, 677, 677, 224, 1002, 223, 2, 223, 1006, 224, 419, 101, 1, 223, 223, 107, 677, 677, 224, 102, 2, 223, 223, 1006, 224, 434, 101, 1, 223, 223, 108, 226, 226, 224, 1002, 223, 2, 223, 1006, 224, 449, 1001, 223, 1, 223, 8, 677, 226, 224, 1002, 223, 2, 223, 1005, 224, 464, 101, 1, 223, 223, 1108, 226, 677, 224, 1002, 223, 2, 223, 1006, 224, 479, 1001, 223, 1, 223, 1108, 677, 677, 224, 1002, 223, 2, 223, 1005, 224, 494, 101, 1, 223, 223, 7, 677, 677, 224, 1002, 223, 2, 223, 1005, 224, 509, 101, 1, 223, 223, 1007, 677, 677, 224, 1002, 223, 2, 223, 1005, 224, 524, 101, 1, 223, 223, 7, 677, 226, 224, 1002, 223, 2, 223, 1005, 224, 539, 101, 1, 223, 223, 1107, 677, 226, 224, 102, 2, 223, 223, 1006, 224, 554, 101, 1, 223, 223, 107, 226, 677, 224, 1002, 223, 2, 223, 1005, 224, 569, 101, 1, 223, 223, 107, 226, 226, 224, 1002, 223, 2, 223, 1005, 224, 584, 101, 1, 223, 223, 1108, 677, 226, 224, 102, 2, 223, 223, 1006, 224, 599, 1001, 223, 1, 223, 1008, 677, 677, 224, 102, 2, 223, 223, 1006, 224, 614, 101, 1, 223, 223, 7, 226, 677, 224, 102, 2, 223, 223, 1005, 224, 629, 101, 1, 223, 223, 1008, 226, 677, 224, 1002, 223, 2, 223, 1006, 224, 644, 101, 1, 223, 223, 1007, 226, 226, 224, 1002, 223, 2, 223, 1005, 224, 659, 1001, 223, 1, 223, 1008, 226, 226, 224, 102, 2, 223, 223, 1006, 224, 674, 1001, 223, 1, 223, 4, 223, 99, 226,
	}
	input := 5

	// code := []int{3, 9, 8, 9, 10, 9, 4, 9, 99, -1, 8}
	// input := 4

	// code := []int{3, 9, 7, 9, 10, 9, 4, 9, 99, -1, 8}
	// input := 4

	// code := []int{3, 3, 1108, -1, 8, 3, 4, 3, 99}
	// input := 12

	// code := []int{3, 3, 1107, -1, 8, 3, 4, 3, 99}
	// input := 1

	p := NewProgram(code)
	output := &Output{}

	err := p.Run(input, output)

	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(output)
}
