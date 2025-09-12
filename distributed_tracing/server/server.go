package server

import (
	"bytes"
	"context"

	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/whatap/go-api/httpc"
	"github.com/whatap/go-api/trace"
)

type Server struct {
}

func (s *Server) Start(port int, apiHost string, depth int) {
	// Use trace.StartWithRequest
	http.HandleFunc(fmt.Sprintf("%s_%d", "/trace1", depth), func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(r)
		// Inside StartWithRequest
		// use function trace.UpdateMtrace(http.Header)
		// The headers received in the request are analyzed to generate the headers required for distributed tracing.
		// Distributed tracing headers include W3C's traceparent header and Wattap's x-wtp-xxx.

		// whatapCtx is a context that contains TraceContext information used by whatap with the name "whatap" inside the context.
		ctx, _ := trace.StartWithRequest(r)
		defer trace.End(ctx, nil)

		w.Header().Add("Content-Type", "application/json")
		var buffer bytes.Buffer
		buffer.WriteString("{json:{message:'trace1'}}")

		// There is no need to execute the function.
		// _, traceCtx := trace.GetTraceContext(whatapCtx)
		// trace.UpdateMtrace(traceCtx, r.Header)

		// Get additional WhaTap headers .
		wHeader := trace.GetMtrace(ctx)

		// External API Request, set whatap header
		callUrl, _ := url.JoinPath(apiHost, fmt.Sprintf("%s_%d", "trace1", depth+1))
		timeout := time.Duration(10 * time.Second)
		client := http.Client{
			Timeout: timeout,
		}
		req, err := http.NewRequest("GET", callUrl, bytes.NewBufferString("call"))
		if err != nil {
			return
		}

		// Set additional WhaTap headers .
		if wHeader != nil {
			for key, _ := range wHeader {
				req.Header.Add(key, wHeader.Get(key))
			}
		}

		// Trace external api request
		whatapHttpcCtx, _ := httpc.Start(ctx, callUrl)
		if resp, err := client.Do(req); err == nil {
			httpc.End(whatapHttpcCtx, resp.StatusCode, "", nil)
			defer resp.Body.Close()
		} else {
			httpc.End(whatapHttpcCtx, -1, "", err)
		}
	})

	// Use trace.Start
	http.HandleFunc(fmt.Sprintf("%s_%d", "/trace2", depth), func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(r)
		ctx := context.Background()
		//ctx := r.Context()

		// Start() cannot parse request header information because there is no http.Request
		// whatapCtx is a context that contains TraceContext information used by whatap with the name "whatap" inside the context.
		ctx, _ = trace.Start(ctx, fmt.Sprintf("%s_%d", "/trace2", depth))
		defer trace.End(ctx, nil)

		w.Header().Add("Content-Type", "application/json")
		var buffer bytes.Buffer
		buffer.WriteString("{json:{message:'/trace2'}}")

		// use function trace.UpdateMtrace(http.Header)
		// The headers received in the request are analyzed to generate the headers required for distributed tracing.
		// Distributed tracing headers include W3C's traceparent header and Wattap's x-wtp-xxx.
		trace.UpdateMtraceWithContext(ctx, r.Header)

		// Get additional WhaTap headers .
		wHeader := trace.GetMtrace(ctx)

		// External API Request
		callUrl, _ := url.JoinPath(apiHost, fmt.Sprintf("%s_%d", "trace2", depth+1))
		timeout := time.Duration(10 * time.Second)
		client := http.Client{
			Timeout: timeout,
		}
		req, err := http.NewRequest("GET", callUrl, bytes.NewBufferString("call"))
		if err != nil {
			return
		}

		// Set additional WhaTap headers. From trace.GetMtrace
		if wHeader != nil {
			for key, _ := range wHeader {
				req.Header.Add(key, wHeader.Get(key))
			}
		}

		// Trace external api request
		whatapHttpcCtx, _ := httpc.Start(ctx, callUrl)
		if resp, err := client.Do(req); err == nil {
			httpc.End(whatapHttpcCtx, resp.StatusCode, "", nil)
			defer resp.Body.Close()
		} else {
			httpc.End(whatapHttpcCtx, -1, "", err)
		}
	})

	// Use trace.StartWithContext
	http.HandleFunc(fmt.Sprintf("%s_%d", "/trace3", depth), func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(r)

		//ctx := context.Background()
		ctx := r.Context()

		// whatapCtx is a context that contains TraceContext information used by whatap with the name "whatap" inside the context.
		ctx, _ = trace.NewTraceContext(ctx)

		// StartWithContext() cannot parse request header information because there is no http.Request
		// StartWithContext() requires an already created whatapCtx.
		ctx, _ = trace.StartWithContext(ctx, fmt.Sprintf("%s_%d", "/trace3", depth))
		defer trace.End(ctx, nil)

		w.Header().Add("Content-Type", "application/json")
		var buffer bytes.Buffer
		buffer.WriteString("{json:{message:'/trace3'}}")

		// use function trace.UpdateMtrace(http.Header)
		// The headers received in the request are analyzed to generate the headers required for distributed tracing.
		// Distributed tracing headers include W3C's traceparent header and Wattap's x-wtp-xxx.
		trace.UpdateMtraceWithContext(ctx, r.Header)

		// Get additional WhaTap headers.
		wHeader := trace.GetMtrace(ctx)

		// External API Request
		callUrl, _ := url.JoinPath(apiHost, fmt.Sprintf("%s_%d", "trace3", depth+1))
		timeout := time.Duration(10 * time.Second)
		client := http.Client{
			Timeout: timeout,
		}
		req, err := http.NewRequest("GET", callUrl, bytes.NewBufferString("call"))
		if err != nil {
			return
		}

		// Set additional WhaTap headers. . From trace.GetMtrace
		if wHeader != nil {
			for key, _ := range wHeader {
				req.Header.Add(key, wHeader.Get(key))
			}
		}

		// Trace external api request
		httpcCtx, _ := httpc.Start(ctx, callUrl)
		if resp, err := client.Do(req); err == nil {
			httpc.End(httpcCtx, resp.StatusCode, "", nil)
			defer resp.Body.Close()
		} else {
			httpc.End(httpcCtx, -1, "", err)
		}
	})
	fmt.Println("Start :", port)

	_ = http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}
