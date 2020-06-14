package main

import (
	"alitest2020/agentd/pb"
	"bufio"
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// 数据过滤代理服务agentd
type agentServer struct {
	opt *option //服务参数
	f   *filter //数据过滤
	pb.UnimplementedAgentServiceServer
}

func (s *agentServer) initServer(opt *option) {
	s.opt = opt
	s.f = newFilter(opt)
}

// 被processd调用，通知agentd需要上报的traceid
func (s *agentServer) NotifyTargetTraceid(ctx context.Context, in *pb.TraceidRequest) (*pb.Reply, error) {
	traceid := in.GetTraceid()
	info := s.f.addTraceidKey(traceid, time.Now().Unix())
	if !info.isTarget {
		if info.hasFile {
			info.loadFromFile(traceid)
			info.hasFile = false
		}
		info.isTarget = true
	}
	defer func() {
		if err := recover(); err != nil {
			log.Println("NotifyTargetTraceid panic:%v", err)
		}
	}()
	info.sendAllCurrentDataToBuf(s.f.cli)
	return &pb.Reply{Reply: []byte("ok")}, nil
}

func (s *agentServer) NotifyAllFilterOver(ctx context.Context, in *pb.Req) (*pb.Reply, error) {
	s.f.cli.endSend()
	return &pb.Reply{Reply: []byte("ok")}, nil
}

//下载数据并过滤
// TODO 考虑是否要分片下载
func (s *agentServer) getData(url string) {
	log.Printf("start get data")
	//	ctx, cancel := context.WithCancel(context.Background())
	//	defer cancel()
	go s.f.cli.runStreamSendToProcessd()
	startTime := time.Now().Unix()
	res, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	reader := bufio.NewReader(res.Body)
	for {
		span, err := reader.ReadBytes('\n')
		s.f.handleSpan(span)
		if err == io.EOF {
			break
		}
	}
	s.f.cli.notifyFilterOver()
	log.Printf("finish get data, time %d s", time.Now().Unix()-startTime)
}

func (s *agentServer) StartGrpcServer() {
	lis, err := net.Listen("tcp", ":"+s.opt.grpcPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	var kasp = keepalive.ServerParameters{
		Time:    5 * time.Second,  // Ping the client if it is idle for 5 seconds to ensure the connection is still active
		Timeout: 10 * time.Second, // Wait 1 second for the ping ack before assuming the connection is dead
	}
	srv := grpc.NewServer(grpc.KeepaliveParams(kasp))
	pb.RegisterAgentServiceServer(srv, s)
	if err := srv.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
func (s *agentServer) StartHttpServer() {
	http.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {})
	http.HandleFunc("/setParameter", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		port := r.Form.Get("port")
		if port != "" {
			url := "http://localhost:" + port + "/" + s.opt.dataFilename
			go s.getData(url)
		}
	})
	log.Fatal(http.ListenAndServe(":"+s.opt.port, nil))

}
func (s *agentServer) SignalHandle() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM,
		syscall.SIGINT, syscall.SIGSTOP, syscall.SIGUSR1, syscall.SIGPIPE)

	for {
		select {
		case s := <-c:
			switch s {
			case syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT:
				log.Printf("receive signal '%v' server quit", s)
				return
			default:
				log.Printf("receive signal '%v' but no processor", s)
			}
		}
	}
}
