package lpwrap

import "sort"

type CompKind string

const (
	EQ CompKind = "EQ"
	LE          = "LE"
	GE          = "GE"
)

type OptKind string

const (
	Maximize OptKind = "Maximize"
	Minimize         = "Minimize"
)

// Constant represents a constant term.
const Constant = "__constant"

type Solver interface {
	Solve(lp LP) (Result, error)
}

// LP represents a linear program.
type LP struct {
	Objective   Objective
	Constraints []Constraint
}

// indexVariables assigns each variable to an index.
func indexVariables(lp LP) ([]string, map[string]int) {
	var names []string
	nameMap := make(map[string]int)
	for _, term := range lp.Objective.Terms {
		names, nameMap = addNameIfNew(term.Var, names, nameMap)
	}

	for _, con := range lp.Constraints {
		for _, term := range con.Left {
			names, nameMap = addNameIfNew(term.Var, names, nameMap)
		}
		for _, term := range con.Right {
			names, nameMap = addNameIfNew(term.Var, names, nameMap)
		}
	}
	return names, nameMap
}

// addNameIfNew adds the name
func addNameIfNew(newName string, names []string, nameMap map[string]int) ([]string, map[string]int) {
	if newName == Constant {
		return names, nameMap
	}
	_, ok := nameMap[newName]
	if !ok {
		idx := len(names)
		names = append(names, newName)
		nameMap[newName] = idx
	}
	return names, nameMap
}

type Term struct {
	Var   string
	Value float64
}

type Constraint struct {
	Left  []Term
	Comp  CompKind
	Right []Term
}

type Objective struct {
	Terms   []Term
	OptKind OptKind
}

type Result struct {
	Value  float64
	VarMap map[string]float64
}

// Ordered returns the variables ordered alphabetically.
func (r Result) Ordered() []Term {
	var terms []Term
	for key, val := range r.VarMap {
		terms = append(terms, Term{key, val})
	}
	sort.Sort(termSorter(terms))
	return terms
}

type termSorter []Term

func (t termSorter) Len() int {
	return len(t)
}

func (t termSorter) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}

func (t termSorter) Less(i, j int) bool {
	return t[i].Var < t[j].Var
}
