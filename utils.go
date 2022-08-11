// Copyright 2022 YBCZ, Inc. All rights reserved.
//
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file in the root of the source
// tree.

package mqms

import (
	"bytes"
	"fmt"
	"reflect"
	"runtime"
	"runtime/debug"
	"time"
)

func nameOfFunction(f interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
}

func stack() string {
	arr := bytes.Split(debug.Stack(), []byte("\n"))
	if len(arr) < 8 {
		return "unknown"
	}
	var lines [][]byte
	for i := 8; i < len(arr); i++ {
		if bytes.HasPrefix(arr[i], []byte("\t")) {
			lines = append(lines, bytes.Split(arr[i][1:], []byte(" "))[0])
		}
		if len(lines) >= 4 {
			break
		}
	}
	return string(bytes.Join(lines, []byte(" | ")))
}

func normalLogFormat(format string, a ...any) string {
	a = append([]any{time.Now().Format("15:04:05")}, a...)
	return fmt.Sprintf("[MQMS]%v "+format+"\n", a...)
}
