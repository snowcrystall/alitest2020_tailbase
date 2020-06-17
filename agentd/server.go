package main

import (
	"alitest2020/agentd/pb"
	"bufio"
	"bytes"
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"sync"
	"syscall"
	"time"
)

// 数据过滤代理服务agentd
type agentServer struct {
	opt *option //服务参数
	f   *filter //数据过滤
	pb.UnimplementedAgentServiceServer
	wg sync.WaitGroup
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
func (s *agentServer) getData(url string) {
	startTime := time.Now().Unix()
	log.Printf("start get data")
	go s.f.cli.runStreamSendToProcessd()
	req, err := http.NewRequest(http.MethodHead, url, nil)
	if err != nil {
		log.Fatalf("Invalid url for downloading, error: %v", err)
	}
	client := &http.Client{}
	res, err := client.Do(req)
	var i int64
	s.wg = sync.WaitGroup{}
	length, _ := strconv.ParseInt(res.Header["Content-Length"][0], 10, 64)
	each := length / int64(s.opt.fn)
	lruTraceidList := make([]*Queue, s.opt.fn+1)
	rangeNum := 0
	for i < length {
		lruTraceidList[rangeNum] = NewQueue()
		s.wg.Add(1)
		go s.rangeData(url, i, each, lruTraceidList, rangeNum)
		i = i + each - 512
		rangeNum++
	}
	s.wg.Wait()
	s.f.cli.notifyFilterOver()
	log.Printf("finish get data, time %d s", time.Now().Unix()-startTime)
	s.PrintMemUsage()
}

func (s *agentServer) rangeData(url string, start int64, each int64, lruTraceidList []*Queue, rangeNum int) bool {
	log.Printf("start get range data,from %d , each %d", start, each)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Fatalf("Invalid url for downloading, error: %v", err)
	}
	client := &http.Client{}
	var span []byte
	strStart := strconv.FormatInt(start, 10)
	strEnd := strconv.FormatInt(start+each, 10)
	req.Header.Set("Range", "bytes="+strStart+"-"+strEnd)
	res, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	if res.StatusCode != http.StatusPartialContent {
		return false
	}
	reader := bufio.NewReader(res.Body)
	var fields [][]byte
	for {
		span, err = reader.ReadBytes('\n')
		if err == io.EOF {
			break
		}
		fields = bytes.Split(span, []byte("|"))
		if len(fields) < 9 {
			continue
		}
		s.f.handleSpan(span, fields, lruTraceidList, rangeNum)
	}
	defer res.Body.Close()
	defer s.wg.Done()
	log.Printf("end get range data,from %d , each %d", start, each)
	return true
}

func (s *agentServer) PrintMemUsage() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	log.Printf("Alloc = %v MiB", bToMb(m.Alloc))
	log.Printf("\tTotalAlloc = %v MiB", bToMb(m.TotalAlloc))
	log.Printf("\tSys = %v MiB", bToMb(m.Sys))
	log.Printf("\tNumGC = %v\n", m.NumGC)
}
func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
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
