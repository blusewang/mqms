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
	Group(string) *Router
	Item(string, HandlerFunc) *Router
}

type Router struct {
	handlers []HandlerFunc
	basePath string
	engine   *Engine
}

func (r *Router) Use(handlers ...HandlerFunc) IRouter {
	r.handlers = handlers
	return r
}

func (r *Router) Group(name string) *Router {
	return &Router{
		handlers: r.handlers,
		basePath: r.combineRoute(name),
		engine:   r.engine,
	}
}

func (r *Router) Item(name string, handlerFunc HandlerFunc) *Router {
	uri := r.combineRoute(name)
	r.engine.routes[uri] = append(r.handlers, handlerFunc)
	r.engine.handler.Log(fmt.Sprintf("[MQMS] %-25s --> %s (%d handlers)\n", uri, nameOfFunction(handlerFunc), len(r.engine.routes[uri])))
	return &Router{
		handlers: r.handlers,
		basePath: uri,
		engine:   r.engine,
	}
}

func (r *Router) combineRoute(name string) string {
	if r.basePath == "" {
		return name
	} else if name == "" {
		return r.basePath
	} else {
		return r.basePath + "." + name
	}
}

var _ IRouter = &Router{}

func nameOfFunction(f interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
}
