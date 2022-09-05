// Copyright 2022 YBCZ, Inc. All rights reserved.
//
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file in the root of the source
// tree.

package mqms

import (
	"context"
	"encoding/json"
	"log"
	"runtime"
	"runtime/debug"
	"testing"
	"time"
)

func TestUUID(t *testing.T) {
	var evt, e Trace
	evt.Delay = time.Minute
	raw, _ := json.Marshal(evt)
	log.Println(string(raw))
	log.Println(json.Unmarshal(raw, &e))
	log.Println(e.Delay)
}

func TestCtx(t *testing.T) {
	c := context.TODO()
	c.Value("sdf")
}

func TestStack(t *testing.T) {
	raw := debug.Stack()
	log.Println(string(raw))
	for i := 0; ; i++ {
		_, f, n, ok := runtime.Caller(i)
		if !ok {
			break
		}
		log.Println(f, n)
	}

}
