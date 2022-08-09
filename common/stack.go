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
	"container/list"
	"sync"
)

// The stack is implemented using a double-linked list from
// Go standard library
type Stack struct {
	list *list.List
	mux  sync.Mutex
}

// NewStack returns a new stack
func NewStack() (stack Stack) {
	stack.list = list.New()
	return
}

// The length of the stack is that of the underlying list
func (stack *Stack) Len() int {
	return stack.list.Len()
}

// Removes all items from the stack
func (stack *Stack) Reset() {
	stack.mux.Lock()
	for stack.list.Len() > 0 {
		stack.list.Remove(stack.list.Front())
	}
	stack.mux.Unlock()
}

// Push() inserts an item to the front of the stack.
// The item can be of any type
func (stack *Stack) Push(item interface{}) {
	stack.list.PushFront(item)
}

// Pop() returns the object stored as .Value in the top list element
// Client calls will need to cast the object to the expected type.
// The object is removed from the list
// e.g.:
//
//	type MyType struct { ... }
//	var lastOne MyType
//	lastOne = stack.Pop().(MyType)
func (stack *Stack) Pop() interface{} {
	// Locks so that it is safe to pop from concurrent goroutines
	stack.mux.Lock()
	latest := stack.list.Front().Value
	// After extracting the item, we remove the first
	// list element
	stack.list.Remove(stack.list.Front())
	stack.mux.Unlock()
	return latest
}

// Top() returns the object stored as .Value in the top list element
// The object is NOT removed from the list
func (stack *Stack) Top() interface{} {
	return stack.list.Front().Value
}

// Bottom() returns the object stored as .Value in the bottom list element
// The object is NOT removed from the list
func (stack *Stack) Bottom() interface{} {
	return stack.list.Back().Value
}
