package main

import (
	"alitest2020_tailbase/pb"
	"bufio"
	"bytes"
	"github.com/cornelk/hashmap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// agentd server struct , filte source data ,than send target data to processd
type agentServerV2 struct {
	opt         *option        //option from cmd args
	wgCheck     sync.WaitGroup // wait for check data gorutines complete, which targeted datas
	wgSend      sync.WaitGroup // wait for send data gorutines complete, which send targeted datas to processd
	lock        *sync.Mutex
	isRunning   bool
	isReady     bool
	readWaiting int
	sendWaiting int

	pb.UnimplementedAgentServiceServer //pb grpc

	b        *LinesBuffer
	checkEnd bool // if check gotutines finished
	//traceMap   sync.Map     //target data map
	traceMap   *hashmap.HashMap
	cli        *processdCli //processd cli
	sendNum    int64        //total send num
	readSignal *sync.Cond   //used for send gorutine to signal read gorutine
	sendSignal *sync.Cond   //used for check gorutine to signal send gorutine

}

type LinesBuffer struct {
	Lines     []*LInfo
	SendLineN int
	PeerLineN int
}

type LInfo struct {
	TraceidIndex int
	LineN        int
	Line         []byte
}

const (
	//LINEBUFSIZE       = 5000000
	//SENDLINE_DISTANCE = 100000

	LINEBUFSIZE       = 1000000 // data buf size, 400m ~ 1500000
	SENDLINE_DISTANCE = 100000  //every check SEND_DISTANCE data,it will be send

)

func NewAgentServerV2(opt *option) (s *agentServerV2) {
	s = &agentServerV2{}
	s.opt = opt
	s.b = &LinesBuffer{}
	s.b.Lines = make([]*LInfo, LINEBUFSIZE, LINEBUFSIZE)
	s.traceMap = &hashmap.HashMap{}
	s.wgSend = sync.WaitGroup{}
	s.readSignal = sync.NewCond(&sync.Mutex{})
	s.sendSignal = sync.NewCond(&sync.Mutex{})
	s.checkEnd = false
	s.lock = &sync.Mutex{}
	s.cli = NewProcessdCli(s.opt)
	return s
}

// 被processd调用，通知agentd需要上报的traceid
func (s *agentServerV2) NotifyTargetTraceids(gs pb.AgentService_NotifyTargetTraceidsServer) error {
	for {
		in, err := gs.Recv()
		if err == io.EOF {
			gs.SendAndClose(&pb.Reply{Reply: []byte("ok")})
			break
		}
		if err != nil {
			log.Printf("failed to recv: %v", err)
			break
		}
		if in.Traceid != 0 {
			s.traceMap.Set(in.Traceid, true)
		}
		s.b.PeerLineN = int(in.Checkcur)
		if s.sendWaiting > 0 || s.b.PeerLineN == -1 {
			s.sendSignal.L.Lock()
			s.sendSignal.Signal()
			s.sendSignal.L.Unlock()
		}

	}

	return nil
}

//下载数据并过滤
func (s *agentServerV2) GetData(url string) {
	log.Printf("start GetData")
	startTime := time.Now().UnixNano()
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Fatalf("Invalid url for downloading, error: %v", err)
	}
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		log.Fatalf(err.Error())
	}
	go s.SendData()
	s.DealRes(res)
	log.Printf("finish download and check data, time %d ms", (time.Now().UnixNano()-startTime)/1000000)

	s.checkEnd = true
	s.sendSignal.L.Lock()
	s.sendSignal.Broadcast()
	s.sendSignal.L.Unlock()

	s.wgSend.Wait()
	s.cli.NotifySendOver()
	s.cli.Close()
	log.Printf("finish send  data, time %d ms, num %d ", (time.Now().UnixNano()-startTime)/1000000, s.sendNum)
	if s.opt.debug == 1 {
		pprofMemory()
		PrintMemUsage()
	}
	//runtime.GC() // get up-to-date statistics
	//debug.FreeOSMemory()
}

func (s *agentServerV2) DealRes(res *http.Response) {
	defer func() {
		res.Body.Close()
		if r := recover(); r != nil {
			log.Println("Recovered panic", r)
			s.ShutDown()
		}
	}()

	reader := bufio.NewReaderSize(res.Body, 64*1024*1024)
	cli := NewProcessdCli(s.opt)
	defer cli.Close()
	stream := cli.GetSendTargetIdStream()
	i := 0
	var line []byte
	var err error
	for {
		line, err = reader.ReadSlice('\n')
		if err == io.EOF {
			stream.Send(&pb.TargetInfo{Traceid: 0, Checkcur: -1})
			s.b.Lines[i%LINEBUFSIZE] = &LInfo{-1, -1, nil}
			break
		}
		traceIdIndex := bytes.IndexByte(line, '|')
		linfo := &LInfo{traceIdIndex, i, make([]byte, 0, 400)}
		linfo.Line = append(linfo.Line, line...)
		if s.b.Lines[i%LINEBUFSIZE] != nil && len(s.b.Lines[i%LINEBUFSIZE].Line) == 0 {
			log.Printf("read %d wait for SendLineN %d ,", i, s.b.SendLineN)
			s.readSignal.L.Lock()
			s.readWaiting++
			s.readSignal.Wait()
			s.readWaiting--
			s.readSignal.L.Unlock()
		}
		s.b.Lines[i%LINEBUFSIZE] = linfo
		i++
		traceId := bytesToInt64(line[:traceIdIndex])
		tagi := bytes.LastIndexByte(line, '|')
		_, ok := s.traceMap.Get(traceId)
		if !ok && checkIsTarget(line[tagi:]) {
			s.traceMap.Set(traceId, true)
			stream.Send(&pb.TargetInfo{Checkcur: int64(i), Traceid: traceId})
			//log.Printf("%s", traceId)
		}
		if i%SEND_DISTANCE == 0 {
			if s.sendWaiting > 0 {
				s.sendSignal.L.Lock()
				s.sendSignal.Signal()
				s.sendSignal.L.Unlock()
			}
			stream.Send(&pb.TargetInfo{Checkcur: int64(i), Traceid: 0})
		}
	}
	_, err = stream.CloseAndRecv()
	if err != nil {
		log.Panicf("%v.CloseAndRecv() got error %v, want %v", stream, err, nil)
	}
}

func (s *agentServerV2) SendData() {
	s.wgSend.Add(1)
	defer func() {
		/*if r := recover(); r != nil {
			log.Println("Recovered panic", r)
			s.ShutDown()
		}*/
		s.wgSend.Done()
	}()

	cli := NewProcessdCli(s.opt)
	defer cli.Close()
	stream := cli.GetSendDataStream()
	i := 0
	for {
		//log.Printf("SendRangeData from %d, to %d ,num: %d ", s.sendCur, offset, offset-s.sendCur)
		//for s.b.Lines[i%LINEBUFSIZE] == nil || (s.b.PeerLineN != -1 && s.b.PeerLineN-s.b.SendLineN < SENDLINE_DISTANCE) {
		for s.b.Lines[i%LINEBUFSIZE] == nil || len(s.b.Lines[i%LINEBUFSIZE].Line) == 0 {
			//log.Printf("sendLineN %d wait for PeerLineN %d ,nil:%t", s.b.SendLineN, s.b.PeerLineN, s.b.Lines[i%LINEBUFSIZE] == nil)
			s.sendSignal.L.Lock()
			s.sendWaiting++
			s.sendSignal.Wait()
			s.sendWaiting--
			s.sendSignal.L.Unlock()
		}
		linfo := s.b.Lines[i%LINEBUFSIZE]
		i++
		if linfo.LineN == -1 {
			break
		}
		_, ok := s.traceMap.Get(linfo.Line[:linfo.TraceidIndex])
		if ok {
			stream.Send(&pb.TraceData{Tracedata: linfo.Line})
			s.sendNum++
		}
		s.b.SendLineN = linfo.LineN
		linfo.Line = linfo.Line[:0]
		if s.readWaiting > 0 {
			s.readSignal.L.Lock()
			s.readSignal.Signal()
			s.readSignal.L.Unlock()
		}
	}
	_, err := stream.CloseAndRecv()
	if err != nil {
		log.Panicf("%v.CloseAndRecv() got error %v, want %v", stream, err, nil)
	}
}

func (s *agentServerV2) StartGrpcServer() {
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

func (s *agentServerV2) StartHttpServer() {
	http.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		s.lock.Lock()
		if s.isReady == true {
			s.lock.Unlock()
			return
		}
		s.isReady = true
		s.lock.Unlock()
	})
	http.HandleFunc("/setParameter", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		port := r.Form.Get("port")
		if port != "" {
			s.lock.Lock()
			if s.isRunning == true {
				s.lock.Unlock()
				return
			}
			s.isRunning = true
			s.lock.Unlock()
			url := "http://localhost:" + port + "/" + s.opt.dataFilename
			go s.GetData(url)
		}
	})
	log.Fatal(http.ListenAndServe(":"+s.opt.port, nil))

}

func (s *agentServerV2) ShutDown() {
	syscall.Kill(syscall.Getpid(), syscall.SIGINT)
}
func (s *agentServerV2) SignalHandle() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM,
		syscall.SIGINT, syscall.SIGSTOP, syscall.SIGUSR1, syscall.SIGPIPE)

	for {
		select {
		case v := <-c:
			switch v {
			case syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT:
				log.Printf("receive signal '%v' server quit", v)
				return
			default:
				log.Printf("receive signal '%v' but no processor", v)
			}
		}
	}
}
