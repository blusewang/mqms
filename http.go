// Copyright 2022 YBCZ, Inc. All rights reserved.
//
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file in the root of the source
// tree.

package mqms

import (
	"github.com/google/uuid"
	"net/http"
	"time"
)

// HttpTrace http 性能信息
type HttpTrace struct {
	EvtID            uuid.UUID `json:"evt_id"`
	Method           string    `json:"method"`
	Url              string    `json:"url"`
	BeginAt          time.Time `json:"begin_at"`
	EndAt            time.Time `json:"end_at"`
	ReqContentLength int64     `json:"req_content_length"`
	ResStatus        string    `json:"res_status"`
}

// 检测HTTP性能
type httpTransport struct {
	t      http.Transport
	engine *Engine
	evtID  uuid.UUID
}

func (m *httpTransport) RoundTrip(req *http.Request) (res *http.Response, err error) {
	ht := HttpTrace{
		EvtID:            m.evtID,
		Method:           req.Method,
		Url:              req.URL.String(),
		BeginAt:          time.Now(),
		ReqContentLength: req.ContentLength,
	}
	res, err = m.t.RoundTrip(req)
	ht.EndAt = time.Now()
	ht.ResStatus = res.Status
	// 忽略跳转跟踪
	if res.StatusCode >= 300 && res.StatusCode < 400 {
		return
	}
	m.engine.handler.HttpTrace(ht)
	return
}
