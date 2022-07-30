// Copyright 2022 YBCZ, Inc. All rights reserved.
//
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file in the root of the source
// tree.

package mqms

import (
	"encoding/json"
	"github.com/google/uuid"
	"time"
)

// Event 事件
type Event struct {
	TransactionID uuid.UUID       `json:"transaction_id"` // TransactionID 业务ID，是所有事件的根
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
	Pub(evtRaw json.RawMessage, duration time.Duration) error
	// Save 存储延迟事件至数据库
	Save(evtID uuid.UUID, evtRaw json.RawMessage, duration time.Duration) error
	// Trace 链路信息
	Trace(trace Trace)
}

// IClient 客户端协议
type IClient interface {
	// Emit 发布事件
	Emit(path string, body interface{}) (err error)
	// EmitDefer 发布延迟事件
	EmitDefer(path string, body interface{}, duration time.Duration) (err error)
}

// Client 客户端
type Client struct {
	name    string
	handler IClientHandler
}

// Emit 发布实时事件
func (c *Client) Emit(path string, body interface{}) (err error) {
	var evt Event
	evt.TransactionID = uuid.New()
	evt.ID = uuid.New()
	evt.Path = path
	evt.CreateAt = time.Now()
	evt.Body, _ = json.Marshal(body)
	raw, _ := json.Marshal(evt)
	defer c.handler.Trace(Trace{
		Status:  TraceStatusEmit,
		Event:   evt,
		BeginAt: time.Now(),
	})
	return c.handler.Pub(raw, 0)
}

// EmitDefer 发岸上延时事件
func (c *Client) EmitDefer(path string, body interface{}, duration time.Duration) (err error) {
	var evt Event
	evt.Path = path
	evt.TransactionID = uuid.New()
	evt.ID = uuid.New()
	evt.CreateAt = time.Now()
	evt.Body, _ = json.Marshal(body)
	raw, _ := json.Marshal(evt)
	defer c.handler.Trace(Trace{
		Status:  TraceStatusError,
		Event:   evt,
		BeginAt: time.Now(),
	})
	if duration > time.Minute {
		return c.handler.Save(evt.ID, raw, duration)
	} else {
		return c.handler.Pub(raw, duration)
	}
}

// NewClient 创建客户端
func NewClient(name string, handler IClientHandler) *Client {
	return &Client{
		name:    name,
		handler: handler,
	}
}
