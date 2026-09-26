package motion

import (
	"fmt"
	"math"
)

// FormulaOp is a serializable scalar operation. Formula dimensions are the
// caller's editable viewport radii, independent of image or font metrics.
type FormulaOp string

const (
	FormulaConstant      FormulaOp = "constant"
	FormulaTime          FormulaOp = "time"
	FormulaSecondaryTime FormulaOp = "secondary_time"
	FormulaIndex         FormulaOp = "index"
	FormulaWidth         FormulaOp = "width"
	FormulaHeight        FormulaOp = "height"
	FormulaCount         FormulaOp = "count"
	FormulaAdd           FormulaOp = "add"
	FormulaSub           FormulaOp = "subtract"
	FormulaMul           FormulaOp = "multiply"
	FormulaDiv           FormulaOp = "divide"
	FormulaSin           FormulaOp = "sine"
	FormulaCos           FormulaOp = "cosine"
	FormulaFloor         FormulaOp = "floor"
)

// FormulaExpr is data, suitable for an editor or a Go preset. Args are kept
// in arithmetic order; the compiler never reassociates floating-point sums.
type FormulaExpr struct {
	Op    FormulaOp     `json:"op"`
	Value float64       `json:"value,omitempty"`
	Args  []FormulaExpr `json:"args,omitempty"`
}

func ExprConst(v float64) FormulaExpr { return FormulaExpr{Op: FormulaConstant, Value: v} }
func ExprTime() FormulaExpr           { return FormulaExpr{Op: FormulaTime} }
func ExprSecondaryTime() FormulaExpr  { return FormulaExpr{Op: FormulaSecondaryTime} }
func ExprIndex() FormulaExpr          { return FormulaExpr{Op: FormulaIndex} }
func ExprWidth() FormulaExpr          { return FormulaExpr{Op: FormulaWidth} }
func ExprHeight() FormulaExpr         { return FormulaExpr{Op: FormulaHeight} }
func ExprCount() FormulaExpr          { return FormulaExpr{Op: FormulaCount} }
func ExprAdd(a, b FormulaExpr) FormulaExpr {
	return FormulaExpr{Op: FormulaAdd, Args: []FormulaExpr{a, b}}
}
func ExprSub(a, b FormulaExpr) FormulaExpr {
	return FormulaExpr{Op: FormulaSub, Args: []FormulaExpr{a, b}}
}
func ExprMul(a, b FormulaExpr) FormulaExpr {
	return FormulaExpr{Op: FormulaMul, Args: []FormulaExpr{a, b}}
}
func ExprDiv(a, b FormulaExpr) FormulaExpr {
	return FormulaExpr{Op: FormulaDiv, Args: []FormulaExpr{a, b}}
}
func ExprSin(a FormulaExpr) FormulaExpr { return FormulaExpr{Op: FormulaSin, Args: []FormulaExpr{a}} }
func ExprCos(a FormulaExpr) FormulaExpr { return FormulaExpr{Op: FormulaCos, Args: []FormulaExpr{a}} }
func ExprFloor(a FormulaExpr) FormulaExpr {
	return FormulaExpr{Op: FormulaFloor, Args: []FormulaExpr{a}}
}

type formulaOpcode uint8

const (
	formulaConstant formulaOpcode = iota
	formulaTime
	formulaSecondaryTime
	formulaIndex
	formulaWidth
	formulaHeight
	formulaCount
	formulaAdd
	formulaSub
	formulaMul
	formulaDiv
	formulaSin
	formulaCos
	formulaFloor
	formulaTimeSin
	formulaTimeCos
	formulaLagSin
	formulaLagCos
)

type formulaInstruction struct {
	op            formulaOpcode
	value, value2 float64
}

func formulaTimeWave(expr FormulaExpr) (rate float64, ok bool) {
	if expr.Op != FormulaMul || expr.Value != 0 || len(expr.Args) != 2 || expr.Args[0].Op != FormulaTime || expr.Args[0].Value != 0 || len(expr.Args[0].Args) != 0 || expr.Args[1].Op != FormulaConstant || len(expr.Args[1].Args) != 0 {
		return 0, false
	}
	return expr.Args[1].Value, true
}

func formulaLagWave(expr FormulaExpr) (spacing, rate float64, ok bool) {
	if expr.Op != FormulaMul || expr.Value != 0 || len(expr.Args) != 2 || expr.Args[1].Op != FormulaConstant || len(expr.Args[1].Args) != 0 {
		return 0, 0, false
	}
	lag := expr.Args[0]
	if lag.Op != FormulaSub || lag.Value != 0 || len(lag.Args) != 2 || lag.Args[0].Op != FormulaTime || lag.Args[0].Value != 0 || len(lag.Args[0].Args) != 0 {
		return 0, 0, false
	}
	index := lag.Args[1]
	if index.Op != FormulaMul || index.Value != 0 || len(index.Args) != 2 || index.Args[0].Op != FormulaIndex || index.Args[0].Value != 0 || len(index.Args[0].Args) != 0 || index.Args[1].Op != FormulaConstant || len(index.Args[1].Args) != 0 {
		return 0, 0, false
	}
	return index.Args[1].Value, expr.Args[1].Value, true
}

// FormulaProgram is compiled once into a bounded postfix program. At has no
// heap allocation and uses a fixed local stack for mobile-friendly sampling.
type FormulaProgram struct{ code []formulaInstruction }

func CompileFormula(expr FormulaExpr) (*FormulaProgram, error) {
	program := &FormulaProgram{code: make([]formulaInstruction, 0, 64)}
	stack, maximum := 0, 0
	var compile func(FormulaExpr, int) error
	compile = func(node FormulaExpr, depth int) error {
		if depth > 64 || len(program.code) >= 512 || math.IsNaN(node.Value) || math.IsInf(node.Value, 0) {
			return fmt.Errorf("motion: invalid formula depth, length or value")
		}
		if (node.Op == FormulaSin || node.Op == FormulaCos) && len(node.Args) == 1 {
			if spacing, rate, ok := formulaLagWave(node.Args[0]); ok && !math.IsNaN(spacing) && !math.IsInf(spacing, 0) && !math.IsNaN(rate) && !math.IsInf(rate, 0) {
				opcode := formulaLagSin
				if node.Op == FormulaCos {
					opcode = formulaLagCos
				}
				program.code = append(program.code, formulaInstruction{op: opcode, value: spacing, value2: rate})
				stack++
				maximum = max(maximum, stack)
				if maximum > 64 {
					return fmt.Errorf("motion: formula stack exceeds limit")
				}
				return nil
			}
			if rate, ok := formulaTimeWave(node.Args[0]); ok && !math.IsNaN(rate) && !math.IsInf(rate, 0) {
				opcode := formulaTimeSin
				if node.Op == FormulaCos {
					opcode = formulaTimeCos
				}
				program.code = append(program.code, formulaInstruction{op: opcode, value: rate})
				stack++
				maximum = max(maximum, stack)
				if maximum > 64 {
					return fmt.Errorf("motion: formula stack exceeds limit")
				}
				return nil
			}
		}
		arity := 0
		var opcode formulaOpcode
		switch node.Op {
		case FormulaConstant:
			opcode = formulaConstant
		case FormulaTime:
			opcode = formulaTime
		case FormulaSecondaryTime:
			opcode = formulaSecondaryTime
		case FormulaIndex:
			opcode = formulaIndex
		case FormulaWidth:
			opcode = formulaWidth
		case FormulaHeight:
			opcode = formulaHeight
		case FormulaCount:
			opcode = formulaCount
		case FormulaSin:
			opcode = formulaSin
			arity = 1
		case FormulaCos:
			opcode = formulaCos
			arity = 1
		case FormulaFloor:
			opcode = formulaFloor
			arity = 1
		case FormulaAdd:
			opcode = formulaAdd
			arity = 2
		case FormulaSub:
			opcode = formulaSub
			arity = 2
		case FormulaMul:
			opcode = formulaMul
			arity = 2
		case FormulaDiv:
			opcode = formulaDiv
			arity = 2
		default:
			return fmt.Errorf("motion: unknown formula operation %q", node.Op)
		}
		if len(node.Args) != arity {
			return fmt.Errorf("motion: wrong formula operation arity")
		}
		for _, argument := range node.Args {
			if err := compile(argument, depth+1); err != nil {
				return err
			}
		}
		program.code = append(program.code, formulaInstruction{op: opcode, value: node.Value})
		stack += 1 - arity
		maximum = max(maximum, stack)
		if maximum > 64 {
			return fmt.Errorf("motion: formula stack exceeds limit")
		}
		return nil
	}
	if err := compile(expr, 0); err != nil {
		return nil, err
	}
	if stack != 1 {
		return nil, fmt.Errorf("motion: invalid formula stack")
	}
	return program, nil
}

// At evaluates a compiled formula in authored operation order. Nonfinite
// inputs and division by zero return zero rather than entering the renderer.
func (program *FormulaProgram) At(time, index, width, height, count float64) float64 {
	return program.AtWithSecondaryTime(time, 0, index, width, height, count)
}

// AtWithSecondaryTime supplies a second independent clock for coupled motion.
// Formulas without secondary_time behave exactly like At.
func (program *FormulaProgram) AtWithSecondaryTime(time, secondaryTime, index, width, height, count float64) float64 {
	if math.IsNaN(time) || math.IsInf(time, 0) || math.IsNaN(secondaryTime) || math.IsInf(secondaryTime, 0) || math.IsNaN(index) || math.IsInf(index, 0) ||
		math.IsNaN(width) || math.IsInf(width, 0) || math.IsNaN(height) || math.IsInf(height, 0) ||
		math.IsNaN(count) || math.IsInf(count, 0) {
		return 0
	}
	var stack [64]float64
	sp := 0
	for _, instruction := range program.code {
		switch instruction.op {
		case formulaConstant:
			stack[sp] = instruction.value
			sp++
		case formulaTime:
			stack[sp] = time
			sp++
		case formulaSecondaryTime:
			stack[sp] = secondaryTime
			sp++
		case formulaIndex:
			stack[sp] = index
			sp++
		case formulaWidth:
			stack[sp] = width
			sp++
		case formulaHeight:
			stack[sp] = height
			sp++
		case formulaCount:
			stack[sp] = count
			sp++
		case formulaAdd:
			stack[sp-2] += stack[sp-1]
			sp--
		case formulaSub:
			stack[sp-2] -= stack[sp-1]
			sp--
		case formulaMul:
			stack[sp-2] *= stack[sp-1]
			sp--
		case formulaDiv:
			if stack[sp-1] == 0 {
				return 0
			}
			stack[sp-2] /= stack[sp-1]
			sp--
		case formulaSin:
			stack[sp-1] = math.Sin(stack[sp-1])
		case formulaCos:
			stack[sp-1] = math.Cos(stack[sp-1])
		case formulaFloor:
			stack[sp-1] = math.Floor(stack[sp-1])
		case formulaTimeSin:
			stack[sp] = math.Sin(time * instruction.value)
			sp++
		case formulaTimeCos:
			stack[sp] = math.Cos(time * instruction.value)
			sp++
		case formulaLagSin:
			stack[sp] = math.Sin((time - index*instruction.value) * instruction.value2)
			sp++
		case formulaLagCos:
			stack[sp] = math.Cos((time - index*instruction.value) * instruction.value2)
			sp++
		}
	}
	if math.IsNaN(stack[0]) || math.IsInf(stack[0], 0) {
		return 0
	}
	return stack[0]
}

type FormulaFormationConfig struct{ X, Y FormulaExpr }

// FormulaFormation combines two compiled scalar programs into a reusable
// sprite, logo or glyph trajectory with independent caller-supplied geometry.
type FormulaFormation struct{ X, Y *FormulaProgram }

func NewFormulaFormation(config FormulaFormationConfig) (*FormulaFormation, error) {
	x, err := CompileFormula(config.X)
	if err != nil {
		return nil, err
	}
	y, err := CompileFormula(config.Y)
	if err != nil {
		return nil, err
	}
	return &FormulaFormation{X: x, Y: y}, nil
}

func (formation *FormulaFormation) At(time float64, index int, width, height float64, count int) Point {
	i, c := float64(index), float64(count)
	return Point{X: formation.X.At(time, i, width, height, c), Y: formation.Y.At(time, i, width, height, c)}
}
