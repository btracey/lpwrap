package lpwrap_test

import (
	"fmt"

	"github.com/btracey/lpwrap"
)

func ExampleLP_Gonum() {
	// Solve the optimization problem
	// minimize_{a,b,c} 5*a + 3*c + 6
	//   s.t. b >= 3
	//        b + c = 10
	//        a >= 2*b
	//        3*c + 5 >= a
	//        c <= 9

	obj := lpwrap.Objective{
		[]lpwrap.Term{
			{Var: "a", Value: 5},
			{Var: "c", Value: 3},
			{Var: lpwrap.Constant, Value: 6},
		},
		lpwrap.Minimize,
	}

	c1 := lpwrap.Constraint{
		Left:  []lpwrap.Term{{"b", 1}},
		Comp:  lpwrap.GE,
		Right: []lpwrap.Term{{lpwrap.Constant, 3}},
	}
	c2 := lpwrap.Constraint{
		Left:  []lpwrap.Term{{"b", 1}, {"c", 1}},
		Comp:  lpwrap.EQ,
		Right: []lpwrap.Term{{lpwrap.Constant, 10}},
	}
	c3 := lpwrap.Constraint{
		Left:  []lpwrap.Term{{"a", 1}},
		Comp:  lpwrap.GE,
		Right: []lpwrap.Term{{"b", 2}},
	}
	c4 := lpwrap.Constraint{
		Left:  []lpwrap.Term{{"c", 3}, {lpwrap.Constant, 5}},
		Comp:  lpwrap.GE,
		Right: []lpwrap.Term{{"a", 1}},
	}
	c5 := lpwrap.Constraint{
		Left:  []lpwrap.Term{{"c", 1}},
		Comp:  lpwrap.LE,
		Right: []lpwrap.Term{{lpwrap.Constant, 9}},
	}
	constraints := []lpwrap.Constraint{c1, c2, c3, c4, c5}

	prob := lpwrap.LP{obj, constraints}

	result, err := lpwrap.Gonum{}.Solve(prob)
	if err != nil {
		fmt.Println(err)
	}
	ord := result.Ordered()
	fmt.Println("Optimal variable values")
	for i := range ord {
		fmt.Println(ord[i].Var, "=", ord[i].Value)
	}
	// Output:
	// Optimal variable values
	// a = 6
	// b = 3
	// c = 7
}
