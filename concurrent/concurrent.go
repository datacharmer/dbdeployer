// DBDeployer - The MySQL Sandbox
// Copyright © 2006-2020 Giuseppe Maxia
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
	"os/exec"
	"sync"
	"time"

	"github.com/datacharmer/dbdeployer/common"
	"github.com/datacharmer/dbdeployer/defaults"
)

type CommonChan chan *exec.Cmd
type TraceInfo struct {
	Time  time.Time
	Cmd   string
	Args  []string
	Level int
}
type Trace func(ti TraceInfo)

type ExecCommand struct {
	Cmd    string
	Args   []string
	Tracer Trace
}

type ExecCommands []ExecCommand

type ExecutionList struct {
	Logger   *defaults.Logger
	Priority int
	Command  ExecCommand
}

var DebugConcurrency bool
var VerboseConcurrency bool

func addTask(num int, wg *sync.WaitGroup, tasks CommonChan, cmd string, args []string) {
	wg.Add(1)
	go startTask(num, wg, tasks)
	// #nosec G204
	tasks <- exec.Command(cmd, args...)
}

func startTask(num int, w *sync.WaitGroup, tasks CommonChan) {
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

func runParallelTasks(priorityLevel int, operations ExecCommands) {
	tasks := make(CommonChan, 64)

	var wg sync.WaitGroup

	for N, ec := range operations {
		addTask(N, &wg, tasks, ec.Cmd, ec.Args)
	}
	close(tasks)
	wg.Wait()
	if VerboseConcurrency {
		fmt.Printf("#%d\n", priorityLevel)
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

func RunParallelTasksByPriority(execLists []ExecutionList) {
	maxPriority := 0
	if len(execLists) == 0 {
		return
	}
	if DebugConcurrency {
		fmt.Printf("RunParallelTasksByPriority exec_list %#v\n", execLists)
	}
	for _, list := range execLists {
		if list.Priority > maxPriority {
			maxPriority = list.Priority
		}
	}
	for N := 0; N <= maxPriority; N++ {
		var operations ExecCommands
		for _, list := range execLists {
			if list.Priority == N {
				operations = append(operations, list.Command)
				if list.Command.Tracer != nil {
					list.Command.Tracer(TraceInfo{Time: time.Now(), Cmd: list.Command.Cmd, Args: list.Command.Args, Level: list.Priority})
				}
				if list.Logger != nil {
					list.Logger.Printf(" Queueing command %s [%v] with priority # %d\n",
						list.Command.Cmd, list.Command.Args, list.Priority)
				}
			}
		}
		if DebugConcurrency {
			fmt.Printf("%d %v\n", N, operations)
		}
		runParallelTasks(N, operations)
	}
}

func init() {
	if common.IsEnvSet("DEBUG_CONCURRENCY") {
		DebugConcurrency = true
	}
	if common.IsEnvSet("VERBOSE_CONCURRENCY") {
		VerboseConcurrency = true
	}
}
