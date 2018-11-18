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
	"path"
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
	var newArgs []interface{}
	caller := CallFuncName()
	opNum := GetOperationNumber(caller)

	// injects operation number and caller into the function arguments
	newArgs = append(newArgs, opNum)
	for _, arg := range args {
		newArgs = append(newArgs, arg)
	}
	// Calls the logger's Printf function, with the additional prefix
	l.logger.Printf("[%s] "+format, newArgs...)
}

var operationNum int

func GetOperationNumber(caller string) string {
	caller = common.BaseName(caller)
	var m sync.Mutex
	m.Lock()
	operationNum += 1
	operationId := fmt.Sprintf("%07d-%05d %s", os.Getpid(), operationNum, caller)
	m.Unlock()
	return operationId
}

func NewLogger(logDir, logFileName string) (error, string, *Logger) {
	noLogger := &Logger{logger: log.New(ioutil.Discard, "", log.Ldate|log.Ltime)}
	if !LogSBOperations {
		return nil, "", noLogger
	}
	if !common.DirExists(Defaults().LogDirectory) {
		err := os.Mkdir(Defaults().LogDirectory, 0755)
		if err != nil {
			return err, "", noLogger
		}
	}
	fullLogDir := path.Join(Defaults().LogDirectory, logDir)
	if !common.DirExists(fullLogDir) {
		err := os.Mkdir(fullLogDir, 0755)
		if err != nil {
			return err, "", noLogger
		}
	}
	var logFileFullName string = path.Join(fullLogDir, logFileName+".log")
	var err error
	var logFile *os.File
	logFile, err = os.OpenFile(logFileFullName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		return fmt.Errorf("error opening log file %s : %v", logFileFullName, err), "", noLogger
	}
	return nil, logFileFullName, &Logger{logger: log.New(logFile, "", log.Ldate|log.Ltime)}
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
