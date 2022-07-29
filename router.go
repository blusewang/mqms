// Copyright 2022 YBCZ, Inc. All rights reserved.
//
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file in the root of the source
// tree.

package mqms

import (
	"fmt"
	"reflect"
	"runtime"
)

type HandlerFunc func(*Context) error

type IRouter interface {
	Use(...HandlerFunc) IRouter
	Group(string) *Route
	Item(string, HandlerFunc) *Route
}

type Route struct {
	functions []HandlerFunc
	basePath  string
	engine    *Engine
}

func (r *Route) Use(handlers ...HandlerFunc) IRouter {
	r.functions = handlers
	return r
}

func (r *Route) Group(name string) *Route {
	return &Route{
		functions: r.functions,
		basePath:  r.combineRoute(name),
		engine:    r.engine,
	}
}

func (r *Route) Item(name string, handlerFunc HandlerFunc) *Route {
	uri := r.combineRoute(name)
	r.engine.routes[uri] = append(r.functions, handlerFunc)
	r.engine.handler.Log(fmt.Sprintf("[MQMS] %-25s --> %s (%d functions)\n", uri, nameOfFunction(handlerFunc), len(r.engine.routes[uri])))
	return &Route{
		functions: r.functions,
		basePath:  uri,
		engine:    r.engine,
	}
}

func (r *Route) combineRoute(name string) string {
	if r.basePath == "" {
		return name
	} else if name == "" {
		return r.basePath
	} else {
		return r.basePath + "." + name
	}
}

var _ IRouter = &Route{}

func nameOfFunction(f interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
}
