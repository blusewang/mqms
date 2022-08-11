// Copyright 2022 YBCZ, Inc. All rights reserved.
//
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file in the root of the source
// tree.

package mqms

import (
	"context"
	"encoding/json"
	"github.com/google/uuid"
	"math"
	"net/http"
	"runtime/debug"
	"time"
)

// Context 函数上下文
type Context struct {
	ctx      context.Context
	evt      Event
	engine   *Engine
	Http     http.Client
	index    int
	handlers []HandlerFunc
	err      string
	stack    string
}

func (c *Context) reset() {
	c.ctx = context.TODO()
	c.index = -1
	c.handlers = make([]HandlerFunc, 0)
	c.err = ""
	c.stack = ""
}

func (c *Context) Bind(o interface{}) (err error) {
	return json.Unmarshal(c.evt.Body, o)
}

// Next 先执行下一个函数
func (c *Context) Next() (err error) {
	c.index++
	for c.index < len(c.handlers) {
		err = c.handlers[c.index](c)
		c.index++
	}
	return
}

// Abort 中止后面的调用链
func (c *Context) Abort() {
	c.index = math.MaxInt16 >> 1
}

// Error 错误处理
func (c *Context) Error(es error) error {
	if es != nil {
		c.err = es.Error()
		c.stack = string(debug.Stack())
	}
	return es
}

// Emit 内部发布事件，在新协程中直接执行
func (c *Context) Emit(path string, body interface{}) {
	go func() {
		var evt Event
		evt.TransactionID = c.evt.TransactionID
		evt.ID = uuid.New()
		evt.ParentID = &c.evt.ID
		evt.Path = path
		evt.CreateAt = time.Now()
		evt.Body, _ = json.Marshal(body)
		evt.Origin = stack()
		raw, _ := json.Marshal(evt)
		defer c.engine.handler.Trace(Trace{
			Status:  TraceStatusEmit,
			Event:   evt,
			BeginAt: time.Now(),
		})
		c.engine.Handle(raw)
	}()
}

// EmitDefer 内部发布延时事件，按情况将消息送入队列或存储
func (c *Context) EmitDefer(path string, body interface{}, duration time.Duration) {
	if duration == 0 {
		c.Emit(path, body)
		return
	}
	var evt Event
	evt.Path = path
	evt.TransactionID = c.evt.TransactionID
	evt.ID = uuid.New()
	evt.ParentID = &c.evt.ID
	evt.CreateAt = time.Now()
	evt.Body, _ = json.Marshal(body)
	evt.Origin = stack()
	raw, _ := json.Marshal(evt)
	defer c.engine.handler.Trace(Trace{
		Status:  TraceStatusEmit,
		Event:   evt,
		BeginAt: time.Now(),
	})
	if duration > time.Minute {
		if err := c.engine.handler.Save(evt.ID, raw, duration); err != nil {
			c.engine.handler.Log(normalLogFormat("事件存储错误：%v", err.Error()))
			c.engine.handler.Fail(evt.ID, raw, err, string(debug.Stack()))
		}
	} else {
		if err := c.engine.handler.Pub(raw, duration); err != nil {
			c.engine.handler.Log(normalLogFormat("事件发布错误：%v", err.Error()))
			c.engine.handler.Fail(evt.ID, raw, err, string(debug.Stack()))
		}
	}
}
