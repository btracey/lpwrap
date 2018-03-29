package lpwrap

import (
	"gonum.org/v1/gonum/mat"
	"gonum.org/v1/gonum/optimize/convex/lp"
)

// GonumData represents a problem in a format useable by gonum/optimize/convex/lp.
//
//  minimize c^T * x
//  s.t      G * x <= h
//           A * x = b
type GonumData struct {
	C       []float64
	G       mat.Matrix
	H       []float64
	A       mat.Matrix
	B       []float64
	Offset  float64 // term left off objective
	Names   []string
	NameMap map[string]int
}

// Gonum is a wrapper for a solver using Gonum.
type Gonum struct{}

// Gonum converts the problem into Gonum format.
func (gonum Gonum) ConvertGonum(lp LP) GonumData {
	names, nameMap := IndexVariables(lp)
	nVar := len(names)
	c := make([]float64, nVar)
	var offset float64
	for _, term := range lp.Objective.Terms {
		if term.Var == Constant {
			offset = term.Value
			continue // constant value doesn't affect optimal. TODO(btracey): vet check?
		}
		idx, ok := nameMap[term.Var]
		if !ok {
			panic("term not present")
		}
		c[idx] += term.Value
	}
	switch lp.Objective.OptKind {
	default:
		panic("lpwrap: unknown optkind")
	case Minimize:
	case Maximize:
		for i := range c {
			c[i] *= -1
		}
	}

	var eqs, ineqs []Constraint
	for _, con := range lp.Constraints {
		switch con.Comp {
		default:
			panic("unknown op")
		case EQ:
			eqs = append(eqs, con)
		case LE, GE:
			ineqs = append(ineqs, con)
		}
	}

	a, b := gonum.constraintsToMatrix(eqs, nameMap)
	g, h := gonum.constraintsToMatrix(ineqs, nameMap)

	return GonumData{
		C:       c,
		A:       a,
		B:       b,
		G:       g,
		H:       h,
		Offset:  offset,
		Names:   names,
		NameMap: nameMap,
	}
}

func (g Gonum) constraintsToMatrix(cons []Constraint, nameMap map[string]int) (*mat.Dense, []float64) {
	nCon := len(cons)
	nVar := len(nameMap)
	a := mat.NewDense(nCon, nVar, nil)
	b := make([]float64, nCon)

	row := make([]float64, nVar)
	for i, con := range cons {
		var constant float64 // the constant term
		for _, term := range con.Left {
			if term.Var == Constant {
				constant -= term.Value // will be moved to the right
				continue
			}
			idx, ok := nameMap[term.Var]
			if !ok {
				panic("name not present")
			}
			row[idx] += term.Value
		}
		for _, term := range con.Right {
			if term.Var == Constant {
				constant += term.Value
				continue
			}
			idx, ok := nameMap[term.Var]
			if !ok {
				panic("name not present")
			}
			row[idx] -= term.Value // moved to the left
		}
		if con.Comp == GE {
			// Multiply the whole term by -1 to make it <=
			for i := range row {
				row[i] *= -1
			}
			constant *= -1
		}
		a.SetRow(i, row)
		b[i] = constant
		for i := range row {
			row[i] = 0
		}
	}
	return a, b
}

func (gonum Gonum) Solve(prob LP) (*Result, error) {
	gnm := gonum.ConvertGonum(prob)

	c, A, b := lp.Convert(gnm.C, gnm.G, gnm.H, gnm.A, gnm.B)

	optF, optX, err := lp.Simplex(c, A, b, 1e-6, nil)
	if err != nil {
		return nil, err
	}

	if prob.Objective.OptKind == Maximize {
		optF *= -1
	}
	optF += gnm.Offset

	varMap := make(map[string]float64)
	for i, name := range gnm.Names {
		varMap[name] = optX[i]
	}
	return &Result{optF, varMap}, nil
}
