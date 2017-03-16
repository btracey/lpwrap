package lpwrap

import (
	"bufio"
	"errors"
	"io"
	"strconv"
	"strings"
)

// Gurobi interfaces
type Gurobi struct{}

// WriteGurobi writes a gurobi file.
func (gur Gurobi) WriteGurobi(f io.Writer, lp LP) error {
	names, nameMap := IndexVariables(lp)

	// Temporary memory
	var b []byte

	// Write objective
	switch lp.Objective.OptKind {
	default:
		panic("lp: bad objective")
	case Maximize:
		f.Write([]byte("Maximize\n"))
	case Minimize:
		f.Write([]byte("Minimize\n"))
	}

	b = gur.objectiveBytes(b, lp, names, nameMap)
	if _, err := f.Write(b); err != nil {
		return err
	}
	f.Write([]byte("\n"))
	// Not sure if can use the offset term in the objective, probably not.

	f.Write([]byte("Subject To\n"))
	// Write constraints
	for _, c := range lp.Constraints {
		b = gur.constraintBytes(b, c, names, nameMap)
		f.Write(b)
	}
	return nil
}

func (gur Gurobi) objectiveBytes(b []byte, lp LP, names []string, nameMap map[string]int) []byte {
	b = b[:0]
	m, _ := CondenseTerms(lp.Objective.Terms, nameMap)
	b = append(b, '\t')
	b = gur.termsBytes(b, m)
	b = append(b, '\n')
	return b
}

func (gur Gurobi) constraintBytes(b []byte, c Constraint, names []string, nameMap map[string]int) []byte {
	b = b[:0]
	m, con := CondenseConstraint(c, nameMap)
	b = gur.termsBytes(b, m)

	var opstr string
	switch c.Comp {
	default:
		panic("lp: bad comp")
	case GE:
		opstr = " >= "
	case LE:
		opstr = " <= "
	case EQ:
		opstr = " = "
	}
	b = append(b, []byte(opstr)...)

	str := strconv.FormatFloat(con, 'g', 16, 64)
	b = append(b, []byte(str)...)
	b = append(b, []byte("\n")...)
	return b
}

func (gur Gurobi) termsBytes(b []byte, m map[string]float64) []byte {
	first := true
	for name, v := range m {
		if v == 0 {
			continue
		}
		if !first {
			b = append(b, []byte(" + ")...)
		} else {
			first = false
		}
		str := strconv.FormatFloat(v, 'g', 16, 64)
		b = append(b, []byte(str)...)
		b = append(b, []byte(" ")...)
		b = append(b, []byte(name)...)
	}
	return b
}

// ParseSol parses a solution file in Gurobi `.sol` format.
func (gur Gurobi) ParseSol(f io.Reader) (map[string]float64, error) {
	sol := make(map[string]float64)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		txt := scanner.Text()
		txt = strings.TrimSpace(txt)
		if len(txt) == 0 {
			continue
		}
		if txt[0] == '#' {
			continue
		}
		strs := strings.Split(txt, " ")
		if len(strs) != 2 {
			return nil, errors.New("lpwrap: unexpected file format")
		}
		v, err := strconv.ParseFloat(strs[1], 64)
		if err != nil {
			return nil, errors.New("lpwrap: bad float parse")
		}
		sol[strs[0]] = v
	}
	return sol, nil
}
