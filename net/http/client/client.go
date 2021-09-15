package main

import (
	"bytes"
	"context"
	"fmt"
	"net/http"

	"io/ioutil"
	"net/url"
	"strings"
	"time"

	"github.com/whatap/go-api/httpc"
	"github.com/whatap/go-api/trace"
)

func httpGet(callUrl string) (int, error) {
	fmt.Println("httpGet ", callUrl)
	// GET 호출
	if resp, err := http.Get(callUrl); err == nil {
		defer resp.Body.Close()
		fmt.Println("status=", resp.StatusCode)
		// 결과 출력
		if data, err := ioutil.ReadAll(resp.Body); err == nil {
			fmt.Printf("%s\n", string(data))
		} else {
			fmt.Println(err)
		}
		return resp.StatusCode, err
	} else {
		fmt.Println(err)
		return -1, err
	}

}

func httpPost(callUrl, body string) (int, error) {
	fmt.Println("httpPost ", callUrl, ", ", body)
	reqBody := bytes.NewBufferString(body)
	if resp, err := http.Post(callUrl, "text/plain", reqBody); err == nil {
		defer resp.Body.Close()
		fmt.Println("status=", resp.StatusCode)
		// Response 체크.
		if data, err := ioutil.ReadAll(resp.Body); err == nil {
			fmt.Printf("%s\n", string(data))
		} else {
			fmt.Println(err)
		}
		return resp.StatusCode, err
	} else {
		fmt.Println(err)
		return -1, err
	}

}

func httpWithRequest(method string, callUrl string, body string, headers map[string]string) (int, error) {
	fmt.Println("httpGetWithRequest ", method, ", ", callUrl, ", ", body, ", ", headers)
	timeout := time.Duration(10 * time.Second)
	client := http.Client{
		Timeout: timeout,
	}

	if req, err := http.NewRequest(strings.ToUpper(method), callUrl, bytes.NewBufferString(body)); err == nil {
		if headers != nil {
			for key, val := range headers {
				req.Header.Add(key, val)
			}
		}
		if resp, err := client.Do(req); err == nil {
			defer resp.Body.Close()
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
	trace.Init(nil)
	//It must be executed before closing the app.
	defer trace.Shutdown()
	ctx := context.Background()

	callUrl := "https://www.google.com"
	wCtx, _ := trace.Start(ctx, "Http call")
	defer trace.End(wCtx, nil)

	mCtx, _ := httpc.Start(wCtx, callUrl)
	if statusCode, err := httpGet(callUrl); err == nil {
		httpc.End(mCtx, statusCode, "", nil)
	} else {
		httpc.End(mCtx, -1, "", err)
	}

	mCtx, _ = httpc.Start(wCtx, callUrl)
	if statusCode, err := httpPost(callUrl, ""); err == nil {
		httpc.End(mCtx, statusCode, "", nil)
	} else {
		httpc.End(mCtx, -1, "", err)
	}

	mCtx, _ = httpc.Start(wCtx, callUrl)
	if statusCode, err := httpWithRequest("GET", callUrl, "body", nil); err == nil {
		httpc.End(mCtx, statusCode, "", nil)
	} else {
		httpc.End(mCtx, -1, "", err)
	}

	mCtx, _ = httpc.Start(wCtx, callUrl)
	if statusCode, err := httpWithRequest("POST", callUrl, "body", nil); err == nil {
		httpc.End(mCtx, statusCode, "", nil)
	} else {
		httpc.End(mCtx, -1, "", err)
	}

	mCtx, _ = httpc.Start(wCtx, callUrl)
	if statusCode, err := httpPostForm(callUrl, ""); err == nil {
		httpc.End(mCtx, statusCode, "", nil)
	} else {
		httpc.End(mCtx, -1, "", err)
	}
	fmt.Println("Exit")
}
