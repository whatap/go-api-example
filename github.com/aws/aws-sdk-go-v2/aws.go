package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/whatap/go-api/instrumentation/github.com/aws/aws-sdk-go-v2/whatapaws"
	"github.com/whatap/go-api/method"
	"github.com/whatap/go-api/trace"

	_ "net/http/pprof"
)

type HTMLData struct {
	Title   string
	Content string
	//HTMLContent template.HTML
}

type HTMLLink struct {
	Href   string
	Name   string
	Alt    string
	Target string
}
type IndexHTMLData struct {
	HTMLData
	ALink []HTMLLink
}

type testAPI = func(context.Context, http.ResponseWriter, *http.Request)
type httpHandler = func(http.ResponseWriter, *http.Request)

var GlobalConfig aws.Config

func traceHandler(ctxGetter func(*http.Request) context.Context, api testAPI) httpHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := ctxGetter(r)
		api(ctx, w, r)
	}
}

func requestCtx(r *http.Request) context.Context {
	return r.Context()
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
		fmt.Println("Error ListS3 ", err)
		return
	}

	ret := []string{}
	for _, object := range output.Contents {
		ret = append(ret, *object.Key)
	}

	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)
	encoder.Encode(ret)
}

type EC2DescribeInstancesAPI interface {
	DescribeInstances(ctx context.Context,
		params *ec2.DescribeInstancesInput,
		optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error)
}

func GetInstances(c context.Context, api EC2DescribeInstancesAPI, input *ec2.DescribeInstancesInput) (*ec2.DescribeInstancesOutput, error) {
	return api.DescribeInstances(c, input)
}

func DescribeEC2(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	client := ec2.NewFromConfig(GlobalConfig)

	input := &ec2.DescribeInstancesInput{}

	result, err := GetInstances(context.TODO(), client, input)
	if err != nil {
		return
	}

	ret := []string{}
	for _, r := range result.Reservations {
		for _, i := range r.Instances {
			ret = append(ret, *i.InstanceId)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)
	encoder.Encode(ret)
}

func getUser(ctx context.Context) {
	methodCtx, _ := method.Start(ctx, "getUser")
	defer method.End(methodCtx, nil)
	time.Sleep(time.Duration(1) * time.Second)
}

func main() {
	portPtr := flag.Int("p", 8080, "web port. default 8080  ")
	udpHostPtr := flag.String("uh", "127.0.0.1", "agent host(udp). defalt 127.0.0.1 ")
	udpPortPtr := flag.Int("up", 6600, "agent port(udp). defalt 6600 ")
	whatapEnabeld := flag.String("whatap.enabled", "true", " whatap enabled ")
	whatapDebug := flag.String("whatap.debug", "false", " whatap debug flag")
	// dataSourcePtr := flag.String("ds", "doremimaker:doremimaker@tcp(phpdemo2:3306)/doremimaker", " dataSourceName ")
	flag.Parse()
	port := *portPtr
	udpHost := *udpHostPtr
	udpPort := *udpPortPtr
	whatapConfigEnabled := *whatapEnabeld
	whatapConfigDebug := *whatapDebug
	// dataSource := *dataSourcePtr

	whatapConfig := make(map[string]string)
	whatapConfig["net_udp_host"] = udpHost
	whatapConfig["net_udp_port"] = fmt.Sprintf("%d", udpPort)
	whatapConfig["enabled"] = whatapConfigEnabled
	whatapConfig["debug"] = whatapConfigDebug

	trace.Init(whatapConfig)
	defer trace.Shutdown()

	// Load the Shared AWS Configuration (~/.aws/config)
	config, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		fmt.Println("Error LoadDefaultConfig ", err)
	}
	GlobalConfig = whatapaws.AppendMiddleware(config)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		ctx, _ := trace.StartWithRequest(r)
		defer trace.End(ctx, nil)
		fmt.Println("Request -", r)

		tp, err := template.ParseFiles("templates/github.com/aws/aws-sdk-go-v2/index.html")
		if err != nil {
			fmt.Println("Template not loaded, ", err)
			return
		}
		data := &HTMLData{}
		data.Title = "aws-sdk-go-v2"
		data.Content = r.RequestURI
		tp.Execute(w, data)

		trace.Step(ctx, "Text Message", "Message", 3, 3)

		getUser(ctx)
		fmt.Println("Response -", r.Response)
	})

	http.HandleFunc("/ListS3", trace.HandlerFunc(traceHandler(requestCtx, ListS3)))
	http.HandleFunc("/SleepAndListS3", trace.HandlerFunc(traceHandler(sleepRequestCtx, ListS3)))

	http.HandleFunc("/DescribeEC2", trace.HandlerFunc(traceHandler(requestCtx, DescribeEC2)))
	http.HandleFunc("/SleepAndDescribeEC2", trace.HandlerFunc(traceHandler(sleepRequestCtx, DescribeEC2)))

	// http.HandleFunc("/debug/pprof/", pprof.Index) // Profile Endpoint for Heap, Block, ThreadCreate, Goroutine, Mutex
	// http.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	// http.HandleFunc("/debug/pprof/profile", pprof.Profile) // Profile Endpoint for CPU
	// http.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	// http.HandleFunc("/debug/pprof/trace", pprof.Trace)

	fmt.Println("Start :", port, ", Agent Udp Port:", udpPort)
	_ = http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}
