// Copyright 2022 YBCZ, Inc. All rights reserved.
//
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file in the root of the source
// tree.

package mqms

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"runtime/debug"
	"sync"
	"time"
)

type Row struct {
	Raw json.RawMessage `json:"raw"`
	At  time.Time       `json:"at"`
}

type IEngineHandler interface {
	IClientHandler
	// DoRead 从数据库中读出 time.Now().Add(time.Minute) 之前所有没有入列的消息
	DoRead() (list []Row, err error)
	// Fail 失败通知
	Fail(evt Event, err error, stack []byte)
	// Log 引擎日志
	Log(l string)
	// HttpTrace http请求监测
	HttpTrace(ht HttpTrace)
}

type Engine struct {
	Router
	ctx     context.Context
	handler IEngineHandler
	routes  map[string][]HandlerFunc
	pool    sync.Pool
}

func (s *Engine) Shutdown() (err error) {
	select {
	case <-s.ctx.Done():
		return
	}
}

func (s *Engine) readLooper() {
	list, err := s.handler.DoRead()
	if err == nil {
		for i := range list {
			d := list[i].At.Sub(time.Now())
			if d < 0 {
				s.Handle(list[i].Raw)
			} else {
				if err = s.handler.Pub(list[i].Raw, d); err != nil {
					s.handler.Log(fmt.Sprintf("消息入列🙅：%v", err))
				}
			}
		}
	} else {
		s.handler.Log(fmt.Sprintf("消息读取🙅：%v", err))
	}
	time.Sleep(time.Minute)
	go s.readLooper()
}

// Handle 消息入口
func (s *Engine) Handle(raw json.RawMessage) {
	var trace Trace
	if err := json.Unmarshal(raw, &trace.Event); err != nil {
		s.handler.Log(fmt.Sprintf("消息解码🙅：%v", err))
		return
	}
	trace.ExecID = uuid.New()
	if s.routes[trace.Event.Path] != nil {
		trace.Status = TraceStatusExecBegin
		trace.BeginAt = time.Now()
		s.handler.Trace(trace)

		// 构建上下文
		c := s.pool.Get().(*Context)
		c.reset()
		c.evt = trace.Event
		c.engine = s
		c.handlers = s.routes[trace.Event.Path]

		defer func() {
			if e2 := recover(); e2 != nil {
				trace.Status = TraceStatusError
				trace.Error = e2.(error).Error()
				trace.Stack = string(debug.Stack())
				s.handler.Trace(trace)
				s.handler.Log(fmt.Sprintf("[panic] [%v] --> %v : %v\n", trace.Event.Path, nameOfFunction(c.handlers[c.index]), e2))
			}
		}()

		// 开始执行
		err := c.Next()

		// 记录并回收
		trace.EndAt = new(time.Time)
		*trace.EndAt = time.Now()
		if err != nil {
			trace.Status = TraceStatusError
			trace.Error = c.err
			trace.Stack = c.stack
			s.handler.Trace(trace)
			s.handler.Log(fmt.Sprintf("[error] [%v] --> %v : %v\n", trace.Event.Path, nameOfFunction(c.handlers[c.index-1]), err))
		} else {
			trace.Status = TraceStatusExecEnd
			s.handler.Trace(trace)
		}
		s.pool.Put(c)
	} else {
		trace.Status = TraceStatusError
		trace.Error = "没有可命中的服务"
		s.handler.Trace(trace)
		s.handler.Log(fmt.Sprintf("[error] [%v] --> nil : %v\n", trace.Event.Path, trace.Error))
	}
	return
}

func New(handler IEngineHandler) (e *Engine) {
	e = &Engine{
		ctx:     context.Background(),
		handler: handler,
		routes:  make(map[string][]HandlerFunc),
	}
	e.engine = e
	e.pool.New = func() any {
		return &Context{
			ctx:    context.TODO(),
			evt:    Event{},
			engine: e,
			Http:   client(e),
		}
	}
	go e.readLooper()
	return
}
