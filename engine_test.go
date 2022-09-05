// Copyright 2022 YBCZ, Inc. All rights reserved.
//
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file in the root of the source
// tree.

package mqms

import (
	"encoding/json"
	"github.com/google/uuid"
	"log"
	"testing"
	"time"
)

type testHandler struct {
}

func (t testHandler) Pub(evtRaw json.RawMessage, duration time.Duration) (err error) {
	log.Println(string(evtRaw), duration)
	return
}

func (t testHandler) Save(evtID uuid.UUID, evtRaw json.RawMessage, duration time.Duration) (err error) {
	log.Println(evtID, string(evtRaw), duration)
	return
}

func (t testHandler) Trace(trace Trace) {
	raw, _ := json.Marshal(trace)
	log.Println(string(raw))
}

func (t testHandler) Log(l string) {
	log.Println(l)
}

func (t testHandler) Fail(evtID uuid.UUID, evtRaw json.RawMessage, err error, stack []string) {
	log.Println(evtID, string(evtRaw), err, stack)
}

func (t testHandler) DoRead() (list []Row, err error) {
	return
}

func (t testHandler) HttpTrace(ht HttpTrace) {
	raw, _ := json.Marshal(ht)
	log.Println(string(raw))
}

var _ IEngineHandler = &testHandler{}

func TestMqmsTrace(t *testing.T) {
	log.SetFlags(log.Ltime | log.Lshortfile)
	var h = new(testHandler)
	engine := New(h)
	engine.Item("test", func(c *Context) (err error) {
		//err = c.Error(errors.New("test error"))
		return
	})
	engine.Emit("test", "F")
	time.Sleep(time.Hour)
}
