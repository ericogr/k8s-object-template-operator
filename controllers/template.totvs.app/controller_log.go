/*


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"errors"
	"strings"

	"github.com/go-logr/logr"
)

type singleLog struct {
	error
	message string
}

// LogUtil log utilities
type LogUtil struct {
	Log        logr.Logger
	singleLogs []singleLog
}

// AppendErrorMessage show error and append messages to array
func (lu *LogUtil) Error(err error, msg string) {
	if lu.singleLogs == nil {
		lu.singleLogs = []singleLog{}
	}

	lu.singleLogs = append(lu.singleLogs, singleLog{err, msg})
	lu.Log.Error(err, msg)
}

// HasError has any error
func (lu *LogUtil) HasError() bool {
	return len(lu.singleLogs) > 0
}

// AllErrors all errors
func (lu *LogUtil) AllErrors() error {
	if lu.HasError() {
		return errors.New(lu.AllErrorsMessages())
	}

	return nil
}

// AllErrorsMessages all logs message to string
func (lu *LogUtil) AllErrorsMessages() string {
	sb := strings.Builder{}
	for _, sl := range lu.singleLogs {
		sb.WriteString(sl.message)
		sb.WriteString(", ")
	}

	return sb.String()
}
