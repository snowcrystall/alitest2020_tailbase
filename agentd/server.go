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
	pb.UnimplementedAgentServiceServer
	wg     sync.WaitGroup
	wgSend sync.WaitGroup
	f      *filter
}

func (s *agentServer) initServer(opt *option) {
	s.opt = opt
	s.f = &filter{sync.Map{}, newProcessdCli(opt)}
}

// 被processd调用，通知agentd需要上报的traceid
func (s *agentServer) NotifyTargetTraceid(ctx context.Context, in *pb.TraceidRequest) (*pb.Reply, error) {
	traceid := in.GetTraceid()
	s.f.traceMap.LoadOrStore(bytesToInt64(traceid), true)
	return &pb.Reply{Reply: []byte("ok")}, nil
}

func (s *agentServer) NotifyAllFilterOver(ctx context.Context, in *pb.Req) (*pb.Reply, error) {
	log.Printf("NotifyAllFilterOver")
	return &pb.Reply{Reply: []byte("ok")}, nil
}

//下载数据并过滤
func (s *agentServer) getData(url string) {
	startTime := time.Now().Unix()
	log.Printf("start get data")
	req, err := http.NewRequest(http.MethodHead, url, nil)
	if err != nil {
		log.Fatalf("Invalid url for downloading, error: %v", err)
	}
	client := &http.Client{}
	res, err := client.Do(req)
	s.wg = sync.WaitGroup{}
	s.wgSend = sync.WaitGroup{}
	length, _ := strconv.ParseInt(res.Header["Content-Length"][0], 10, 64)
	each := length / int64(s.opt.fn)
	dup := 20000 * 256
	for i := 0; i < s.opt.fn; i++ {
		s.wg.Add(1)
		start := int64(i) * each
		end := start + each + int64(dup)
		go s.runRangeData(url, start, end)
		if end >= length {
			break
		}
	}
	s.wg.Wait()
	//s.f.cli.notifyFilterOver()
	log.Printf("finish get data, time %d s", time.Now().Unix()-startTime)
	s.wgSend.Wait()
	s.f.cli.notifySendOver()
	log.Printf("finish send  data, time %d s", time.Now().Unix()-startTime)
	s.PrintMemUsage()
}

func (s *agentServer) runRangeData(url string, start int64, end int64) bool {
	log.Printf("start get range data,from %d , to %d", start, end)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Fatalf("Invalid url for downloading, error: %v", err)
	}
	client := &http.Client{}
	req.Header.Set("Range", "bytes="+strconv.FormatInt(start, 10)+"-"+strconv.FormatInt(end, 10))
	res, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	if res.StatusCode != http.StatusPartialContent {
		return false
	}
	reader := bufio.NewReaderSize(res.Body, 1024*1024*10)

	var fields [][]byte
	var span []byte
	r := newRangeFilter(s.opt, s.f)

	s.wgSend.Add(1)
	go r.runSendData(&s.wgSend)
	for {
		span, err = reader.ReadBytes('\n')
		if err == io.EOF {
			break
		}
		fields = bytes.Split(span, []byte("|"))
		if len(fields) < 9 {
			continue
		}
		r.checkSpan(span, fields)
	}
	r.checkOver = true
	r.sendCurCond.L.Lock()
	r.sendCurCond.Signal()
	r.sendCurCond.L.Unlock()
	defer res.Body.Close()
	defer s.wg.Done()
	log.Printf("end get range data,from %d , to %d", start, end)
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

func bytesToInt64(a []byte) (u int64) {
	for _, c := range a {
		u *= int64(16)
		switch {
		case '0' <= c && c <= '9':
			u += int64(c - '0')
		case 'a' <= c && c <= 'z':
			u += int64(c - 'a' + 10)
		}
	}
	return u
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
