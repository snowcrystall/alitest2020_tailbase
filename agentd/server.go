package main

import (
	"alitest2020/agentd/pb"
	"bufio"
	"context"
	"google.golang.org/grpc"
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
	info := s.f.addTraceidKey(traceid)
	info.isTarget = true
	//go info.runSendData(traceid, s.f.cli.sendchan, s.f.doneChan)
	return &pb.Reply{Reply: "ok"}, nil
}

//下载数据并过滤
// TODO 考虑是否要分片下载
func (s *agentServer) getData(url string) {
	log.Printf("start get data")
	startTime := time.Now().Unix()
	res, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	scanner := bufio.NewScanner(res.Body)
	go s.f.runFilterdata()
	for scanner.Scan() {
		select {
		case s.f.spanchan <- scanner.Text() + "\n":
		case <-time.After(time.Second * 1):
			log.Printf("wait 1 second , spanchan is blocked ,can't put data in it")
		}
	}
	s.f.spanchan <- "EOF"
	log.Printf("finish get data, time %d s", time.Now().Unix()-startTime)
}

func (s *agentServer) StartGrpcServer() {
	lis, err := net.Listen("tcp", ":"+s.opt.grpcPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	srv := grpc.NewServer()
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
