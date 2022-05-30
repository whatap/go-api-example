package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	pb "github.com/whatap/go-api-example/google.golang.org/grpc/proto"

	"github.com/whatap/go-api/instrumentation/google.golang.org/grpc/whatapgrpc"
	"github.com/whatap/go-api/instrumentation/net/http/whataphttp"
	"github.com/whatap/go-api/trace"

	// "google.golang.org/genproto"
	"google.golang.org/grpc"
)

var sendLock sync.Mutex

func RecvHealth(stream pb.ServiceNotice_HealthClient) error {
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			log.Println("Error EOF, End stream")
			return nil
		}
		if err != nil {
			log.Println("Error ", err)
			return err
		}
		log.Println("Response server health ", in.Status)
	}
}

func SendHealth(stream pb.ServiceNotice_HealthClient) error {
	for {
		req := &pb.ServerHealth{}
		req.Status = pb.STATUS_STATUS_READY
		req.CurrentMillis = time.Now().UnixNano() / int64(time.Millisecond)
		if err := sendHealth(req, stream); err != nil {
			log.Println("Errror send request ", err)
			return err
		}
		log.Println("Request server health, my status ", req.Status)

		time.Sleep(time.Second * 5)
	}
}
func sendHealth(req *pb.ServerHealth, stream pb.ServiceNotice_HealthClient) error {
	sendLock.Lock()
	defer sendLock.Unlock()
	if stream != nil {
		if err := stream.Send(req); err != nil {
			log.Println("Errror send request ", err)
			return err
		}
	}
	return nil
}

func recvHealth(stream pb.ServiceNotice_HealthClient) (*pb.ServerHealth, error) {
	return stream.Recv()
}

func parseParams(r *http.Request) []*pb.Param {
	rt := make([]*pb.Param, 0)
	r.ParseForm()
	if len(r.Form) > 0 {
		for k, v := range r.Form {
			tmp := &pb.Param{}
			tmp.Key = k
			if len(v) > 0 {
				tmp.Value = v[0]
			} else {
				tmp.Value = ""
			}
			rt = append(rt, tmp)
		}
	}
	return rt
}

type HTMLData struct {
	Title   string
	Content string
	//HTMLContent template.HTML
}

func main() {
	port := flag.Int("p", 8083, "web port. default 8080  ")
	grpcHost := flag.String("gh", "localhost", "grpc host. default localhost")
	grpcPort := flag.Int("gp", 8082, "grpc port. defalt 8082 ")
	udpPort := flag.Int("up", 6600, "agent port(udp). defalt 6600 ")

	flag.Parse()

	// Init WhaTap Trace
	config := make(map[string]string)
	config["net_udp_port"] = fmt.Sprintf("%d", *udpPort)
	trace.Init(config)
	defer trace.Shutdown()

	// Set the whatap interceptor to grpc
	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", *grpcHost, *grpcPort), grpc.WithInsecure(),
		grpc.WithBlock(),
		grpc.WithUnaryInterceptor(whatapgrpc.UnaryClientInterceptor()),
		grpc.WithStreamInterceptor(whatapgrpc.StreamClientInterceptor()))

	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	c := pb.NewServiceNoticeClient(conn)
	ctx := context.Background()
	//ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	//defer cancel()
	var globalStream pb.ServiceNotice_HealthClient

	mux := http.NewServeMux()

	mux.Handle("/health", whataphttp.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		//ctx := r.Context()
		w.Header().Add("Content-Type", "text/html")
		var buffer bytes.Buffer

		req := &pb.ServerHealth{}
		req.Status = pb.STATUS_STATUS_BUSY
		req.CurrentMillis = time.Now().UnixNano() / int64(time.Millisecond)
		if globalStream != nil {
			if err := sendHealth(req, globalStream); err != nil {
				buffer.WriteString(fmt.Sprintln("Request to grpc, my health ", req.Status, ",time=", req.CurrentMillis))
			}
		} else {
			buffer.WriteString("not open stream")
		}
		_, _ = w.Write(buffer.Bytes())

	}))

	mux.Handle("/startHealthStream", whataphttp.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if globalStream == nil {
			log.Println("globalStream nil")
		} else {
			if globalStream != nil {
				globalStream.CloseSend()
			}
		}
		// request 종료되면서 ctx cancel. Error  rpc error: code = Canceled desc = context canceled
		// 전역 ctx 를 넣어 주면 종료 안됨.
		ctx = r.Context()
		if tmp, err := c.Health(ctx); err != nil {
			log.Fatalf("Error stream : %v", err)
		} else {
			globalStream = tmp
		}

		log.Println("globalStream", globalStream)

		go SendHealth(globalStream)
		go RecvHealth(globalStream)

		//ctx := r.Context()
		w.Header().Add("Content-Type", "text/html")
		var buffer bytes.Buffer

		req := &pb.ServerHealth{}
		req.Status = pb.STATUS_STATUS_BUSY
		req.CurrentMillis = time.Now().UnixNano() / int64(time.Millisecond)
		if err := sendHealth(req, globalStream); err != nil {
			buffer.WriteString(fmt.Sprintln("Request to grpc, my health ", req.Status, ",time=", req.CurrentMillis))
		}
		_, _ = w.Write(buffer.Bytes())

	}))

	mux.Handle("/stopHealthStream", whataphttp.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "text/html")
		var buffer bytes.Buffer

		// close stream
		if globalStream != nil {
			if err := globalStream.CloseSend(); err == nil {
				buffer.WriteString("Closed Stream")
				globalStream = nil
			} else {
				buffer.WriteString("Error closing stream <hr/>")
				buffer.WriteString(err.Error())
			}
		}
		_, _ = w.Write(buffer.Bytes())

	}))

	mux.Handle("/ReqSub", whataphttp.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		log.Println("Request ", r.RequestURI)

		w.Header().Add("Content-Type", "text/html")
		var buffer bytes.Buffer
		buffer.WriteString("Request: <hr/>")
		buffer.WriteString(r.RequestURI)
		buffer.WriteString("<hr/>")

		req := &pb.Request{}
		req.Uri = r.RequestURI
		req.Params = parseParams(r)
		buffer.WriteString("<hr/>")
		for _, it := range req.Params {
			buffer.WriteString(it.Key + "=" + it.Value)
		}
		buffer.WriteString("<hr/>")

		if res, err := c.ReqSub(ctx, req); err == nil {
			buffer.WriteString("request sub grpc<hr/>")
			buffer.WriteString(fmt.Sprintf("status_code=%d", res.GetStatusCode()))
			buffer.WriteString("body<hr/>")
			buffer.WriteString(res.Body)
			buffer.WriteString("<hr/>")
			log.Println("Request to grpc, response ", res.GetStatusCode(), ",body=", res.Body)
		} else {
			buffer.WriteString("request grpc<hr/>")
			buffer.WriteString("body<hr/>")
			buffer.WriteString("Error " + err.Error())
			buffer.WriteString("<hr/>")
			log.Println("Error Reqeust to grpc ", err)
		}

		_, _ = w.Write(buffer.Bytes())

	}))

	mux.Handle("/streamHealth", whataphttp.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "text/html")
		var buffer bytes.Buffer

		health, err := c.Health(r.Context())
		if err != nil {
			log.Fatalf("Error stream : %v", err)
		} else {
			log.Println("streamHealth stream : ", health)
		}

		for i := 1; i <= 3; i++ {
			func() {
				defer func() {
					if r := recover(); r != nil {
						log.Println("Recover ", r)
					}
				}()
				req := &pb.ServerHealth{}
				req.Status = pb.STATUS_STATUS_BUSY
				req.CurrentMillis = time.Now().UnixNano() / int64(time.Millisecond)
				if err := sendHealth(req, health); err != nil {
					log.Println("Error streamHealth Send", err)
				} else {
					log.Println("streamHealth Send", req)
					buffer.WriteString(fmt.Sprintln("Request to grpc, my health ", req.Status, ",time=", req.CurrentMillis))
				}

				if in, err := recvHealth(health); err == nil {
					log.Println("streamHealth Recv", in)
					buffer.WriteString(fmt.Sprintln("Request to grpc, my health ", in.GetStatus(), ",time=", in.GetCurrentMillis()))
				} else {
					log.Println("Error streamHealth Recv ", err)
				}
			}()
			time.Sleep(1 * time.Second)
		}

		if err := health.CloseSend(); err != nil {
			log.Println("Error streamHealth CloseSend")
		} else {
			log.Println("streamHealth CloseSend")
		}

		_, _ = w.Write(buffer.Bytes())

	}))

	mux.Handle("/index", whataphttp.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		log.Println("Request ", r.RequestURI)

		w.Header().Add("Content-Type", "text/html")
		var buffer bytes.Buffer
		buffer.WriteString("Request: <hr/>")
		buffer.WriteString(r.RequestURI)
		buffer.WriteString("<hr/>")

		req := &pb.Request{}
		req.Uri = r.RequestURI
		req.Params = parseParams(r)
		buffer.WriteString("<hr/>")
		for _, it := range req.Params {
			buffer.WriteString(it.Key + "=" + it.Value)
		}
		buffer.WriteString("<hr/>")

		if res, err := c.Req(ctx, req); err == nil {
			buffer.WriteString("request grpc<hr/>")
			buffer.WriteString(fmt.Sprintf("status_code=%d", res.GetStatusCode()))
			buffer.WriteString("body<hr/>")
			buffer.WriteString(res.Body)
			buffer.WriteString("<hr/>")
			log.Println("Request to grpc, response ", res.GetStatusCode(), ",body=", res.Body)
		} else {
			buffer.WriteString("request grpc<hr/>")
			buffer.WriteString("body<hr/>")
			buffer.WriteString("Error " + err.Error())
			buffer.WriteString("<hr/>")
			log.Println("Error Reqeust to grpc ", err)
		}

		_, _ = w.Write(buffer.Bytes())

	}))

	mux.Handle("/", whataphttp.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tp, err := template.ParseFiles("templates/google.golang.org/grpc/client/index.html")
		if err != nil {
			fmt.Println("Template not loaded, ", err)
			return
		}
		data := &HTMLData{}
		data.Title = "grpc client"
		data.Content = r.RequestURI
		tp.Execute(w, data)
	}))

	log.Println("Start :", *port, ", Grpc port:", *grpcPort, ", Whatap Udp port:", *udpPort)

	_ = http.ListenAndServe(fmt.Sprintf(":%d", *port), mux)

}
