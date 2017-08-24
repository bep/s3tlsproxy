// Copyright © 2017 Bjørn Erik Pedersen <bjorn.erik.pedersen@gmail.com>.
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

package lib

import (
	"fmt"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
)

type Logger struct {
	logger      log.Logger
	errorLogger log.Logger
	infoLogger  log.Logger
	debugLogger log.Logger
}

func NewLogger(logger log.Logger) *Logger {
	return &Logger{
		errorLogger: level.Error(logger),
		infoLogger:  level.Info(logger),
		debugLogger: level.Debug(logger),
	}
}

func (l *Logger) Error(keyvals ...interface{}) error {
	if err := l.errorLogger.Log(keyvals...); err != nil {
		fmt.Println("error when logging:", err)
	}
	return nil
}

func (l *Logger) Info(keyvals ...interface{}) error {
	if err := l.infoLogger.Log(keyvals...); err != nil {
		fmt.Println("error when logging:", err)
	}
	return nil
}

func (l *Logger) Debug(keyvals ...interface{}) error {
	if err := l.debugLogger.Log(keyvals...); err != nil {
		fmt.Println("error when logging:", err)
	}
	return nil
}
