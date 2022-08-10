// Copyright 2022 YBCZ, Inc. All rights reserved.
//
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file in the root of the source
// tree.

package mqms

import (
	"encoding/json"
	"github.com/google/uuid"
	"runtime/debug"
	"time"
)

// Event 事件
type Event struct {
	TransactionID uuid.UUID       `json:"transaction_id"` // TransactionID 业务ID，是所有事件的根
	Origin        string          `json:"origin"`         // 来源
	ID            uuid.UUID       `json:"id"`             // ID 事件ID
	ParentID      *uuid.UUID      `json:"parent_id"`      // ParentID 来源事件ID
	Delay         time.Duration   `json:"delay"`          // 延迟
	Path          string          `json:"path"`
	Body          json.RawMessage `json:"body"`
	CreateAt      time.Time       `json:"create_at"`
}

// Trace 链路信息
type Trace struct {
	Event
	ExecID  uuid.UUID   `json:"exec_id"`
	Status  TraceStatus `json:"status"`
	BeginAt time.Time   `json:"begin_at"`
	EndAt   *time.Time  `json:"end_at"`
	Error   string      `json:"error"`
	Stack   string      `json:"stack"`
}

// IClientHandler 客户端代理协议
type IClientHandler interface {
	// Pub 发布消息至队列的代理
	Pub(evtRaw json.RawMessage, duration time.Duration) (err error)
	// Save 存储延迟事件至数据库
	Save(evtID uuid.UUID, evtRaw json.RawMessage, duration time.Duration) (err error)
	// Trace 链路信息
	Trace(trace Trace)
	// Log 引擎日志
	Log(l string)
	// Fail 失败通知
	Fail(evtID uuid.UUID, evtRaw json.RawMessage, err error, stack string)
}

// IClient 客户端协议
type IClient interface {
	// Emit 发布事件
	Emit(path string, body interface{})
	// EmitDefer 发布延迟事件
	EmitDefer(path string, body interface{}, duration time.Duration)
	// EmitEvent 发布延迟事件
	EmitEvent(evtRaw json.RawMessage)
}

// Client 客户端
type Client struct {
	name    string
	handler IClientHandler
}

// Emit 发布实时事件
func (c *Client) Emit(path string, body interface{}) {
	var evt Event
	evt.TransactionID = uuid.New()
	evt.ID = uuid.New()
	evt.Path = path
	evt.CreateAt = time.Now()
	evt.Body, _ = json.Marshal(body)
	evt.Origin = stack()
	raw, _ := json.Marshal(evt)
	defer c.handler.Trace(Trace{
		Status:  TraceStatusEmit,
		Event:   evt,
		BeginAt: time.Now(),
	})
	if err := c.handler.Pub(raw, 0); err != nil {
		c.handler.Log("事件发布错误：" + err.Error())
		c.handler.Fail(evt.ID, raw, err, string(debug.Stack()))
	}
	return
}

// EmitDefer 发布延时事件
func (c *Client) EmitDefer(path string, body interface{}, duration time.Duration) {
	var evt Event
	evt.Path = path
	evt.TransactionID = uuid.New()
	evt.ID = uuid.New()
	evt.Delay = duration
	evt.CreateAt = time.Now()
	evt.Body, _ = json.Marshal(body)
	evt.Origin = stack()
	raw, _ := json.Marshal(evt)
	defer c.handler.Trace(Trace{
		Status:  TraceStatusEmit,
		Event:   evt,
		BeginAt: time.Now(),
	})
	if duration > time.Minute {
		if err := c.handler.Save(evt.ID, raw, duration); err != nil {
			c.handler.Log("事件存储错误：" + err.Error())
			c.handler.Fail(evt.ID, raw, err, string(debug.Stack()))
		}
	} else {
		if err := c.handler.Pub(raw, duration); err != nil {
			c.handler.Log("事件发布错误：" + err.Error())
			c.handler.Fail(evt.ID, raw, err, string(debug.Stack()))
		}
	}
	return
}

// EmitEvent 发布事件
func (c *Client) EmitEvent(evtRaw json.RawMessage) {
	var evt Event
	if err := json.Unmarshal(evtRaw, &evt); err != nil {
		c.handler.Log("事件发布格式错误：" + err.Error())
		return
	}
	defer c.handler.Trace(Trace{
		Status:  TraceStatusEmit,
		Event:   evt,
		BeginAt: time.Now(),
	})
	if err := c.handler.Pub(evtRaw, 0); err != nil {
		c.handler.Log("事件发布错误：" + err.Error())
		c.handler.Fail(evt.ID, evtRaw, err, string(debug.Stack()))
	}
	return
}

var _ IClient = &Client{}

// NewClient 创建客户端
func NewClient(name string, handler IClientHandler) *Client {
	return &Client{
		name:    name,
		handler: handler,
	}
}
