// Copyright 2022 YBCZ, Inc. All rights reserved.
//
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file in the root of the source
// tree.

package mqms

import (
	"net/http"
	"time"
)

// HttpTrace http 性能信息
type HttpTrace struct {
	Method           string
	Url              string
	BeginAt          time.Time
	EndAt            time.Time
	ReqContentLength int64
	ResStatus        string
}

// 检测HTTP性能
type httpTransport struct {
	t      http.Transport
	engine *Engine
}

func (m *httpTransport) RoundTrip(req *http.Request) (res *http.Response, err error) {
	ht := HttpTrace{
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

func client(e *Engine) http.Client {
	return http.Client{Transport: &httpTransport{http.Transport{}, e}}
}
