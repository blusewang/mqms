// Copyright 2022 YBCZ, Inc. All rights reserved.
//
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file in the root of the source
// tree.

package mqms

import (
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"time"
)

type event struct {
	ID       uuid.UUID       `json:"id"`
	ParentID *uuid.UUID      `json:"parent_id"`
	Path     string          `json:"path"`
	Body     json.RawMessage `json:"body"`
}

type client struct {
	name      string
	onPub     func(body []byte, duration time.Duration) error
	onStorage func(evt event, duration time.Duration) error
}

func (c *client) SetOnPub(pub func(body []byte, duration time.Duration) error) {
	c.onPub = pub
}

func (c *client) SetOnStorage(storage func(evt event, duration time.Duration) error) {
	c.onStorage = storage
}

func (c *client) Emit(path string, body interface{}) (err error) {
	var evt event
	evt.Path = path
	evt.Body, _ = json.Marshal(body)
	raw, _ := json.Marshal(evt)
	return c.onPub(raw, 0)
}

func (c *client) EmitDefer(path string, body interface{}, duration time.Duration) (err error) {
	var evt event
	evt.Path = path
	evt.Body, _ = json.Marshal(body)
	raw, _ := json.Marshal(evt)
	if duration > time.Minute {
		return c.onStorage(evt, duration)
	} else {
		return c.onPub(raw, duration)
	}
}

func NewClient(name string) *client {
	return &client{
		name: name,
		onPub: func(body []byte, duration time.Duration) error {
			return errors.New("pub func not set")
		},
		onStorage: func(evt event, duration time.Duration) error {
			return errors.New("storage func not set")
		},
	}
}
