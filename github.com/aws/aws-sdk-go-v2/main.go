package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/sirupsen/logrus"
	whatapaws "github.com/whatap/go-api/instrumentation/github.com/aws/aws-sdk-go-v2"
	"github.com/whatap/go-api/trace"
)

type testAPI = func(context.Context, http.ResponseWriter, *http.Request)
type httpHandler = func(http.ResponseWriter, *http.Request)

var GlobalConfig aws.Config

func traceHandler(ctxGetter func(*http.Request) context.Context, api testAPI) httpHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := ctxGetter(r)
		api(ctx, w, r)
		trace.End(ctx, nil)
	}
}

func requestCtx(r *http.Request) context.Context {
	ret, err := trace.StartWithRequest(r)
	if err != nil {
		return context.TODO()
	}
	return ret
}

func sleepRequestCtx(r *http.Request) context.Context {
	ret, err := trace.StartWithRequest(r)
	time.Sleep(time.Second)
	if err != nil {
		return context.TODO()
	}
	return ret
}

func ListS3(ctx context.Context, w http.ResponseWriter, r *http.Request) {

	client := s3.NewFromConfig(GlobalConfig)
	output, err := client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String("dev-default-region"),
	})
	if err != nil {
		logrus.Panic(err)
	}

	ret := []string{}
	for _, object := range output.Contents {
		ret = append(ret, *object.Key)
	}

	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)
	encoder.Encode(ret)
}

func main() {

	port := "9090"
	udpPort := "6600"
	whatapConfig := map[string]string{
		"net_udp_port":              udpPort,
		"debug":                     "true",
		"profile_sql_param_enabled": "true",
	}
	trace.Init(whatapConfig)
	defer trace.Shutdown()

	// Load the Shared AWS Configuration (~/.aws/config)
	config, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		logrus.Panic()
	}
	GlobalConfig = whatapaws.AppendMiddleware(config)

	http.HandleFunc("/ListS3", traceHandler(requestCtx, ListS3))
	http.HandleFunc("/SleepAndListS3", traceHandler(sleepRequestCtx, ListS3))

	http.HandleFunc("/DescribeEC2", traceHandler(requestCtx, DescribeEC2))
	http.HandleFunc("/SleepAndDescribeEC2", traceHandler(sleepRequestCtx, DescribeEC2))

	_ = http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
}
