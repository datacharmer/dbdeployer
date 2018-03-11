package concurrent

import (
    "fmt"
    "os"
    "os/exec"
    "sync"
)

type CommonChan chan *exec.Cmd

type ExecCommand struct {
	Cmd string
	Args []string
}

type ExecCommands []ExecCommand

type ExecutionList struct {
	Priority int
	Command ExecCommand
}

var DebugConcurrency bool 
var VerboseConcurrency bool 


func add_task (num int, wg *sync.WaitGroup, tasks CommonChan, cmd string, args []string) {
	wg.Add(1)
	go start_task(num, wg, tasks)
    tasks <- exec.Command(cmd, args...)
}

func start_task (num int, w *sync.WaitGroup, tasks CommonChan) {
	defer w.Done()
	var (
		out []byte
		err error
	)
	for cmd := range tasks { // this will exit the loop when the channel closes
		out, err = cmd.Output()
		if err != nil {
			fmt.Printf("Error executing goroutine %d : %s", num,  err)
			//os.Exit(1)
		}
		if DebugConcurrency {
			fmt.Printf("goroutine %d command output: %s", num, string(out))
		} else {
			if VerboseConcurrency {
				fmt.Printf("%s",string(out))
			}
		}
	}
}


func RunParallelTasks( priority_level int, operations ExecCommands ) {
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


func RunParallelTasksByPriority ( exec_lists []ExecutionList) {
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
	for N := 0; N <= maxPriority ; N++ {
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
