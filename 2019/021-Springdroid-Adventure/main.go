package main

import (
	"fmt"
)

const debug = false

const (
	OpcodeAdd                = 1
	OpcodeMultiply           = 2
	OpcodeGetInput           = 3
	OpcodeWriteOutput        = 4
	OpcodeJumpIfTrue         = 5
	OpcodeJumpIfFalse        = 6
	OpcodeLessThan           = 7
	OpcodeEquals             = 8
	OpcodeAdjustRelativeBase = 9
	OpcodeHalt               = 99
)

const (
	InputModePosition  = 0
	InputModeImmidiate = 1
	InputModeRelative  = 2
)

type Process struct {
	memory        []int
	position      int
	output        []int
	input         []int
	inputPointer  int
	outputPointer int
	relativeBase  int

	halted bool
}

func NewProcess(code []int, input []int) *Process {
	memory := make([]int, len(code)+900000)

	copy(memory, code)

	return &Process{
		memory:       memory,
		position:     0,
		input:        input,
		inputPointer: 0,
		halted:       false,
	}
}

//
// Read & write to program memory
//
func (p *Process) Read(position int) (int, error) {
	if position >= len(p.memory) || position < 0 {
		return 0, fmt.Errorf("Index %d out of range", position)
	}

	return p.memory[position], nil
}

func (p *Process) Write(position int, value int, mode int) error {
	if position >= len(p.memory) || position < 0 {
		return fmt.Errorf("Index out of range")
	}

	switch mode {
	case InputModePosition:
		if debug {
			fmt.Printf(" A %d = %d", p.memory[position], value)
		}
		p.memory[p.memory[position]] = value
	case InputModeRelative:
		if debug {
			fmt.Printf(" R %d = %d", p.memory[position]+p.relativeBase, value)
		}
		p.memory[p.memory[position]+p.relativeBase] = value
	}

	return nil
}

func (p *Process) DumpMemory() {
	for i, v := range p.memory {
		if i == p.position {
			fmt.Printf("[%d] ", v)
		} else {
			fmt.Printf("%d ", v)
		}
	}

	fmt.Print("\n")
}

func (p *Process) LoadParam(position int, mode int) (int, error) {
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

	case InputModeRelative:
		pointer, err := p.Read(position)
		if err != nil {
			return 0, err
		}

		value, err := p.Read(pointer + p.relativeBase)
		if err != nil {
			return 0, err
		}

		return value, nil
	default:
		return 0, fmt.Errorf("Unknonwn input mode")
	}
}

func (p *Process) Debug(name string, length int) {
	if !debug {
		return
	}

	fmt.Printf("%4s ", name)

	mode1 := (p.memory[p.position] / 100) % 10
	mode2 := (p.memory[p.position] / 1000) % 10

	if length >= 1 {
		fmt.Printf(" %10d ", p.memory[p.position+1])
	} else {
		fmt.Printf(" %9s- ", "")
	}

	if length >= 2 {
		fmt.Printf(" %10d ", p.memory[p.position+2])
	} else {
		fmt.Printf(" %9s- ", "")
	}

	if length >= 3 {
		fmt.Printf(" %10d ", p.memory[p.position+3])
	} else {
		fmt.Printf(" %9s- ", "")
	}

	fmt.Printf(" | ")

	if length >= 1 {
		v, _ := p.LoadParam(p.position+1, mode1)
		fmt.Printf(" %16d ", v)
	} else {
		fmt.Printf(" %15s- ", "")
	}

	if length >= 2 {
		v, _ := p.LoadParam(p.position+2, mode2)

		fmt.Printf(" %16d ", v)
	} else {
		fmt.Printf(" %15s- ", "")
	}

	if length >= 3 {
		v, _ := p.Read(p.position + 3)

		fmt.Printf(" %16d ", v)
	} else {
		fmt.Printf(" %15s- ", "")
	}
}

//
// Run program until halt or error.
//
func (p *Process) RunTilInterupt() error {
	operation, err := p.Read(p.position)
	if err != nil {
		return err
	}

	if debug {
		fmt.Println()
		fmt.Printf("%4d (r %4d): ", p.position, p.relativeBase)
	}

	instruction := operation % 100

	param1Mode := (operation / 100) % 10
	param2Mode := (operation / 1000) % 10
	param3Mode := (operation / 10000) % 10

	if debug {
		fmt.Printf("[%d %d %d %2d] ", param1Mode, param2Mode, param3Mode, instruction)
	}

	switch instruction {
	case OpcodeAdd:
		p.Debug("ADD", 3)

		value1, err := p.LoadParam(p.position+1, param1Mode)
		if err != nil {
			return err
		}

		value2, err := p.LoadParam(p.position+2, param2Mode)
		if err != nil {
			return err
		}

		p.Write(p.position+3, value1+value2, param3Mode)

		p.position += 4

		return p.RunTilInterupt()
	case OpcodeMultiply:
		p.Debug("MUL", 3)

		value1, err := p.LoadParam(p.position+1, param1Mode)
		if err != nil {
			return err
		}

		value2, err := p.LoadParam(p.position+2, param2Mode)
		if err != nil {
			return err
		}

		p.Write(p.position+3, value1*value2, param3Mode)

		p.position += 4

		return p.RunTilInterupt()
	case OpcodeGetInput:
		p.Debug("GET", 1)

		if p.inputPointer == len(p.input) {
			// no input, program needs to complete
			// fmt.Printf("Waiting for input")
			return nil
		}

		p.Write(p.position+1, p.input[p.inputPointer], param1Mode)

		p.inputPointer++

		p.position += 2

		return p.RunTilInterupt()
	case OpcodeWriteOutput:
		p.Debug("WRT", 1)

		value, err := p.LoadParam(p.position+1, param1Mode)
		if err != nil {
			return err
		}

		p.output = append(p.output, value)

		p.position += 2

		return p.RunTilInterupt()
	case OpcodeJumpIfTrue:
		p.Debug("JIT", 2)

		value1, err := p.LoadParam(p.position+1, param1Mode)
		if err != nil {
			return err
		}

		if value1 != 0 {
			value2, err := p.LoadParam(p.position+2, param2Mode)
			if err != nil {
				return err
			}

			if debug {
				fmt.Printf("true")
			}

			p.position = value2
		} else {
			if debug {
				fmt.Printf("false")
			}

			p.position += 3
		}
		return p.RunTilInterupt()

	case OpcodeJumpIfFalse:
		p.Debug("JIF", 2)

		value1, err := p.LoadParam(p.position+1, param1Mode)
		if err != nil {
			return err
		}

		if value1 == 0 {
			value2, err := p.LoadParam(p.position+2, param2Mode)
			if err != nil {
				return err
			}

			if debug {
				fmt.Printf("true")
			}

			p.position = value2
		} else {
			if debug {
				fmt.Printf("false ")
			}

			p.position += 3
		}

		return p.RunTilInterupt()
	case OpcodeLessThan:
		p.Debug("LT", 3)

		value1, err := p.LoadParam(p.position+1, param1Mode)
		if err != nil {
			return err
		}

		value2, err := p.LoadParam(p.position+2, param2Mode)
		if err != nil {
			return err
		}

		if value1 < value2 {
			if debug {
				fmt.Printf("true")
			}

			p.Write(p.position+3, 1, param3Mode)
		} else {
			if debug {
				fmt.Printf("false")
			}

			p.Write(p.position+3, 0, param3Mode)
		}

		p.position += 4

		return p.RunTilInterupt()
	case OpcodeEquals:
		p.Debug("EQL", 3)

		value1, err := p.LoadParam(p.position+1, param1Mode)
		if err != nil {
			return err
		}

		value2, err := p.LoadParam(p.position+2, param2Mode)
		if err != nil {
			return err
		}

		if value1 == value2 {
			if debug {
				fmt.Printf("true")
			}

			p.Write(p.position+3, 1, param3Mode)
		} else {
			if debug {
				fmt.Printf("false")
			}

			p.Write(p.position+3, 0, param3Mode)
		}

		p.position += 4

		return p.RunTilInterupt()

	case OpcodeAdjustRelativeBase:
		p.Debug("ADJ", 1)

		value, err := p.LoadParam(p.position+1, param1Mode)
		if err != nil {
			return err
		}

		p.relativeBase += value

		p.position += 2

		return p.RunTilInterupt()
	case OpcodeHalt:
		if debug {
			fmt.Println("halt")
		}

		p.halted = true

		return nil
	default:
		return fmt.Errorf("Unknonwn opcode %d", operation)
	}

	// fmt.Println(operation)
	// fmt.Println(instruction)
	panic("How we got here?")
}

func (p *Process) Run() error {
	for {
		err := p.RunTilInterupt()

		if err != nil {
			return err
		}

		if p.halted {
			return nil
		}
	}
}

func (p *Process) AddInput(val int) {
	p.input = append(p.input, val)
}

func (p *Process) NextOutput() int {
	res := p.output[p.outputPointer]
	p.outputPointer++
	return res
}

func (p *Process) inputString(str string) {
	fmt.Println("Entering: ", str)
	for _, b := range str {
		p.AddInput(int(b))
	}

	p.AddInput(int('\n'))
}

func (p *Process) readOutput() string {
	result := []byte{}

	for _, b := range p.output {
		result = append(result, byte(b))
	}

	return string(result)
}

func show() {
	values := []bool{true, false}

	for _, c1 := range values {
		for _, c2 := range values {
			for _, c3 := range values {
				for _, c4 := range values {
					fmt.Printf("%6t %6t %6t %6t\n", c1, c2, c3, c4)
				}
			}
		}
	}
}

func part1() {
	code := []int{109, 2050, 21101, 966, 0, 1, 21102, 13, 1, 0, 1105, 1, 1378, 21101, 20, 0, 0, 1106, 0, 1337, 21102, 27, 1, 0, 1105, 1, 1279, 1208, 1, 65, 748, 1005, 748, 73, 1208, 1, 79, 748, 1005, 748, 110, 1208, 1, 78, 748, 1005, 748, 132, 1208, 1, 87, 748, 1005, 748, 169, 1208, 1, 82, 748, 1005, 748, 239, 21101, 0, 1041, 1, 21101, 0, 73, 0, 1105, 1, 1421, 21102, 78, 1, 1, 21101, 0, 1041, 2, 21102, 1, 88, 0, 1106, 0, 1301, 21101, 68, 0, 1, 21102, 1041, 1, 2, 21101, 103, 0, 0, 1106, 0, 1301, 1102, 1, 1, 750, 1105, 1, 298, 21102, 1, 82, 1, 21102, 1, 1041, 2, 21102, 1, 125, 0, 1105, 1, 1301, 1101, 0, 2, 750, 1105, 1, 298, 21101, 79, 0, 1, 21102, 1, 1041, 2, 21101, 147, 0, 0, 1105, 1, 1301, 21102, 84, 1, 1, 21102, 1041, 1, 2, 21102, 1, 162, 0, 1105, 1, 1301, 1101, 0, 3, 750, 1106, 0, 298, 21102, 1, 65, 1, 21101, 0, 1041, 2, 21101, 184, 0, 0, 1106, 0, 1301, 21101, 76, 0, 1, 21101, 0, 1041, 2, 21102, 1, 199, 0, 1106, 0, 1301, 21102, 75, 1, 1, 21102, 1041, 1, 2, 21101, 0, 214, 0, 1106, 0, 1301, 21102, 1, 221, 0, 1106, 0, 1337, 21102, 1, 10, 1, 21101, 0, 1041, 2, 21101, 236, 0, 0, 1105, 1, 1301, 1106, 0, 553, 21102, 1, 85, 1, 21102, 1, 1041, 2, 21101, 0, 254, 0, 1106, 0, 1301, 21101, 0, 78, 1, 21102, 1, 1041, 2, 21102, 269, 1, 0, 1105, 1, 1301, 21101, 276, 0, 0, 1105, 1, 1337, 21101, 10, 0, 1, 21101, 1041, 0, 2, 21101, 291, 0, 0, 1105, 1, 1301, 1102, 1, 1, 755, 1106, 0, 553, 21102, 32, 1, 1, 21101, 0, 1041, 2, 21102, 313, 1, 0, 1105, 1, 1301, 21101, 320, 0, 0, 1105, 1, 1337, 21101, 327, 0, 0, 1105, 1, 1279, 2101, 0, 1, 749, 21102, 65, 1, 2, 21101, 0, 73, 3, 21101, 0, 346, 0, 1105, 1, 1889, 1206, 1, 367, 1007, 749, 69, 748, 1005, 748, 360, 1101, 1, 0, 756, 1001, 749, -64, 751, 1106, 0, 406, 1008, 749, 74, 748, 1006, 748, 381, 1102, 1, -1, 751, 1105, 1, 406, 1008, 749, 84, 748, 1006, 748, 395, 1101, -2, 0, 751, 1105, 1, 406, 21102, 1, 1100, 1, 21102, 1, 406, 0, 1106, 0, 1421, 21101, 32, 0, 1, 21102, 1100, 1, 2, 21101, 0, 421, 0, 1106, 0, 1301, 21101, 428, 0, 0, 1106, 0, 1337, 21101, 0, 435, 0, 1105, 1, 1279, 2102, 1, 1, 749, 1008, 749, 74, 748, 1006, 748, 453, 1101, 0, -1, 752, 1105, 1, 478, 1008, 749, 84, 748, 1006, 748, 467, 1101, 0, -2, 752, 1106, 0, 478, 21102, 1168, 1, 1, 21102, 478, 1, 0, 1105, 1, 1421, 21102, 1, 485, 0, 1106, 0, 1337, 21102, 10, 1, 1, 21101, 1168, 0, 2, 21102, 1, 500, 0, 1105, 1, 1301, 1007, 920, 15, 748, 1005, 748, 518, 21101, 0, 1209, 1, 21101, 0, 518, 0, 1105, 1, 1421, 1002, 920, 3, 529, 1001, 529, 921, 529, 101, 0, 750, 0, 1001, 529, 1, 537, 101, 0, 751, 0, 1001, 537, 1, 545, 102, 1, 752, 0, 1001, 920, 1, 920, 1106, 0, 13, 1005, 755, 577, 1006, 756, 570, 21102, 1, 1100, 1, 21101, 570, 0, 0, 1106, 0, 1421, 21102, 1, 987, 1, 1106, 0, 581, 21101, 1001, 0, 1, 21101, 0, 588, 0, 1105, 1, 1378, 1101, 0, 758, 594, 102, 1, 0, 753, 1006, 753, 654, 21002, 753, 1, 1, 21102, 1, 610, 0, 1105, 1, 667, 21101, 0, 0, 1, 21102, 621, 1, 0, 1106, 0, 1463, 1205, 1, 647, 21101, 0, 1015, 1, 21102, 635, 1, 0, 1105, 1, 1378, 21101, 0, 1, 1, 21101, 646, 0, 0, 1106, 0, 1463, 99, 1001, 594, 1, 594, 1105, 1, 592, 1006, 755, 664, 1101, 0, 0, 755, 1106, 0, 647, 4, 754, 99, 109, 2, 1101, 726, 0, 757, 21201, -1, 0, 1, 21101, 9, 0, 2, 21102, 1, 697, 3, 21101, 0, 692, 0, 1105, 1, 1913, 109, -2, 2106, 0, 0, 109, 2, 1001, 757, 0, 706, 1202, -1, 1, 0, 1001, 757, 1, 757, 109, -2, 2106, 0, 0, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 255, 63, 223, 95, 127, 191, 159, 0, 162, 121, 141, 62, 189, 227, 116, 182, 153, 247, 157, 124, 254, 50, 138, 77, 140, 39, 170, 71, 214, 111, 98, 173, 166, 228, 187, 172, 216, 230, 218, 174, 252, 243, 238, 253, 229, 204, 155, 94, 47, 200, 119, 102, 167, 60, 186, 117, 38, 76, 201, 177, 126, 199, 249, 55, 106, 53, 43, 163, 107, 232, 125, 86, 205, 190, 220, 251, 215, 237, 239, 46, 42, 219, 34, 178, 115, 139, 78, 114, 156, 203, 113, 51, 212, 188, 118, 61, 100, 87, 202, 152, 242, 56, 69, 136, 101, 248, 143, 168, 92, 35, 221, 85, 154, 198, 185, 57, 206, 110, 120, 58, 137, 59, 158, 241, 234, 196, 184, 123, 233, 171, 70, 183, 108, 93, 197, 84, 181, 235, 79, 109, 179, 222, 236, 68, 245, 244, 213, 49, 142, 103, 99, 217, 250, 226, 54, 207, 169, 231, 246, 175, 122, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 20, 73, 110, 112, 117, 116, 32, 105, 110, 115, 116, 114, 117, 99, 116, 105, 111, 110, 115, 58, 10, 13, 10, 87, 97, 108, 107, 105, 110, 103, 46, 46, 46, 10, 10, 13, 10, 82, 117, 110, 110, 105, 110, 103, 46, 46, 46, 10, 10, 25, 10, 68, 105, 100, 110, 39, 116, 32, 109, 97, 107, 101, 32, 105, 116, 32, 97, 99, 114, 111, 115, 115, 58, 10, 10, 58, 73, 110, 118, 97, 108, 105, 100, 32, 111, 112, 101, 114, 97, 116, 105, 111, 110, 59, 32, 101, 120, 112, 101, 99, 116, 101, 100, 32, 115, 111, 109, 101, 116, 104, 105, 110, 103, 32, 108, 105, 107, 101, 32, 65, 78, 68, 44, 32, 79, 82, 44, 32, 111, 114, 32, 78, 79, 84, 67, 73, 110, 118, 97, 108, 105, 100, 32, 102, 105, 114, 115, 116, 32, 97, 114, 103, 117, 109, 101, 110, 116, 59, 32, 101, 120, 112, 101, 99, 116, 101, 100, 32, 115, 111, 109, 101, 116, 104, 105, 110, 103, 32, 108, 105, 107, 101, 32, 65, 44, 32, 66, 44, 32, 67, 44, 32, 68, 44, 32, 74, 44, 32, 111, 114, 32, 84, 40, 73, 110, 118, 97, 108, 105, 100, 32, 115, 101, 99, 111, 110, 100, 32, 97, 114, 103, 117, 109, 101, 110, 116, 59, 32, 101, 120, 112, 101, 99, 116, 101, 100, 32, 74, 32, 111, 114, 32, 84, 52, 79, 117, 116, 32, 111, 102, 32, 109, 101, 109, 111, 114, 121, 59, 32, 97, 116, 32, 109, 111, 115, 116, 32, 49, 53, 32, 105, 110, 115, 116, 114, 117, 99, 116, 105, 111, 110, 115, 32, 99, 97, 110, 32, 98, 101, 32, 115, 116, 111, 114, 101, 100, 0, 109, 1, 1005, 1262, 1270, 3, 1262, 21001, 1262, 0, 0, 109, -1, 2106, 0, 0, 109, 1, 21101, 0, 1288, 0, 1106, 0, 1263, 21001, 1262, 0, 0, 1101, 0, 0, 1262, 109, -1, 2106, 0, 0, 109, 5, 21102, 1310, 1, 0, 1106, 0, 1279, 21202, 1, 1, -2, 22208, -2, -4, -1, 1205, -1, 1332, 22102, 1, -3, 1, 21101, 0, 1332, 0, 1106, 0, 1421, 109, -5, 2105, 1, 0, 109, 2, 21101, 1346, 0, 0, 1105, 1, 1263, 21208, 1, 32, -1, 1205, -1, 1363, 21208, 1, 9, -1, 1205, -1, 1363, 1106, 0, 1373, 21102, 1370, 1, 0, 1105, 1, 1279, 1105, 1, 1339, 109, -2, 2105, 1, 0, 109, 5, 2101, 0, -4, 1385, 21001, 0, 0, -2, 22101, 1, -4, -4, 21101, 0, 0, -3, 22208, -3, -2, -1, 1205, -1, 1416, 2201, -4, -3, 1408, 4, 0, 21201, -3, 1, -3, 1105, 1, 1396, 109, -5, 2106, 0, 0, 109, 2, 104, 10, 21201, -1, 0, 1, 21102, 1436, 1, 0, 1106, 0, 1378, 104, 10, 99, 109, -2, 2105, 1, 0, 109, 3, 20002, 594, 753, -1, 22202, -1, -2, -1, 201, -1, 754, 754, 109, -3, 2106, 0, 0, 109, 10, 21101, 5, 0, -5, 21102, 1, 1, -4, 21101, 0, 0, -3, 1206, -9, 1555, 21102, 3, 1, -6, 21101, 0, 5, -7, 22208, -7, -5, -8, 1206, -8, 1507, 22208, -6, -4, -8, 1206, -8, 1507, 104, 64, 1105, 1, 1529, 1205, -6, 1527, 1201, -7, 716, 1515, 21002, 0, -11, -8, 21201, -8, 46, -8, 204, -8, 1106, 0, 1529, 104, 46, 21201, -7, 1, -7, 21207, -7, 22, -8, 1205, -8, 1488, 104, 10, 21201, -6, -1, -6, 21207, -6, 0, -8, 1206, -8, 1484, 104, 10, 21207, -4, 1, -8, 1206, -8, 1569, 21101, 0, 0, -9, 1106, 0, 1689, 21208, -5, 21, -8, 1206, -8, 1583, 21101, 1, 0, -9, 1105, 1, 1689, 1201, -5, 716, 1589, 20101, 0, 0, -2, 21208, -4, 1, -1, 22202, -2, -1, -1, 1205, -2, 1613, 21202, -5, 1, 1, 21101, 1613, 0, 0, 1106, 0, 1444, 1206, -1, 1634, 21202, -5, 1, 1, 21102, 1627, 1, 0, 1106, 0, 1694, 1206, 1, 1634, 21102, 1, 2, -3, 22107, 1, -4, -8, 22201, -1, -8, -8, 1206, -8, 1649, 21201, -5, 1, -5, 1206, -3, 1663, 21201, -3, -1, -3, 21201, -4, 1, -4, 1106, 0, 1667, 21201, -4, -1, -4, 21208, -4, 0, -1, 1201, -5, 716, 1676, 22002, 0, -1, -1, 1206, -1, 1686, 21102, 1, 1, -4, 1106, 0, 1477, 109, -10, 2105, 1, 0, 109, 11, 21102, 0, 1, -6, 21102, 0, 1, -8, 21102, 1, 0, -7, 20208, -6, 920, -9, 1205, -9, 1880, 21202, -6, 3, -9, 1201, -9, 921, 1724, 21002, 0, 1, -5, 1001, 1724, 1, 1733, 20101, 0, 0, -4, 21201, -4, 0, 1, 21101, 0, 1, 2, 21102, 1, 9, 3, 21102, 1, 1754, 0, 1106, 0, 1889, 1206, 1, 1772, 2201, -10, -4, 1767, 1001, 1767, 716, 1767, 20102, 1, 0, -3, 1106, 0, 1790, 21208, -4, -1, -9, 1206, -9, 1786, 22102, 1, -8, -3, 1106, 0, 1790, 22102, 1, -7, -3, 1001, 1733, 1, 1796, 20102, 1, 0, -2, 21208, -2, -1, -9, 1206, -9, 1812, 21201, -8, 0, -1, 1105, 1, 1816, 21201, -7, 0, -1, 21208, -5, 1, -9, 1205, -9, 1837, 21208, -5, 2, -9, 1205, -9, 1844, 21208, -3, 0, -1, 1105, 1, 1855, 22202, -3, -1, -1, 1106, 0, 1855, 22201, -3, -1, -1, 22107, 0, -1, -1, 1106, 0, 1855, 21208, -2, -1, -9, 1206, -9, 1869, 22102, 1, -1, -8, 1105, 1, 1873, 22102, 1, -1, -7, 21201, -6, 1, -6, 1105, 1, 1708, 22101, 0, -8, -10, 109, -11, 2105, 1, 0, 109, 7, 22207, -6, -5, -3, 22207, -4, -6, -2, 22201, -3, -2, -1, 21208, -1, 0, -6, 109, -7, 2105, 1, 0, 0, 109, 5, 2102, 1, -2, 1912, 21207, -4, 0, -1, 1206, -1, 1930, 21102, 1, 0, -4, 22102, 1, -4, 1, 21201, -3, 0, 2, 21102, 1, 1, 3, 21102, 1, 1949, 0, 1105, 1, 1954, 109, -5, 2106, 0, 0, 109, 6, 21207, -4, 1, -1, 1206, -1, 1977, 22207, -5, -3, -1, 1206, -1, 1977, 22101, 0, -5, -5, 1106, 0, 2045, 21202, -5, 1, 1, 21201, -4, -1, 2, 21202, -3, 2, 3, 21102, 1996, 1, 0, 1105, 1, 1954, 22101, 0, 1, -5, 21101, 0, 1, -2, 22207, -5, -3, -1, 1206, -1, 2015, 21101, 0, 0, -2, 22202, -3, -2, -3, 22107, 0, -4, -1, 1206, -1, 2037, 21201, -2, 0, 1, 21101, 2037, 0, 0, 105, 1, 1912, 21202, -3, -1, -3, 22201, -5, -3, -5, 109, -6, 2105, 1, 0}
	p := NewProcess(code, []int{})

	// if A and D are not a whole, but B is
	p.inputString("NOT C T")
	p.inputString("AND A T")
	p.inputString("AND D T")
	p.inputString("OR T J")

	p.inputString("NOT A T")
	p.inputString("OR T J")

	p.inputString("WALK")

	p.RunTilInterupt()

	fmt.Println(p.readOutput())
	fmt.Println(p.output[len(p.output)-1])
}

func main() {
	part1()
}