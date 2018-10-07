// DBDeployer - The MySQL Sandbox
// Copyright © 2006-2018 Giuseppe Maxia
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package common

import (
	"fmt"
	"testing"
)

func okEqualInt(label string, a, b int, t *testing.T) {
	if a == b {
		t.Logf("ok - %s: expected: %d\n", label, a)
	} else {
		t.Logf("not ok - %s: Numbers are not equal - expected %d, but got %d", label, b, a)
		t.Fail()
	}
}

func okEqualString(label, a, b string, t *testing.T) {
	if a == b {
		t.Logf("ok - %s: expected: '%s'\n", label, a)
	} else {
		t.Logf("not ok - %s: Strings are not equal - expected '%s', but got '%s'", label, b, a)
		t.Fail()
	}
}

func okEqualBool(label string, a, b bool, t *testing.T) {
	if a == b {
		t.Logf("ok - %s: expected: %v\n", label, a)
	} else {
		t.Logf("not ok - %s: Values are not the same - expected %v, but got %v", label, b, a)
		t.Fail()
	}
}

func TestStack(t *testing.T) {

	type compound struct {
		i int
		s string
		b bool
	}

	empty_size := 0
	small_size := 1
	long_size := 10000
	first_value := 0
	last_value := long_size - 1
	int_value := 1234
	stack := NewStack()
	stack.Push(int_value)
	okEqualInt("size full", stack.Len(), small_size, t)
	x := stack.Pop().(int)
	okEqualInt("value x", x, int_value, t)
	okEqualInt("size empty", stack.Len(), empty_size, t)

	for N := 0; N < long_size; N++ {
		stack.Push(N) // last one is 99
	}
	okEqualInt("size full", stack.Len(), long_size, t)
	x = stack.Top().(int)
	okEqualInt("value x (Top)", x, last_value, t)
	x = stack.Bottom().(int)
	okEqualInt("value x (Bottom)", x, first_value, t)
	x = stack.Pop().(int)
	okEqualInt("value x (Pop)", x, last_value, t)
	okEqualInt("size after pop", stack.Len(), last_value, t)
	for stack.Len() > 0 {
		stack.Pop()
	}
	okEqualInt("size empty", stack.Len(), empty_size, t)
	for N := 0; N < long_size; N++ {
		odd := true
		if N%2 == 0 {
			odd = false
		}
		c := compound{N, fmt.Sprintf("str%d", N), odd}
		stack.Push(c)
	}
	okEqualInt("size full", stack.Len(), long_size, t)
	c := stack.Top().(compound)
	okEqualInt("value c.i (Top)", c.i, last_value, t)
	okEqualString("value c.s (Top)", c.s, fmt.Sprintf("str%d", last_value), t)
	okEqualBool("value c.b (Top)", c.b, true, t)

	c = stack.Bottom().(compound)
	okEqualInt("value c.i (Bottom)", c.i, first_value, t)
	okEqualString("value c.s (Bottom)", c.s, fmt.Sprintf("str%d", first_value), t)
	okEqualBool("value c.b (Bottom)", c.b, false, t)
	c = stack.Pop().(compound)
	okEqualInt("value c.i (Pop)", c.i, last_value, t)
	okEqualString("value c.s (Pop)", c.s, fmt.Sprintf("str%d", last_value), t)
	okEqualBool("value c.b (Pop)", c.b, true, t)
	c = stack.Pop().(compound)
	okEqualInt("value c.i", c.i, last_value-1, t)
	okEqualString("value c.s ", c.s, fmt.Sprintf("str%d", last_value-1), t)
	okEqualBool("value c.b", c.b, false, t)
	stack.Reset()
	okEqualInt("size after Reset", stack.Len(), empty_size, t)
}
