// net/http Client Instrumentation Example
//
// This example demonstrates how to instrument HTTP client applications
// (non-server, standalone programs) using WhaTap Go API.
//
// INSTRUMENTATION OVERVIEW:
// For standalone programs without HTTP server, use:
// - trace.Start(context.Background(), "name") to start a transaction
// - trace.End(ctx, nil) to end the transaction
//
// HTTP CLIENT INSTRUMENTATION:
// There are TWO methods to track outgoing HTTP calls:
//
// METHOD 1: Manual httpc.Start/End (explicit control)
//   - Call httpc.Start(ctx, url) before the call
//   - Call httpc.End(httpcCtx, statusCode, "", err) after the call
//   - Use trace.GetMTrace(ctx) to propagate trace headers for distributed tracing
//
// METHOD 2: whataphttp.NewRoundTrip() (recommended)
//   - Wrap http.Client.Transport with whataphttp.NewRoundTrip()
//   - Automatically tracks all HTTP calls made by the client
//   - Automatically propagates trace headers for distributed tracing
//
// FEATURES:
// - Transaction tracking for standalone programs
// - HTTP client call tracking (GET, POST, etc.)
// - Distributed tracing via trace.GetMTrace()
// - Status code and error tracking

package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"net/http"

	"io/ioutil"
	"net/url"
	"strings"
	"time"

	"github.com/whatap/go-api/httpc"
	"github.com/whatap/go-api/instrumentation/net/http/whataphttp"
	"github.com/whatap/go-api/trace"
)

// Simple HTTP GET without instrumentation
func httpGet(callUrl string) (int, error) {
	fmt.Println("httpGet ", callUrl)
	// GET 호출
	if resp, err := http.Get(callUrl); err == nil {
		defer resp.Body.Close()
		fmt.Println("status=", resp.StatusCode)
		// 결과 출력
		//if data, err := ioutil.ReadAll(resp.Body); err == nil {
		if _, err := ioutil.ReadAll(resp.Body); err == nil {
			//fmt.Printf("%s\n", string(data))
		} else {
			fmt.Println(err)
		}
		return resp.StatusCode, err
	} else {
		fmt.Println(err)
		return -1, err
	}

}

// Simple HTTP POST without instrumentation
func httpPost(callUrl, body string) (int, error) {
	fmt.Println("httpPost ", callUrl, ", ", body)
	reqBody := bytes.NewBufferString(body)
	if resp, err := http.Post(callUrl, "text/plain", reqBody); err == nil {
		defer resp.Body.Close()
		fmt.Println("status=", resp.StatusCode)
		// Response 체크.
		//if data, err := ioutil.ReadAll(resp.Body); err == nil {
		if _, err := ioutil.ReadAll(resp.Body); err == nil {
			//fmt.Printf("%s\n", string(data))
		} else {
			fmt.Println(err)
		}
		return resp.StatusCode, err
	} else {
		fmt.Println(err)
		return -1, err
	}

}

// HTTP request with custom headers - used with httpc.Start/End for tracking
// headers parameter can include trace.GetMTrace(ctx) for distributed tracing
func httpWithRequest(method string, callUrl string, body string, headers http.Header) (int, error) {
	fmt.Println("httpGetWithRequest ", method, ", ", callUrl, ", ", body, ", ", headers)
	timeout := time.Duration(10 * time.Second)
	client := http.Client{
		Timeout: timeout,
	}

	if req, err := http.NewRequest(strings.ToUpper(method), callUrl, bytes.NewBufferString(body)); err == nil {
		// Add custom headers including trace headers for distributed tracing
		if headers != nil {
			for key, _ := range headers {
				req.Header.Add(key, headers.Get(key))
			}
		}
		if resp, err := client.Do(req); err == nil {
			defer resp.Body.Close()
			//fmt.Println("status=", resp.StatusCode)
			fmt.Println("status=", resp.StatusCode)
			return resp.StatusCode, err
		} else {
			fmt.Println(err)
			return -2, err
		}

	} else {
		fmt.Println(err)
		return -1, err
	}
}

// HTTP POST form without instrumentation
func httpPostForm(callUrl, params string) (int, error) {
	fmt.Println("httpPostForm ", callUrl, ", ", params)
	var urlValues url.Values = url.Values{}
	kv := strings.Split(params, "&")
	if params != "" {
		for _, v := range kv {
			if v != "" {
				k, v := getKV(v, "=")
				urlValues.Set(k, v)
			}
		}
	}

	if resp, err := http.PostForm(callUrl, urlValues); err == nil {
		defer resp.Body.Close()
		fmt.Println("status=", resp.StatusCode)
		return resp.StatusCode, err
	} else {
		fmt.Println(err)
		return -1, err
	}

}
func getKV(str, div string) (string, string) {
	kv := strings.Split(str, div)
	return kv[0], kv[1]
}
func main() {
	udpPortPtr := flag.Int("up", 6600, "part ")
	setWhatapPtr := flag.Bool("whatap", false, "set whatap")

	flag.Parse()
	udpPort := *udpPortPtr
	IsWhatap := *setWhatapPtr

	// ============================================================
	// INSTRUMENTATION STEP 1: Initialize WhaTap Agent
	// ============================================================
	// Call trace.Init() at application startup.
	// Configuration includes mtrace settings for distributed tracing.
	if IsWhatap {
		config := make(map[string]string)
		config["net_udp_port"] = fmt.Sprintf("%d", udpPort)
		config["mtrace_enabled"] = "true"  // Enable multi-transaction tracing
		config["mtrace_rate"] = "100"      // 100% sampling rate
		trace.Init(config)
	}
	defer trace.Shutdown()

	// ============================================================
	// Start Transaction for Standalone Program
	// ============================================================
	// For non-server programs, use trace.Start() to create a transaction.
	// This is different from server programs which use trace.StartWithRequest().
	//
	// BEFORE (No instrumentation):
	//   // Just make HTTP calls directly
	//   http.Get(url)
	//
	// AFTER (Instrumented):
	//   ctx, _ := trace.Start(context.Background(), "Transaction Name")
	//   defer trace.End(ctx, nil)
	//   // Make HTTP calls with context
	ctx, _ := trace.Start(context.Background(), "Http call")
	defer trace.End(ctx, nil)

	// ============================================================
	// METHOD 1: httpc.Start/End for HTTP Client Tracking
	// ============================================================
	// Use httpc.Start/End for explicit control over HTTP client tracking.
	// This method requires manual management of start/end calls.
	//
	// Flow:
	// 1. httpcCtx, _ := httpc.Start(ctx, url) - Start tracking
	// 2. Make HTTP call
	// 3. httpc.End(httpcCtx, statusCode, "", err) - End tracking
	callUrl := "https://www.google.com"

	// Example 1: Simple GET with httpc tracking
	httpcCtx, _ := httpc.Start(ctx, callUrl)
	if statusCode, err := httpGet(callUrl); err == nil {
		httpc.End(httpcCtx, statusCode, "", nil)
	} else {
		httpc.End(httpcCtx, -1, "", err)
	}

	// Example 2: Simple POST with httpc tracking
	httpcCtx, _ = httpc.Start(ctx, callUrl)
	if statusCode, err := httpPost(callUrl, ""); err == nil {
		httpc.End(httpcCtx, statusCode, "", nil)
	} else {
		httpc.End(httpcCtx, -1, "", err)
	}

	// ============================================================
	// Distributed Tracing with trace.GetMTrace()
	// ============================================================
	// Use trace.GetMTrace(ctx) to get HTTP headers for distributed tracing.
	// Pass these headers to the downstream service to link transactions.
	//
	// Headers included:
	// - x-whatap-mtid: Multi-transaction ID
	// - x-whatap-traceid: Trace ID
	// - x-whatap-poid: Parent OID
	callUrl = "http://localhost:8081/httpc"

	// Example 3: GET with distributed tracing headers
	httpcCtx, _ = httpc.Start(ctx, callUrl)
	if statusCode, err := httpWithRequest("GET", callUrl, "body", trace.GetMTrace(ctx)); err == nil {
		httpc.End(httpcCtx, statusCode, "", nil)
	} else {
		httpc.End(httpcCtx, -1, "", err)
	}

	// Example 4: POST with distributed tracing headers
	httpcCtx, _ = httpc.Start(ctx, callUrl)
	if statusCode, err := httpWithRequest("POST", callUrl, "body", trace.GetMTrace(ctx)); err == nil {
		httpc.End(httpcCtx, statusCode, "", nil)
	} else {
		httpc.End(httpcCtx, -1, "", err)
	}

	// Example 5: POST form with httpc tracking
	httpcCtx, _ = httpc.Start(ctx, callUrl)
	if statusCode, err := httpPostForm(callUrl, ""); err == nil {
		httpc.End(httpcCtx, statusCode, "", nil)
	} else {
		httpc.End(httpcCtx, -1, "", err)
	}

	// ============================================================
	// METHOD 2: whataphttp.NewRoundTrip() (Recommended)
	// ============================================================
	// Use whataphttp.NewRoundTrip() to wrap http.Client.Transport.
	// This automatically:
	// - Tracks all HTTP calls made by the client
	// - Propagates trace headers for distributed tracing
	// - Records status codes and errors
	//
	// BEFORE (Original):
	//   client := http.DefaultClient
	//   resp, err := client.Get(url)
	//
	// AFTER (Instrumented):
	//   client := http.DefaultClient
	//   client.Transport = whataphttp.NewRoundTrip(ctx, http.DefaultTransport)
	//   resp, err := client.Get(url)
	callUrl = "http://localhost:8081/httpc"
	client := http.DefaultClient
	// Wrap transport for automatic HTTP client tracking
	client.Transport = whataphttp.NewRoundTrip(ctx, http.DefaultTransport)
	resp, err := client.Get(callUrl)
	if err != nil {
		fmt.Printf("Error %s", err.Error())
	}
	defer resp.Body.Close()

	fmt.Println("Exit")
}
