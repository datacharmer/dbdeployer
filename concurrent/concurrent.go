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
	"os"
	"os/exec"
	"sync"
)

type CommonChan chan *exec.Cmd

type ExecCommand struct {
	Cmd  string
	Args []string
}

type ExecCommands []ExecCommand

type ExecutionList struct {
	Priority int
	Command  ExecCommand
}

var DebugConcurrency bool
var VerboseConcurrency bool

func add_task(num int, wg *sync.WaitGroup, tasks CommonChan, cmd string, args []string) {
	wg.Add(1)
	go start_task(num, wg, tasks)
	tasks <- exec.Command(cmd, args...)
}

func start_task(num int, w *sync.WaitGroup, tasks CommonChan) {
	defer w.Done()
	var (
		out []byte
		err error
	)
	for cmd := range tasks { // this will exit the loop when the channel closes
		out, err = cmd.Output()
		if err != nil {
			fmt.Printf("Error executing goroutine %d : %s", num, err)
			//os.Exit(1)
		}
		if DebugConcurrency {
			fmt.Printf("goroutine %d command output: %s", num, string(out))
		} else {
			if VerboseConcurrency {
				fmt.Printf("%s", string(out))
			}
		}
	}
}

// Run several tasks in parallel

func RunParallelTasks(priority_level int, operations ExecCommands) {
	tasks := make(CommonChan, 64)

	var wg sync.WaitGroup

	for N, ec := range operations {
		add_task(N, &wg, tasks, ec.Cmd, ec.Args)
	}
	close(tasks)
	wg.Wait()
	if VerboseConcurrency {
		fmt.Printf("#%d\n", priority_level)
	}
}

/*
//  Given a list of tasks with different priorities
// This function organizes the queued tasks by priority
// and runs concurrently the tasks with the same priority
// until no task is left in the queue.
// For example we may have:
	priority     command
	1            /some/path/init_db
	2            /some/path/start
	3            /some/path/load_grants
	1            /some/other/path/init_db
	2            /some/other/path/start
	3            /some/other/path/load_grants
	1            /some/alternative/path/init_db
	2            /some/alternative/path/start
	3            /some/alternative/path/load_grants

	This function will receive the commands, and re-arrange them as follows
	run concurrently: {
		1            /some/path/init_db
		1            /some/other/path/init_db
		1            /some/alternative/path/init_db
	}

	run concurrently: {
		2            /some/path/start
		2            /some/other/path/start
		2            /some/alternative/path/start
	}

	run concurrently: {
		3            /some/path/load_grants
		3            /some/other/path/load_grants
		3            /some/alternative/path/load_grants
	}
*/

func RunParallelTasksByPriority(exec_lists []ExecutionList) {
	maxPriority := 0
	if len(exec_lists) == 0 {
		return
	}
	if DebugConcurrency {
		fmt.Printf("RunParallelTasksByPriority exec_list %#v\n", exec_lists)
	}
	for _, list := range exec_lists {
		if list.Priority > maxPriority {
			maxPriority = list.Priority
		}
	}
	for N := 0; N <= maxPriority; N++ {
		var operations ExecCommands
		for _, list := range exec_lists {
			if list.Priority == N {
				operations = append(operations, list.Command)
			}
		}
		if DebugConcurrency {
			fmt.Printf("%d %v\n", N, operations)
		}
		RunParallelTasks(N, operations)
	}
}

func init() {
	if os.Getenv("DEBUG_CONCURRENCY") != "" {
		DebugConcurrency = true
	}
	if os.Getenv("VERBOSE_CONCURRENCY") != "" {
		VerboseConcurrency = true
	}
}
