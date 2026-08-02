// SPDX-FileCopyrightText: 2026 Playground Logic LLC
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/provabl/qualify/internal/auth"
)

// chi's middleware.RealIP is deprecated as unfixably spoofable: it rewrites
// r.RemoteAddr from the leftmost X-Forwarded-For (or True-Client-IP /
// X-Real-IP) value whether or not a trusted proxy set it
// (GHSA-3fxj-6jh8-hvhx / -rjr7-jggh-pgcp / -9g5q-2w5x-hmxf). Because the
// request log is operator evidence for a compliance tool, a forgeable origin
// makes it unreliable. These tests guard the removal — bumping chi does NOT fix
// this, so only keeping the middleware out does.

// TestRouterOmitsRealIP inspects the real production middleware chain by
// function pointer, so re-adding middleware.RealIP fails the build's tests even
// if no request-level assertion happens to cover it.
func TestRouterOmitsRealIP(t *testing.T) {
	h := setupRouter(auth.Config{}, nil, nil)

	mux, ok := h.(*chi.Mux)
	if !ok {
		t.Fatalf("setupRouter returned %T, expected *chi.Mux — update this test", h)
	}

	// Resolve each middleware to its function name rather than comparing against
	// middleware.RealIP directly — naming the deprecated symbol would itself trip
	// staticcheck's SA1019, and this keeps the repo free of any RealIP reference.
	for i, mw := range mux.Middlewares() {
		fn := runtime.FuncForPC(reflect.ValueOf(mw).Pointer())
		if fn == nil {
			continue
		}
		if strings.HasSuffix(fn.Name(), "middleware.RealIP") {
			t.Errorf("middleware[%d] is chi middleware.RealIP: it lets any client forge "+
				"r.RemoteAddr via X-Forwarded-For, corrupting the request log. To honor a "+
				"proxy header, parse it against a configured trusted-proxy list instead.", i)
		}
	}
}

// TestForwardedForDoesNotOverrideRemoteAddr is the behavioral half: a request
// carrying spoofed client-IP headers must reach the handler with r.RemoteAddr
// still set to the true TCP peer.
func TestForwardedForDoesNotOverrideRemoteAddr(t *testing.T) {
	const truePeer = "192.0.2.10:54321"

	var got string
	mux := setupRouter(auth.Config{}, nil, nil).(*chi.Mux)
	mux.Get("/__remote_addr_probe", func(_ http.ResponseWriter, r *http.Request) {
		got = r.RemoteAddr
	})

	for _, hdr := range []string{"X-Forwarded-For", "X-Real-IP", "True-Client-IP"} {
		t.Run(hdr, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/__remote_addr_probe", nil)
			req.RemoteAddr = truePeer
			req.Header.Set(hdr, "203.0.113.99")

			mux.ServeHTTP(httptest.NewRecorder(), req)

			if got != truePeer {
				t.Errorf("%s spoofed the logged client address: RemoteAddr = %q, want the true peer %q",
					hdr, got, truePeer)
			}
		})
	}
}
