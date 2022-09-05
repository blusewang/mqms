// Copyright 2022 YBCZ, Inc. All rights reserved.
//
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file in the root of the source
// tree.

package mqms

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"runtime"
	"runtime/debug"
	"time"
)

func nameOfFunction(f interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
}

func stack() json.RawMessage {
	arr := bytes.Split(debug.Stack(), []byte("\n"))
	if len(arr) < 8 {
		return json.RawMessage("[]")
	}
	var lines []string
	for i := 8; i < len(arr); i++ {
		if bytes.HasPrefix(arr[i], []byte("\t")) {
			lines = append(lines, string(bytes.Split(arr[i][1:], []byte(" "))[0]))
		}
		if len(lines) >= 4 {
			break
		}
	}
	var raw, _ = json.Marshal(lines)
	return raw
}

func normalLogFormat(format string, a ...any) string {
	a = append([]any{time.Now().Format("15:04:05")}, a...)
	return fmt.Sprintf("[MQMS] %v "+format+"\n", a...)
}
