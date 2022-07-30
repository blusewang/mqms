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

// Row 存储单个事件的数据结构
type Row struct {
	Raw json.RawMessage `json:"raw"`
	At  time.Time       `json:"at"`
}

type IEngineHandler interface {
	IClientHandler
	// DoRead 从数据库中读出 time.Now().Add(time.Minute) 之前所有没有入列的消息
	DoRead() (list []Row, err error)
	// Fail 失败通知
	Fail(evtID uuid.UUID, evtRaw json.RawMessage, err error, stack string)
	// Log 引擎日志
	Log(l string)
	// HttpTrace http请求监测
	HttpTrace(ht HttpTrace)
}

// Engine 主引擎
type Engine struct {
	Route
	ctx     context.Context
	handler IEngineHandler
	routes  map[string][]HandlerFunc
	pool    sync.Pool
	gw      sync.WaitGroup
}

// Shutdown 安全地中止业务，等待最后一个函数执行完
func (s *Engine) Shutdown() {
	s.gw.Wait()
}

// Emit 发事件，在新协程中直接执行
func (s *Engine) Emit(path string, body interface{}) (err error) {
	go func() {
		var evt Event
		evt.TransactionID = uuid.New()
		evt.ID = uuid.New()
		evt.Path = path
		evt.CreateAt = time.Now()
		evt.Body, _ = json.Marshal(body)
		raw, _ := json.Marshal(evt)
		defer s.handler.Trace(Trace{
			Status:  TraceStatusEmit,
			Event:   evt,
			BeginAt: time.Now(),
		})
		s.Handle(raw)
	}()
	return
}

// EmitDefer 按情况将消息送入队列或存储
func (s *Engine) EmitDefer(path string, body interface{}, duration time.Duration) (err error) {
	if duration == 0 {
		return s.Emit(path, body)
	}
	var evt Event
	evt.Path = path
	evt.TransactionID = uuid.New()
	evt.ID = uuid.New()
	evt.CreateAt = time.Now()
	evt.Body, _ = json.Marshal(body)
	raw, _ := json.Marshal(evt)
	defer s.handler.Trace(Trace{
		Status:  TraceStatusError,
		Event:   evt,
		BeginAt: time.Now(),
	})
	if duration > time.Minute {
		return s.handler.Save(evt.ID, raw, duration)
	} else {
		return s.handler.Pub(raw, duration)
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
	s.gw.Add(1)
	defer s.gw.Done()
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

		// 闪退捕获
		defer func() {
			if e2 := recover(); e2 != nil {
				trace.Status = TraceStatusError
				trace.Error = e2.(error).Error()
				trace.Stack = string(debug.Stack())
				s.handler.Fail(trace.Event.ID, raw, e2.(error), trace.Stack)
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
			s.handler.Fail(trace.Event.ID, raw, err, trace.Stack)
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

// New 创建微服务
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
