// Copyright 2022 YBCZ, Inc. All rights reserved.
//
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file in the root of the source
// tree.

package mqms

type TraceStatus string

const (
	TraceStatusEmit      = "emit"
	TraceStatusError     = "error"
	TraceStatusExecBegin = "exec_begin"
	TraceStatusExecEnd   = "exec_end"
)
