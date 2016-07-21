package main

import (
	"github.com/zenazn/goji/web"
	"net/http"
	"strings"
)

// Disable cache and allow CORS requests in HTTP response headers. The CORS header additions are recommended by Swagger.
func NoCacheAllowCORS(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "must-revalidate")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, PUT, PATCH, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
}

type autoOptionsState int

const (
	aosInit autoOptionsState = iota
	aosHeaderWritten
	aosProxying
)

type CORSOptionsProxy struct {
	w     http.ResponseWriter
	c     *web.C
	state autoOptionsState
}

func (p *CORSOptionsProxy) Header() http.Header {
	return p.w.Header()
}

func (p *CORSOptionsProxy) Write(buf []byte) (int, error) {
	switch p.state {
	case aosInit:
		p.state = aosHeaderWritten
	case aosProxying:
		return len(buf), nil
	}
	return p.w.Write(buf)
}

func (p *CORSOptionsProxy) WriteHeader(code int) {
	methods := getValidMethods(*p.c)
	switch p.state {
	case aosInit:
		if methods != nil && code == http.StatusNotFound {
			p.state = aosProxying
			break
		}
		p.state = aosHeaderWritten
		fallthrough
	default:
		p.w.WriteHeader(code)
		return
	}

	hasOptions := false
	for _, m := range methods {
		if m == "OPTIONS" {
			hasOptions = true
			break
		}
	}
	if !hasOptions {
		methods = append(methods, "OPTIONS")
	}

	p.w.Header().Set("Allow", strings.Join(methods, ", "))
	NoCacheAllowCORS(p.w)
	p.w.WriteHeader(http.StatusOK)
}

// AutomaticCORSOptions automatically returns Allow header.
func AutomaticCORSOptions(c *web.C, h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "OPTIONS" {
			w = &CORSOptionsProxy{c: c, w: w}
		}
		h.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

func getValidMethods(c web.C) []string {
	if c.Env == nil {
		return nil
	}
	v, ok := c.Env[web.ValidMethodsKey]
	if !ok {
		return nil
	}
	if methods, ok := v.([]string); ok {
		return methods
	}
	return nil
}
