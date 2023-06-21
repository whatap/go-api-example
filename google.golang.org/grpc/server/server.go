package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"time"

	pb "github.com/whatap/go-api-example/google.golang.org/grpc/proto"
	// "google.golang.org/genproto"
	"github.com/whatap/go-api/instrumentation/google.golang.org/grpc/whatapgrpc"
	"github.com/whatap/go-api/method"
	"github.com/whatap/go-api/trace"
	"google.golang.org/grpc"
)

type NoticeServer struct {
	c pb.ServiceNoticeClient
}

func NewNoticeServer() *NoticeServer {
	return new(NoticeServer)
}

// GetFeature returns the feature at the given point.
func (s *NoticeServer) Req(ctx context.Context, req *pb.Request) (*pb.Response, error) {
	log.Println("Start Req")
	return &pb.Response{Uri: req.Uri, Params: req.Params, Body: "Req to Response"}, nil
}

// GetFeature returns the feature at the given point.
func (s *NoticeServer) ReqSub(ctx context.Context, req *pb.Request) (resp *pb.Response, err error) {
	log.Println("Start ReqSub")
	if s.c != nil {
		return s.c.ReqSub(ctx, req)
	} else {
		mCtx, _ := method.Start(ctx, "Req")
		resp, err = s.Req(ctx, req)
		method.End(mCtx, err)
		return resp, err
	}
}

func (s *NoticeServer) Health(stream pb.ServiceNotice_HealthServer) error {
	log.Println("Start Health")
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
		log.Println("Client Heath ", in.Status)

		res := &pb.ServerHealth{}
		if in.Status == pb.STATUS_STATUS_BUSY {
			res.Status = pb.STATUS_STATUS_BUSY
		} else {
			res.Status = pb.STATUS_STATUS_READY
		}
		res.CurrentMillis = time.Now().UnixNano() / int64(time.Millisecond)
		if err := stream.Send(res); err != nil {
			log.Println("Error send response ", err)
			return err
		}
	}
}

func main() {
	port := flag.Int("p", 8082, "grpc port. default 8082  ")
	udpPort := flag.Int("up", 6600, "whatap agent udp port")
	grpcEnable := flag.Bool("use_client", false, "use connect to other grpc server")
	grpcHost := flag.String("gh", "localhost", "grpc host. default localhost")
	grpcPort := flag.Int("gp", 8084, "grpc port. defalt 8082 ")
	flag.Parse()

	// Init WhaTap Trace
	config := make(map[string]string)
	config["net_udp_port"] = fmt.Sprintf("%d", *udpPort)
	trace.Init(config)
	defer trace.Shutdown()

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen %v", err)
	}
	// Set the whatap interceptor to grpc
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(whatapgrpc.UnaryServerInterceptor()),
		grpc.StreamInterceptor(whatapgrpc.StreamServerInterceptor()))

	ns := NewNoticeServer()
	//connect another server
	if *grpcEnable {
		conn, err := grpc.Dial(fmt.Sprintf("%s:%d", *grpcHost, *grpcPort), grpc.WithInsecure(),
			grpc.WithBlock(),
			grpc.WithUnaryInterceptor(whatapgrpc.UnaryClientInterceptor()),
			grpc.WithStreamInterceptor(whatapgrpc.StreamClientInterceptor()),
			grpc.WithTimeout(10*time.Second))

		if err != nil {
			log.Fatalf("did not connect: %v", err)
		}
		log.Println("Open grpc client to ", *grpcHost, ":", *grpcPort)
		ns.c = pb.NewServiceNoticeClient(conn)
	} else {
		ns.c = nil
	}
	pb.RegisterServiceNoticeServer(grpcServer, ns)

	log.Println("Open grpc server :", *port, ", Whatap udp port:", *udpPort)

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
