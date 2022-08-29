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

package concurrent

import (
	"fmt"
	"sort"
	"testing"
)

type Times []int64

func (t Times) Len() int {
	return len(t)
}

func (t Times) Less(i, j int) bool {
	return t[i] < t[j]
}

func (t Times) Swap(i, j int) {
	tmp := t[i]
	t[i] = t[j]
	t[j] = tmp
}

func TestConcurrency(t *testing.T) {

	type CheckConcurrency struct {
		level         int
		ID            string
		expectedIndex int
	}

	// Initialization is given in unsorted order
	// The concurrency function RunParallelTasksByPriority is supposed
	// to sort the tasks by priority and queue them accordingly.
	var testData = []CheckConcurrency{
		{0, "one", 0},
		{0, "two", 1},
		{0, "three", 2},
		{3, "one.3", 7},
		{2, "one.2", 6},
		{1, "one.1", 4},
		{0, "four", 3},
		{1, "three.1", 5},
	}
	var results []string
	var times Times
	var traceConcurrency = func(ti TraceInfo) {
		// This function is called by RunParallelTasksByPriority during the sorting
		fmt.Printf("%2d %-10s %d\n", ti.Level, ti.Args[0], ti.Time.UnixNano())

		// The execution commands are inserted here after they are sorted
		// by RunParallelTasksByPriority
		results = append(results, ti.Args[0])
		// For an extra check, we also save the time when the item was processed
		// for concurrency
		times = append(times, ti.Time.UnixNano())
	}

	// We build an execution list based on the test data
	var execLists []ExecutionList
	for _, item := range testData {
		var execList = ExecutionList{
			Logger:   nil,
			Priority: item.level,
			Command: ExecCommand{
				Tracer: traceConcurrency,
				Cmd:    "echo",
				Args:   []string{item.ID},
			},
		}
		execLists = append(execLists, execList)
	}

	RunParallelTasksByPriority(execLists)

	fmt.Printf("# Results:\n")
	for N, item := range testData {
		fmt.Printf("index: %2d [level: %2d] ID: %-10s - expected index: %d\n", N, item.level, item.ID, item.expectedIndex)
	}
	// Now we check that the items are in the expected position within the results array.
	for _, item := range testData {
		wantedIndex := item.expectedIndex
		if len(results) < wantedIndex {
			t.Fatalf("result does not have %d items", wantedIndex)
		}
		if results[wantedIndex] == item.ID {
			t.Logf("ok - Item %-10s is at position %d in the results\n", item.ID, wantedIndex)
		} else {
			t.Logf(" not ok - Item %-10s is not at position %d in the results\n", item.ID, wantedIndex)
			t.Fail()
		}
	}
	// We also check that the times are sorted. If they are, it's further evidence that
	// the results are in the right order.
	if sort.IsSorted(times) {
		t.Logf("ok - Times for the executed elements are sorted")
	} else {
		for N, t := range times {
			fmt.Printf(" #time> %2d %d\n", N, t)
		}
		t.Logf(" not ok - Times for the executed elements are not sorted")
		t.Fail()
	}
}
