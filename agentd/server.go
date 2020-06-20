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
	"strings"
	"sync"
	"syscall"
	"time"
)

// agentd server struct , filte source data ,than send target data to processd
type agentServer struct {
	opt     *option        //option from cmd args
	wg      sync.WaitGroup // wait for download gorutines complete
	wgCheck sync.WaitGroup // wait for check data gorutines complete, which targeted datas
	wgSend  sync.WaitGroup // wait for send data gorutines complete, which send targeted datas to processd

	buf            []byte       // data buffer
	sendCur        int64        // the position where data are sending
	checkCur       int64        // the position where data are checking
	checkEndChan   chan int64   //use for tell send gorutines where data has been checked
	checkOffset    chan int64   //use for tell check gorutines where data can be start checking
	sendOffsetChan chan int64   //use for tell download gorutines which is waiting for sendCur to goon download
	traceMap       sync.Map     //target data map
	cli            *processdCli //processd cli
	sendNum        int64        //total send num

	pb.UnimplementedAgentServiceServer //pb grpc
}

const (
	BUFSIZE        = 2048 * 1024 * 1024 // data buf size
	EACH_DOWNLOAD  = 500 * 1024 * 1024  //each download data size
	CHECK_DISTANCE = 20 * 1024 * 1024   // every download CHECK_DISTANCE data ,it will be check
	SEND_DISTANCE  = 300 * 1024 * 1024  //every check SEND_DISTANCE data,it will be send
	/*BUFSIZE        = 150 * 1024 * 1024 // data buf size
	EACH_DOWNLOAD  = 50 * 1024 * 1024  //each download data size
	CHECK_DISTANCE = 5 * 1024 * 1024   // every download CHECK_DISTANCE data ,it will be check
	SEND_DISTANCE  = 30 * 1024 * 1024  //every check SEND_DISTANCE data,it will be send
	*/
)

func (s *agentServer) initServer(opt *option) {
	s.opt = opt
	s.buf = make([]byte, BUFSIZE)
	s.sendCur = 0
	s.checkCur = 0
	s.checkEndChan = make(chan int64, 50)
	s.checkOffset = make(chan int64, 50)
	s.sendOffsetChan = make(chan int64, 50)
	s.traceMap = sync.Map{}
	s.cli = NewProcessdCli(s.opt)
}

// 被processd调用，通知agentd需要上报的traceid
func (s *agentServer) NotifyTargetTraceid(ctx context.Context, in *pb.TraceidRequest) (*pb.Reply, error) {
	traceid := in.GetTraceid()
	s.traceMap.LoadOrStore(bytesToInt64(traceid), true)
	return &pb.Reply{Reply: []byte("ok")}, nil
}

//下载数据并过滤
func (s *agentServer) GetData(url string) {
	startTime := time.Now().UnixNano()
	s.wg = sync.WaitGroup{}
	s.wgCheck = sync.WaitGroup{}
	var start int64 = 0
	resChan := make(chan *http.Response, 10)
	for i := 0; i < s.opt.downn; i++ {
		s.wg.Add(1)
		go s.ReadRangeData(resChan, i)
		break
	}
	for i := 0; i < s.opt.sendn; i++ {
		s.wgSend.Add(1)
		go s.SendRangeData()
		break
	}
	for i := 0; i < s.opt.sendn*2; i++ {
		s.wgCheck.Add(1)
		go s.CheckTargetData()
		break
	}

	for {
		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			log.Fatalf("Invalid url for downloading, error: %v", err)
		}
		client := &http.Client{}
		req.Header.Set("Range", "bytes="+strconv.FormatInt(start, 10)+"-"+strconv.FormatInt(start+EACH_DOWNLOAD, 10))
		res, err := client.Do(req)
		if err != nil {
			log.Fatal(err)
		}
		if res.StatusCode != http.StatusPartialContent {
			close(resChan)
			break
		}
		start += res.ContentLength
		resChan <- res
	}
	s.wg.Wait()
	close(s.checkOffset)
	log.Printf("finish download data, time %d ms", (time.Now().UnixNano()-startTime)/1000000)
	s.wgCheck.Wait()
	close(s.checkEndChan)
	log.Printf("finish check data, time %d ms, %d", (time.Now().UnixNano()-startTime)/1000000, s.checkCur)
	s.wgSend.Wait()
	s.cli.NotifySendOver()
	s.cli.Close()
	log.Printf("finish send  data, time %d ms, num %d ", (time.Now().UnixNano()-startTime)/1000000, s.sendNum)
	s.PrintMemUsage()
}

func (s *agentServer) ReadRangeData(resChan chan *http.Response, who int) {
	defer s.wg.Done()
	for {
		res, ok := <-resChan
		if !ok {
			break
		}
		//reader := bufio.NewReader(res.Body)
		//reader := bufio.NewReaderSize(res.Body, 1024*1024)
		startistr := strings.TrimLeft(strings.Split(res.Header["Content-Range"][0], "-")[0], "bytes ")
		starti, _ := strconv.ParseInt(startistr, 10, 64)
		var sum, offset int64 = 0, 0
		var n, m int = 0, 0
		var err error
		reader := bufio.NewReaderSize(res.Body, 1048576)
		for {
			offset = starti + sum
			if offset >= BUFSIZE {
				offset = offset % BUFSIZE
			}
			n, err = reader.Read(s.buf[offset:])
			sum = sum + int64(n)
			m += n
			if m > CHECK_DISTANCE {
				ii := bytes.LastIndex(s.GetRangeBufFromAbsoluteOffset(starti, starti+sum), []byte{'\n'})
				s.checkOffset <- starti + int64(ii) + 1
				if starti+sum+CHECK_DISTANCE >= BUFSIZE && s.sendCur < BUFSIZE {
					for s.sendCur%BUFSIZE <= (starti+sum+CHECK_DISTANCE)%BUFSIZE {
						log.Printf("ReadRangeData waiting for sendCur")
						<-s.sendOffsetChan
					}
				}

			}
			//log.Printf("reader num ,%d", n)
			if err == io.EOF {
				break
			}
		}
		res.Body.Close()
		//log.Printf("read range data,from %d , to %d", starti, starti+sum)
	}
}

func (s *agentServer) GetRangeBufFromAbsoluteOffset(start, end int64) []byte {
	if start < 0 {
		start = 0
	}
	if start >= BUFSIZE {
		start = start % BUFSIZE
	}
	if end >= BUFSIZE {
		end = end % BUFSIZE
	}

	if start > end {
		return append(s.buf[start:BUFSIZE], s.buf[:end]...)
	} else {
		return s.buf[start:end]
	}
}

func (s *agentServer) CheckTargetData() {
	//offset, end maybe larger than BUFSIZE
	defer s.wgCheck.Done()
	var offset int64
	var ok bool
	var err error
	for {
		offset, ok = <-s.checkOffset
		if !ok {
			break
		}
		bs := s.GetRangeBufFromAbsoluteOffset(s.checkCur, offset)
		buffer := NewBuffer(bs)
		span := []byte{}
		var tagi, traceIdIndex int
		var traceId int64
		//log.Printf("%s", s.buf)
		for {
			span, err = buffer.ReadSlice([]byte{'\n'})
			if err == io.EOF {
				break
			}
			tagi = bytes.LastIndex(span, []byte{'|'})
			traceIdIndex = bytes.Index(span, []byte{'|'})
			traceId = bytesToInt64(span[:traceIdIndex])
			if checkIsTarget(span[tagi:]) {
				//log.Printf("%s  ,  %s  ,   %x ", span[tagi:], span[:traceIdIndex], traceId)
				s.traceMap.LoadOrStore(traceId, true)
				s.cli.SetTargetTraceidToProcessd(span[:traceIdIndex])
			}
		}
		//log.Printf("CheckTargetData from %d ,to %d ", lastOffset, offset)
		s.checkCur = offset
		if s.checkCur-s.sendCur > SEND_DISTANCE {
			s.checkEndChan <- offset
		}
	}
	s.checkEndChan <- s.checkCur
}

func (s *agentServer) SendRangeData() {
	defer s.wgSend.Done()
	cli := NewProcessdCli(s.opt)
	stream := cli.GetStream()
	var ok bool
	for {
		offset, chanOk := <-s.checkEndChan
		if !chanOk {
			log.Printf("get checkEndChan close, %d ", offset)
			break
		}
		//log.Printf("SendRangeData from %d, to %d  ", s.sendCur, offset)
		bs := s.GetRangeBufFromAbsoluteOffset(s.sendCur, offset)
		s.sendCur = offset
		//log.Printf("%s", s.buf)
		buffer := NewBuffer(bs)
		span := []byte{}
		var err error
		var traceIdIndex int
		var traceId int64
		for {
			span, err = buffer.ReadSlice([]byte{'\n'})
			if err == io.EOF {
				break
			}
			traceIdIndex = bytes.Index(span, []byte{'|'})
			traceId = bytesToInt64(span[:traceIdIndex])
			_, ok = s.traceMap.Load(traceId)
			if ok {
				//log.Printf("%s", span)
				stream.Send(&pb.TraceData{Tracedata: span})
				s.sendNum++
			}
		}
		s.sendOffsetChan <- s.sendCur
	}
	_, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("%v.CloseAndRecv() got error %v, want %v", stream, err, nil)
	}
	cli.Close()
}

func checkIsTarget(tag []byte) bool {
	//判断error 等于1的调用链路
	if bytes.Contains(tag, []byte("error=1")) {
		return true
	}
	// 找到所有tags中存在 http.status_code 不为 200
	if index := bytes.Index(tag, []byte("http.status_code=")); index != -1 {
		if !(tag[index+17] == 50 && tag[index+18] == 48 && tag[index+19] == 48) {
			return true
		}
	}
	return false
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
			go s.GetData(url)
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
