package main

import (
	"alitest2020_tailbase/pb"
	"bytes"
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
	opt         *option        //option from cmd args
	wg          sync.WaitGroup // wait for download gorutines complete
	wgCheck     sync.WaitGroup // wait for check data gorutines complete, which targeted datas
	wgSend      sync.WaitGroup // wait for send data gorutines complete, which send targeted datas to processd
	lock        *sync.Mutex
	isRunning   bool
	isReady     bool
	readWaiting int

	pb.UnimplementedAgentServiceServer //pb grpc

	buf          []byte              // data buffer
	resChan      chan *http.Response // http.Response
	sendOffset   *Offset             // the position where data are sending
	checkOffset  *Offset             // the position where data are checking
	checkEndChan chan [2]int64       //use for tell send gorutines where data has been checked
	checkEnd     bool                // if check gotutines finished
	checkChan    chan [2]int64       //use for tell check gorutines where data can be start checking
	traceMap     sync.Map            //target data map
	cli          *processdCli        //processd cli
	sendNum      int64               //total send num
	readSignal   *sync.Cond          //used for send gorutine to signal read gorutine
	sendSignal   *sync.Cond          //used for check gorutine to signal send gorutine

}

const (
	/*BUFSIZE        = 3072 * 1024 * 1024 // data buf size
	  EACH_DOWNLOAD  = 512 * 1024 * 1024  //each download data size
	  CHECK_DISTANCE = 64 * 1024 * 1024   // every download CHECK_DISTANCE data ,it will be check
	  SEND_DISTANCE  = 1280 * 1024 * 1024 //every check SEND_DISTANCE data,it will be send
	*/
	BUFSIZE        = 500 * 1024 * 1024 // data buf size
	EACH_DOWNLOAD  = 100 * 1024 * 1024 //each download data size
	CHECK_DISTANCE = 20 * 1024 * 1024  // every download CHECK_DISTANCE data ,it will be check
	SEND_DISTANCE  = 300 * 1024 * 1024 //every check SEND_DISTANCE data,it will be send

)

func (s *agentServer) initServer(opt *option) {
	s.opt = opt
	s.buf = make([]byte, BUFSIZE, BUFSIZE)
	s.sendOffset = NewOffset()
	s.checkOffset = NewOffset()
	// checkEndChan must bigger than SEND_DISTANCE/CHECK_DISTANCE
	s.checkEndChan = make(chan [2]int64, 500)
	s.resChan = make(chan *http.Response, 5) // opt.downn
	s.checkChan = make(chan [2]int64, 500)   // opt.downn*EACH_DOWNLOAD/CHECK_DISTANCE
	s.traceMap = sync.Map{}
	s.cli = NewProcessdCli(s.opt)
	s.wg = sync.WaitGroup{}
	s.wgCheck = sync.WaitGroup{}
	s.wgSend = sync.WaitGroup{}
	s.readSignal = sync.NewCond(&sync.Mutex{})
	s.sendSignal = sync.NewCond(&sync.Mutex{})
	s.checkEnd = false
	s.lock = &sync.Mutex{}
}

// 被processd调用，通知agentd需要上报的traceid
func (s *agentServer) NotifyTargetTraceids(gs pb.AgentService_NotifyTargetTraceidsServer) error {
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
		s.traceMap.LoadOrStore(in.Traceid, true)
	}

	return nil
}

//下载数据并过滤
func (s *agentServer) GetData(url string) {
	log.Printf("start GetData")
	startTime := time.Now().UnixNano()
	var start int64 = 0
	retry := 3
	for {
		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			log.Fatalf("Invalid url for downloading, error: %v", err)
		}
		client := &http.Client{}
		req.Header.Set("Range", "bytes="+strconv.FormatInt(start, 10)+"-"+strconv.FormatInt(start+EACH_DOWNLOAD, 10))
		res, err := client.Do(req)
		if err != nil {
			log.Printf(err.Error())
			if retry < 0 {
				break
			}
			retry--
			continue
		}
		start += int64(res.ContentLength - 400)
		s.resChan <- res
		if res.ContentLength < EACH_DOWNLOAD {
			break
		}
	}
	close(s.resChan)
	s.wg.Wait()
	close(s.checkChan)
	log.Printf("finish download data, time %d ms", (time.Now().UnixNano()-startTime)/1000000)

	s.wgCheck.Wait()
	close(s.checkEndChan)
	s.checkEnd = true
	log.Printf("finish check data, time %d ms, %d", (time.Now().UnixNano()-startTime)/1000000, s.checkOffset.Cur)
	s.sendSignal.L.Lock()
	s.sendSignal.Broadcast()
	s.sendSignal.L.Unlock()
	/*close(s.cli.SendTargetTraceIdChan)
	s.cli.wg.Wait()
	log.Printf("finish send target id,time %d ms", (time.Now().UnixNano()-startTime)/1000000)
	*/
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
		for absolOffset+CHECK_DISTANCE*3-s.sendOffset.Cur > BUFSIZE {
			log.Printf("Wait sendCur,will be write: %d, sendCur: %d checkCur: %d", absolOffset, s.sendOffset.Cur, s.checkOffset.Cur)
			s.readSignal.L.Lock()
			s.readWaiting++
			s.readSignal.Wait()
			s.readWaiting--
			s.readSignal.L.Unlock()
		}

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
			ii := absolOffset - 400 + int64(bytes.LastIndexByte(s.GetRangeBufFromAbsoluteOffset(absolOffset-400, absolOffset), '\n'))
			if starti == 0 && offsetmLast == 0 {
				offsetmLast = -1
			} else if offsetmLast == 0 {
				offsetmLast = starti + int64(bytes.IndexByte(s.GetRangeBufFromAbsoluteOffset(starti, starti+400), '\n'))
			}
			s.checkChan <- [2]int64{offsetmLast + 1, ii + 1}
			offsetmLast = ii

		}
		if err == io.EOF {
			res.Body.Close()
			//log.Printf("read range data,from %d , to %d", starti, absolOffset)
			break
		}

	}
}
func (s *agentServer) GetRangeBufFromAbsoluteOffset(start, end int64) []byte {
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
	defer func() {
		if r := recover(); r != nil {
			log.Println("Recovered panic", r)
			s.ShutDown()
		}
		s.wgCheck.Done()
	}()
	cli := NewProcessdCli(s.opt)
	stream := cli.GetSendTargetIdStream()
	for {
		offset, ok := <-s.checkChan
		if !ok {
			break
		}
		bs := s.GetRangeBufFromAbsoluteOffset(offset[0], offset[1])
		buffer := NewBuffer(bs)
		for {
			span, traceIdBytes, err := buffer.ReadLineWithTraceId()
			//_, _, err := buffer.ReadLineWithTraceId()
			if err == io.EOF {
				break
			}
			traceId := bytesToInt64(traceIdBytes)
			tagi := bytes.LastIndexByte(span, '|')
			//s.traceMap.Load(traceId)
			if checkIsTarget(span[tagi:]) {
				s.traceMap.LoadOrStore(traceId, true)
				stream.Send(&pb.TargetInfo{Traceid: traceId, Checkcur: s.checkOffset.Cur})
			}
		}
		for _, o := range s.checkOffset.SlideCur(offset, true) {
			s.checkEndChan <- o
			if s.sendOffset.Waiting > 0 {
				s.sendSignal.L.Lock()
				s.sendSignal.Broadcast()
				s.sendSignal.L.Unlock()
			}
		}
		//s.checkEndChan <- offset
	}
	_, err := stream.CloseAndRecv()
	if err != nil {
		log.Panicf("%v.CloseAndRecv() got error %v, want %v", stream, err, nil)
	}
	cli.Close()

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
	stream := cli.GetSendDataStream()
	for {
		offset, chanOk := <-s.checkEndChan
		if !chanOk {
			break
		}
		//log.Printf("SendRangeData from %d, to %d ,num: %d ", s.sendCur, offset, offset-s.sendCur)
		for !s.checkEnd && s.checkOffset.Cur-offset[1] < SEND_DISTANCE {
			//		log.Printf("sendCur %d wait for checkCur %d ,offset[1] %d", s.sendOffset.Cur, s.checkOffset.Cur, offset)
			s.sendSignal.L.Lock()
			s.sendOffset.Waiting++
			s.sendSignal.Wait()
			s.sendOffset.Waiting--
			s.sendSignal.L.Unlock()
		}
		s.sendRangeData(&stream, offset[0], offset[1])
		s.sendOffset.SlideCur(offset, false)
		if s.readWaiting > 0 {
			s.readSignal.L.Lock()
			s.readSignal.Broadcast()
			s.readSignal.L.Unlock()
		}
		//s.sendOffset.Cur = offset[1]
	}
	_, err := stream.CloseAndRecv()
	if err != nil {
		log.Panicf("%v.CloseAndRecv() got error %v, want %v", stream, err, nil)
	}
	cli.Close()
}

func (s *agentServer) sendRangeData(stream *pb.ProcessService_SendTraceDataClient, offset0, offset1 int64) {
	bs := s.GetRangeBufFromAbsoluteOffset(offset0, offset1)
	buffer := NewBuffer(bs)
	tosend := &pb.TraceData{Tracedata: nil}
	for {
		span, traceIdBytes, err := buffer.ReadLineWithTraceId()
		if err == io.EOF {
			break
		}
		//traceIdIndex = bytes.IndexByte(span, '|')
		traceId := bytesToInt64(traceIdBytes)
		_, ok := s.traceMap.Load(traceId)
		if ok {
			tosend.Tracedata = span
			(*stream).Send(tosend)
			s.sendNum++
		}
	}
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
		s.lock.Lock()
		if s.isReady == true {
			s.lock.Unlock()
			return
		}
		s.isReady = true
		s.lock.Unlock()
		for i := 0; i < s.opt.downn*2; i++ {
			s.wg.Add(1)
			go s.ReadRangeData()
		}

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
			for i := 0; i < s.opt.sendn*2; i++ {
				s.wgSend.Add(1)
				go s.SendRangeData()
			}
			for i := 0; i < s.opt.downn*2; i++ {
				s.wgCheck.Add(1)
				go s.CheckTargetData()
			}

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
