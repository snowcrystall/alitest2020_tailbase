package main

import (
	"alitest2020_tailbase/agentd/pb"
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
	"strconv"
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

	buf          []byte              // data buffer
	resChan      chan *http.Response // http.Response
	sendCur      int64               // the position where data are sending
	checkCur     int64               // the position where data are checking
	checkEndChan chan [2]int64       //use for tell send gorutines where data has been checked
	checkEnd     bool                // if check gotutines finished
	checkOffset  chan [2]int64       //use for tell check gorutines where data can be start checking
	traceMap     sync.Map            //target data map
	traceBitmap  [16]uint16          //target data bitmap
	cli          *processdCli        //processd cli
	sendNum      int64               //total send num
	c            *sync.Cond          //used for check gorutine to signal send gorutine

	pb.UnimplementedAgentServiceServer //pb grpc
}

const (
	BUFSIZE        = 2048 * 1024 * 1024 // data buf size
	EACH_DOWNLOAD  = 128 * 1024 * 1024  //each download data size
	CHECK_DISTANCE = 16 * 1024 * 1024   // every download CHECK_DISTANCE data ,it will be check
	SEND_DISTANCE  = 512 * 1024 * 1024  //every check SEND_DISTANCE data,it will be send

	/*BUFSIZE        = 50 * 1024 * 1024 // data buf size
	EACH_DOWNLOAD  = 5 * 1024 * 1024  //each download data size
	CHECK_DISTANCE = 1 * 1024 * 1024  // every download CHECK_DISTANCE data ,it will be check
	SEND_DISTANCE  = 25 * 1024 * 1024 //every check SEND_DISTANCE data,it will be send
	*/
)

func (s *agentServer) initServer(opt *option) {
	s.opt = opt
	s.buf = make([]byte, BUFSIZE, BUFSIZE)
	s.sendCur = 0
	s.checkCur = 0
	// checkEndChan must bigger than SEND_DISTANCE/CHECK_DISTANCE
	s.checkEndChan = make(chan [2]int64, 500)
	s.resChan = make(chan *http.Response, 2)
	s.checkOffset = make(chan [2]int64, 500)
	s.traceMap = sync.Map{}
	s.cli = NewProcessdCli(s.opt)
	s.wg = sync.WaitGroup{}
	s.wgCheck = sync.WaitGroup{}
	s.wgSend = sync.WaitGroup{}
	s.c = sync.NewCond(&sync.Mutex{})
	s.checkEnd = false
}

// 被processd调用，通知agentd需要上报的traceid
func (s *agentServer) NotifyTargetTraceid(ctx context.Context, in *pb.TraceidRequest) (*pb.Reply, error) {
	traceid := in.GetTraceid()
	//s.SetTraceIdTarget(bytesToInt64(traceid))
	s.traceMap.LoadOrStore(bytesToInt64(traceid), true)
	return &pb.Reply{Reply: []byte("ok")}, nil
}

//下载数据并过滤
func (s *agentServer) GetData(url string) {
	startTime := time.Now().UnixNano()
	var start int64 = 0

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
		start += int64(res.ContentLength - 300)
		s.resChan <- res
		if res.ContentLength < EACH_DOWNLOAD {
			close(s.resChan)
			break
		}
	}
	s.wg.Wait()
	close(s.checkOffset)
	log.Printf("finish download data, time %d ms", (time.Now().UnixNano()-startTime)/1000000)
	s.wgCheck.Wait()
	close(s.checkEndChan)
	s.c.L.Lock()
	log.Printf("finish check data, time %d ms, %d", (time.Now().UnixNano()-startTime)/1000000, s.checkCur)
	s.checkEnd = true
	s.c.Signal()
	s.c.L.Unlock()
	s.wgSend.Wait()
	s.cli.NotifySendOver()
	s.cli.Close()
	log.Printf("finish send  data, time %d ms, num %d ", (time.Now().UnixNano()-startTime)/1000000, s.sendNum)
	if s.opt.debug == 1 {
		pprofMemory()
	}
	//runtime.GC() // get up-to-date statistics
	//debug.FreeOSMemory()
}

func (s *agentServer) ReadRangeData() {
	defer s.wg.Done()
	for {
		res, ok := <-s.resChan
		if !ok {
			break
		}
		s.dealres(res)
	}
}

func (s *agentServer) dealres(res *http.Response) {
	//defer fasthttp.ReleaseResponse(res)
	//log.Printf("%s", res.Header.Get("Content-Range"))
	startistr := bytes.TrimLeft(bytes.Split([]byte(res.Header.Get("Content-Range")), []byte{'-'})[0], "bytes ")
	starti, _ := strconv.ParseInt(string(startistr), 10, 64)
	var relatOffset, absolOffset, offsetmLast, readend int64
	var n int = 0
	var err error
	absolOffset = starti
	//reader := bytes.NewBuffer(res.Body())
	for {
		relatOffset = absolOffset
		if relatOffset >= BUFSIZE {
			relatOffset = relatOffset % BUFSIZE
		}
		readend = relatOffset + CHECK_DISTANCE
		if readend > BUFSIZE {
			readend = BUFSIZE
		}
		n, err = res.Body.Read(s.buf[relatOffset:readend])
		absolOffset += int64(n)
		if absolOffset-offsetmLast >= CHECK_DISTANCE || err == io.EOF {
			ii := absolOffset - 400 + int64(bytes.LastIndex(s.GetRangeBufFromAbsoluteOffset(absolOffset-400, absolOffset), []byte{'\n'}))
			if starti == 0 && offsetmLast == 0 {
				offsetmLast = -1
			} else if offsetmLast == 0 {
				offsetmLast = starti + int64(bytes.Index(s.GetRangeBufFromAbsoluteOffset(starti, starti+400), []byte{'\n'}))
			}
			s.checkOffset <- [2]int64{offsetmLast + 1, ii + 1}
			//log.Printf("put int checkOffset %d,%d", offsetmLast+1, ii+1)
			if ii-offsetmLast < 1000 {
				log.Printf("put int checkOffset %d,%d", offsetmLast+1, ii+1)
			}
			offsetmLast = ii + 1
			for absolOffset+CHECK_DISTANCE-s.sendCur > BUFSIZE {
				log.Printf("waiting for sendCur,will be write: %d, sendCur: %d checkCur: %d", absolOffset, s.sendCur, s.checkCur)
				time.Sleep(500 * time.Millisecond)
			}

		}
		if err == io.EOF {
			res.Body.Close()
			break
			//log.Printf("read range data,from %d , to %d", starti, starti+sum)
		}

	}
}
func (s *agentServer) GetRangeBufFromAbsoluteOffset(start, end int64) []byte {
	if start < 0 {
		start = 0
	}
	if start >= BUFSIZE {
		start = start % BUFSIZE
	}
	if end > BUFSIZE {
		end = end % BUFSIZE
	}
	if end == 0 {
		end = BUFSIZE
	}

	if start > end {
		log.Printf("GetRangeBufFromAbsoluteOffset append %d,%d", start, end)
		return append(s.buf[start:BUFSIZE], s.buf[:end]...)
	} else {
		return s.buf[start:end]
	}
}

func (s *agentServer) CheckTargetData() {
	//offset, end maybe larger than BUFSIZE
	defer func() {
		if r := recover(); r != nil {
			log.Println("Recovered panic", r)
			s.ShutDown()
		}
		s.wgCheck.Done()
	}()

	for {
		offset, ok := <-s.checkOffset
		if !ok {
			break
		}
		bs := s.GetRangeBufFromAbsoluteOffset(offset[0], offset[1])
		buffer := NewBuffer(bs)
		for {
			span, err := buffer.ReadSlice([]byte{'\n'})
			if err == io.EOF {
				break
			}
			tagi := bytes.LastIndex(span, []byte{'|'})
			traceIdIndex := bytes.Index(span, []byte{'|'})
			if traceIdIndex == -1 || tagi == -1 {
				log.Printf("%s,%d,%d", span, offset[0], offset[1])
				continue
			}
			traceId := bytesToInt64(span[:traceIdIndex])
			if checkIsTarget(span[tagi:]) {
				s.traceMap.LoadOrStore(traceId, true)
				//s.SetTraceIdTarget(traceId)
				s.cli.SetTargetTraceidToProcessd(span[:traceIdIndex])
			}
		}
		if s.checkCur < offset[1] {
			// gorutine unsafe, but noissue here
			s.checkCur = offset[1]
		}
		s.checkEndChan <- [2]int64{offset[0], offset[1]}
		//log.Printf("CheckTargetData from %d ,to %d ,checkCur %d", offset[0], offset[1], s.checkCur)
	}
}

func (s *agentServer) SendRangeData() {
	defer func() {
		if r := recover(); r != nil {
			log.Println("Recovered panic", r)
			s.ShutDown()
		}
		s.wgSend.Done()
	}()

	cli := NewProcessdCli(s.opt)
	stream := cli.GetStream()
	for {
		offset, chanOk := <-s.checkEndChan
		if !chanOk {
			break
		}
		for !s.checkEnd && s.checkCur-offset[1] < SEND_DISTANCE {
			log.Printf("sendCur %d wait for checkCur %d ,offset[1] %d", s.sendCur, s.checkCur, offset[1])
			time.Sleep(500 * time.Millisecond)
		}
		if s.sendCur < offset[1] {
			//log.Printf("SendRangeData from %d, to %d ,num: %d ", s.sendCur, offset[1], offset[1]-s.sendCur)
			s.sendRangeData(&stream, offset[1])
			s.sendCur = offset[1]
		}
	}
	_, err := stream.CloseAndRecv()
	if err != nil {
		log.Panicf("%v.CloseAndRecv() got error %v, want %v", stream, err, nil)
	}
	cli.Close()
}

func (s *agentServer) sendRangeData(stream *pb.ProcessService_SendTraceDataClient, offset int64) {
	bs := [2][]byte{}
	if s.sendCur/BUFSIZE < offset/BUFSIZE {
		bs[0] = s.GetRangeBufFromAbsoluteOffset(s.sendCur, BUFSIZE)
		bs[1] = s.GetRangeBufFromAbsoluteOffset(0, offset)
	} else {
		bs[0] = s.GetRangeBufFromAbsoluteOffset(s.sendCur, offset)
		bs[1] = []byte{}
	}
	//log.Printf("%s", s.buf)
	for _, v := range bs {
		buffer := NewBuffer(v)
		span := []byte{}
		lastspan := []byte{}
		tosend := &pb.TraceData{Tracedata: nil}
		var err error
		var traceIdIndex int
		var traceId int64
		for {
			span, err = buffer.ReadSlice([]byte{'\n'})
			if err == io.EOF {
				lastspan = span
				break
			}
			if len(lastspan) > 0 {
				span = append(lastspan, span...)
				lastspan = lastspan[:0]
			}
			traceIdIndex = bytes.Index(span, []byte{'|'})
			if traceIdIndex == -1 {
				log.Printf("%s, sendCur: %d, offset :%d ", span, s.sendCur, offset)
				continue
			}
			traceId = bytesToInt64(span[:traceIdIndex])
			_, ok := s.traceMap.Load(traceId)
			if ok {
				tosend.Tracedata = span
				(*stream).Send(tosend)
				s.sendNum++
			}
		}
	}
}

func checkIsTarget(tag []byte) bool {
	//判断error 等于1的调用链路
	if index := bytes.Index(tag, []byte("error=1")); index != -1 {
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
	http.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		for i := 0; i < s.opt.downn; i++ {
			s.wg.Add(1)
			go s.ReadRangeData()
		}
		for i := 0; i < s.opt.sendn; i++ {
			s.wgSend.Add(1)
			go s.SendRangeData()
		}
		for i := 0; i < s.opt.downn; i++ {
			s.wgCheck.Add(1)
			go s.CheckTargetData()
		}

	})
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

func (s *agentServer) ShutDown() {
	syscall.Kill(syscall.Getpid(), syscall.SIGINT)
}
func (s *agentServer) SignalHandle() {
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
