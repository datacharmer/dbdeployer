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
package defaults

import (
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/datacharmer/dbdeployer/common"
	"io/ioutil"
	"runtime"
)

type Logger struct {
	logger *log.Logger
}

// Calling Logger.Printf will print what was requested,
// with an additional prefix made of :
// 		* dbdeployer Process ID
// 		* the current operation number
// 		* the name of the caller function
func (l *Logger) Printf(format string, args ...interface{}) {
	var new_args []interface{}
	caller := CallFuncName()
	op_num := GetOperationNumber(caller)

	// injects operation number and caller into the function arguments
	new_args = append(new_args, op_num)
	for _, arg := range args {
		new_args = append(new_args, arg)
	}
	// Calls the logger's Printf function, with the additional prefix
	l.logger.Printf("[%s] "+format, new_args...)
}

var operationNum int

func GetOperationNumber(caller string) string {
	caller = common.BaseName(caller)
	var m sync.Mutex
	m.Lock()
	operationNum += 1
	operation_id := fmt.Sprintf("%07d-%05d %s", os.Getpid(), operationNum, caller)
	m.Unlock()
	return operation_id
}

func NewLogger(log_dir, log_file_name string) (string, *Logger) {
	if !LogSBOperations {
		return "", &Logger{logger: log.New(ioutil.Discard, "", log.Ldate|log.Ltime)}
	}
	if !common.DirExists(Defaults().LogDirectory) {
		common.Mkdir(Defaults().LogDirectory)
	}
	full_log_dir := Defaults().LogDirectory + "/" + log_dir
	if !common.DirExists(full_log_dir) {
		common.Mkdir(full_log_dir)
	}
	var log_file_full_name string = fmt.Sprintf("%s/%s.log", full_log_dir, log_file_name)
	var err error
	var log_file *os.File
	log_file, err = os.OpenFile(log_file_full_name, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		common.Exit(1, fmt.Sprintf("error opening log file %s : %v", log_file_full_name, err))
	}
	return log_file_full_name, &Logger{logger: log.New(log_file, "", log.Ldate|log.Ltime)}
}

func CallFuncName() string {
	fpcs := make([]uintptr, 1)
	n := runtime.Callers(3, fpcs)
	if n > 0 {
		fun := runtime.FuncForPC(fpcs[0] - 1)
		if fun != nil {
			return fun.Name()
		}
	}
	return ""
}
