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
	"github.com/datacharmer/dbdeployer/compare"
	"testing"
)

func TestStack(t *testing.T) {

	type compound struct {
		i int
		s string
		b bool
	}

	emptySize := 0
	smallSize := 1
	longSize := 10000
	firstValue := 0
	lastValue := longSize - 1
	intValue := 1234
	stack := NewStack()
	stack.Push(intValue)
	compare.OkEqualInt("size full", stack.Len(), smallSize, t)
	x := stack.Pop().(int)
	compare.OkEqualInt("value x", x, intValue, t)
	compare.OkEqualInt("size empty", stack.Len(), emptySize, t)

	for N := 0; N < longSize; N++ {
		stack.Push(N) // last one is 99
	}
	compare.OkEqualInt("size full", stack.Len(), longSize, t)
	x = stack.Top().(int)
	compare.OkEqualInt("value x (Top)", x, lastValue, t)
	x = stack.Bottom().(int)
	compare.OkEqualInt("value x (Bottom)", x, firstValue, t)
	x = stack.Pop().(int)
	compare.OkEqualInt("value x (Pop)", x, lastValue, t)
	compare.OkEqualInt("size after pop", stack.Len(), lastValue, t)
	for stack.Len() > 0 {
		stack.Pop()
	}
	compare.OkEqualInt("size empty", stack.Len(), emptySize, t)
	for N := 0; N < longSize; N++ {
		odd := true
		if N%2 == 0 {
			odd = false
		}
		c := compound{N, fmt.Sprintf("str%d", N), odd}
		stack.Push(c)
	}
	compare.OkEqualInt("size full", stack.Len(), longSize, t)
	c := stack.Top().(compound)
	compare.OkEqualInt("value c.i (Top)", c.i, lastValue, t)
	compare.OkEqualString("value c.s (Top)", c.s, fmt.Sprintf("str%d", lastValue), t)
	compare.OkEqualBool("value c.b (Top)", c.b, true, t)

	c = stack.Bottom().(compound)
	compare.OkEqualInt("value c.i (Bottom)", c.i, firstValue, t)
	compare.OkEqualString("value c.s (Bottom)", c.s, fmt.Sprintf("str%d", firstValue), t)
	compare.OkEqualBool("value c.b (Bottom)", c.b, false, t)
	c = stack.Pop().(compound)
	compare.OkEqualInt("value c.i (Pop)", c.i, lastValue, t)
	compare.OkEqualString("value c.s (Pop)", c.s, fmt.Sprintf("str%d", lastValue), t)
	compare.OkEqualBool("value c.b (Pop)", c.b, true, t)
	c = stack.Pop().(compound)
	compare.OkEqualInt("value c.i", c.i, lastValue-1, t)
	compare.OkEqualString("value c.s ", c.s, fmt.Sprintf("str%d", lastValue-1), t)
	compare.OkEqualBool("value c.b", c.b, false, t)
	stack.Reset()
	compare.OkEqualInt("size after Reset", stack.Len(), emptySize, t)
}
