// Copyright 2022 YBCZ, Inc. All rights reserved.
//
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file in the root of the source
// tree.

package mqms

import (
	"fmt"
	"reflect"
	"runtime"
	"time"
)

func nameOfFunction(f interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
}

func stack() []string {
	var lines []string
	for i := 2; ; i++ {
		_, f, n, ok := runtime.Caller(i)
		if !ok {
			break
		}
		lines = append(lines, fmt.Sprintf("%s:%d", f, n))
	}
	return lines
}

func normalLogFormat(format string, a ...any) string {
	a = append([]any{time.Now().Format("15:04:05")}, a...)
	return fmt.Sprintf("[MQMS] %v "+format+"\n", a...)
}

func stringPtr(s string) *string {
	if s == "" {
		return nil
	} else {
		return &s
	}
}
