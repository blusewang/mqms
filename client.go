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

type Event struct {
	TransactionID uuid.UUID       `json:"transaction_id"` // TransactionID 业务ID，是所有事件的根
	ID            uuid.UUID       `json:"id"`             // ID 事件ID
	ParentID      *uuid.UUID      `json:"parent_id"`      // ParentID 来源事件ID
	Delay         time.Duration   `json:"delay"`
	Path          string          `json:"path"`
	Body          json.RawMessage `json:"body"`
	CreateAt      time.Time       `json:"create_at"`
}

type Trace struct {
	Event
	ExecID  uuid.UUID   `json:"exec_id"`
	Status  TraceStatus `json:"status"`
	BeginAt time.Time   `json:"begin_at"`
	EndAt   *time.Time  `json:"end_at"`
	Error   string      `json:"error"`
	Stack   string      `json:"stack"`
}

type IClientHandler interface {
	// Pub 发布消息至队列的代理
	Pub(evt json.RawMessage, duration time.Duration) error
	// Save 存储延迟事件至数据库
	Save(evt json.RawMessage, duration time.Duration) error
	// Trace 链路信息
	Trace(trace Trace)
}

type Client struct {
	name    string
	handler IClientHandler
}

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
		return c.handler.Save(raw, duration)
	} else {
		return c.handler.Pub(raw, duration)
	}
}

func NewClient(name string, handler IClientHandler) *Client {
	return &Client{
		name:    name,
		handler: handler,
	}
}
