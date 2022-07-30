// Copyright 2022 YBCZ, Inc. All rights reserved.
//
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file in the root of the source
// tree.

package mqms

// TraceStatus 事件链路状态
type TraceStatus string

const (
	TraceStatusEmit      = "emit"       // 发布
	TraceStatusError     = "error"      // 错误
	TraceStatusExecBegin = "exec_begin" // 开始
	TraceStatusExecEnd   = "exec_end"   // 结束
)
