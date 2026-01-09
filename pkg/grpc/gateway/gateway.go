// Package gateway ...
package gateway

import (
	"net/http"
	"reflect"
	"runtime"
	"strings"

	gwruntime "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/hb-chen/opskills/pkg/logger"
)

// HTTPMiddlewareFunc ...
type HTTPMiddlewareFunc func(http.HandlerFunc) http.HandlerFunc

// Gateway ...
type Gateway struct {
	mux           *gwruntime.ServeMux
	premiddleware []HTTPMiddlewareFunc
	middleware    []HTTPMiddlewareFunc
}

// New ...
func New(opts ...gwruntime.ServeMuxOption) *Gateway {
	mux := gwruntime.NewServeMux(opts...)
	return &Gateway{mux: mux}
}

// Mux ...
func (gw *Gateway) Mux() *gwruntime.ServeMux {
	return gw.mux
}

func applyMiddleware(h http.HandlerFunc, middleware ...HTTPMiddlewareFunc) http.HandlerFunc {
	for i := len(middleware) - 1; i >= 0; i-- {
		h = middleware[i](h)
	}
	return h
}

// ServeHTTP ...
func (gw *Gateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Access log will be handled by middleware in server.go
	h := func(w http.ResponseWriter, r *http.Request) {
		gw.mux.ServeHTTP(w, r)
	}

	if gw.premiddleware == nil {
		h = applyMiddleware(h, gw.middleware...)
	} else {
		h = applyMiddleware(h, gw.premiddleware...)
		h = applyMiddleware(h, gw.middleware...)
	}

	h(w, r)
}

// Use adds middleware to the chain which is run after router.
func (gw *Gateway) Use(middleware ...HTTPMiddlewareFunc) {
	gw.middleware = append(gw.middleware, middleware...)
}

// Pre adds middleware to the chain which is run before router.
func (gw *Gateway) Pre(middleware ...HTTPMiddlewareFunc) {
	gw.premiddleware = append(gw.premiddleware, middleware...)
}

// MiddlewareFunc ...
type MiddlewareFunc func(gwruntime.HandlerFunc) gwruntime.HandlerFunc

// Group ...
func (gw *Gateway) Group(prefix string, m ...MiddlewareFunc) (g *Group) {
	g = &Group{
		prefix: prefix,
		gw:     gw,
	}
	g.Use(m...)
	return
}

// Group ...
type Group struct {
	prefix    string
	gw        *Gateway
	middleware []MiddlewareFunc
}

// Mux ...
func (g *Group) Mux() *gwruntime.ServeMux {
	return g.gw.mux
}

// Use ...
func (g *Group) Use(middleware ...MiddlewareFunc) {
	g.middleware = append(g.middleware, middleware...)
}

// GET ...
func (g *Group) GET(path string, h gwruntime.HandlerFunc, m ...MiddlewareFunc) {
	g.add("GET", path, h, append(g.middleware, m...)...)
}

// POST ...
func (g *Group) POST(path string, h gwruntime.HandlerFunc, m ...MiddlewareFunc) {
	g.add("POST", path, h, append(g.middleware, m...)...)
}

// PUT ...
func (g *Group) PUT(path string, h gwruntime.HandlerFunc, m ...MiddlewareFunc) {
	g.add("PUT", path, h, append(g.middleware, m...)...)
}

// DELETE ...
func (g *Group) DELETE(path string, h gwruntime.HandlerFunc, m ...MiddlewareFunc) {
	g.add("DELETE", path, h, append(g.middleware, m...)...)
}

// PATCH ...
func (g *Group) PATCH(path string, h gwruntime.HandlerFunc, m ...MiddlewareFunc) {
	g.add("PATCH", path, h, append(g.middleware, m...)...)
}

func (g *Group) add(method, path string, h gwruntime.HandlerFunc, middleware ...MiddlewareFunc) {
	// Chain middleware
	for i := len(middleware) - 1; i >= 0; i-- {
		h = middleware[i](h)
	}
	if err := g.gw.mux.HandlePath(method, g.prefix+path, h); err != nil {
		logger.Fatalf("Failed to register route: %v", err)
	}
}

func handlerName(h interface{}) string {
	f := runtime.FuncForPC(reflect.ValueOf(h).Pointer())
	name := f.Name()
	if i := strings.LastIndex(name, "."); i > 0 {
		name = name[i+1:]
	}
	return name
}

