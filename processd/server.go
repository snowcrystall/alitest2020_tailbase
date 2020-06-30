package main

import (
	"alitest2020_tailbase/pb"
	"bufio"
	"bytes"
	"context"
	"crypto/md5"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/metadata"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

type processServer struct {
	opt *option
	pb.UnimplementedProcessServiceServer

	revChan      chan []byte //trace data chan
	traceDataMap sync.Map    //traceDataMap map[int64]*traceDataDesc //key为 traceid

	agentDone        chan int
	port             string
	ckStatus         []*checkStatus
	agentClis        []*agentdCli
	agentAddrs       []string
	agentPeerDoneNum int32
}

type checkStatus struct {
	startTime int64
	traceid   []byte
}
type traceDataDesc struct {
	num           int              // traceid的span累计数量
	startTimeList []int64          //startTime 有序列表, 新增时做有序插入
	traceData     map[int64][]byte //key是startTime,value是一条span日志
}

func (s *processServer) initServer(opt *option) {
	s.opt = opt
	s.revChan = make(chan []byte, 2000)
	s.agentDone = make(chan int)
	s.ckStatus = []*checkStatus{}
	s.agentClis = []*agentdCli{NewAgentdCli("localhost:50000"), NewAgentdCli("localhost:50001")}
	s.agentAddrs = []string{"localhost:50000", "localhost:50001"}
	s.agentPeerDoneNum = 0
}

func (s *processServer) broadcastNotifyAllFilterDone() {
	for _, cli := range s.agentClis {
		cli.Connect()
		client := pb.NewAgentServiceClient(cli.conn)
		_, err := client.NotifyAllFilterOver(context.Background(), &pb.Req{Req: []byte("ok")})
		if err != nil {
			log.Printf("could not greet: %v", err)
		}
	}
}

//s.agentPeer 为2个时说明两个agentd都过滤完了
func (s *processServer) NotifyFilterOver(ctx context.Context, in *pb.Addr) (*pb.Reply, error) {
	return &pb.Reply{Reply: []byte("ok")}, nil
}

//s.agentPeer 都为1时说明两个agentd都发送完了
func (s *processServer) NotifySendOver(ctx context.Context, in *pb.Addr) (*pb.Reply, error) {
	s.agentPeerDoneNum = atomic.AddInt32(&s.agentPeerDoneNum, 1)
	if s.agentPeerDoneNum == 2 {
		close(s.agentDone)
	}
	return &pb.Reply{Reply: []byte("ok")}, nil
}
func insertCheckStatusList(list []*checkStatus, item *checkStatus) []*checkStatus {
	list = append(list, item)
	i := len(list) - 1
	for i > 0 {
		if item.startTime < list[i-1].startTime {
			list[i] = list[i-1]
			i--
		} else {
			break
		}
	}
	list[i] = item
	return list
}

func (s *processServer) flushDataTofile(n int, m int) {
	s.traceDataMap.Range(func(traceid interface{}, desc interface{}) bool {
		len := len(desc.(*traceDataDesc).startTimeList)
		if len > n {
			if desc.(*traceDataDesc).num == len {
				tkStatus := &checkStatus{desc.(*traceDataDesc).startTimeList[0], []byte(strconv.FormatUint(traceid.(uint64), 16))}
				s.ckStatus = insertCheckStatusList(s.ckStatus, tkStatus)
			}

			file, err := os.OpenFile("./tracedata/"+strconv.FormatUint(traceid.(uint64), 16)+".data", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				log.Printf("open file failed err:%v", err)
				file.Close()
				return false
			}
			writer := bufio.NewWriter(file)
			for _, st := range desc.(*traceDataDesc).startTimeList[0 : len-m] {
				writer.Write(desc.(*traceDataDesc).traceData[st])
				delete(desc.(*traceDataDesc).traceData, st)
			}
			desc.(*traceDataDesc).startTimeList = desc.(*traceDataDesc).startTimeList[len-m:]
			writer.Flush()
			file.Close()

			//log.Printf("%s, write to file ,%d", traceid.(string), len-m)
		}
		return true
	})
}

func (s *processServer) checksum() {
	json := []byte{'{'}
	s.traceDataMap.Range(func(traceid interface{}, desc interface{}) bool {
		json = append(json, '"')
		json = append(json, []byte(fmt.Sprintf("%x", traceid))...)
		json = append(json, "\":\""...)

		md5hash := md5.New()
		for _, st := range desc.(*traceDataDesc).startTimeList {
			md5hash.Write(desc.(*traceDataDesc).traceData[st])
		}
		json = append(json, fmt.Sprintf("%X", md5hash.Sum(nil))...)
		json = append(json, "\","...)
		return true
	})
	json = json[:len(json)-1]
	json = append(json, '}')
	v := url.Values{}
	v.Set("result", string(json[:]))
	client := &http.Client{}
	res, err := client.PostForm("http://localhost:"+s.port+"/api/finished", v)
	log.Println(res, err)
	file, _ := os.OpenFile("./tracedata/checksum.data", os.O_CREATE|os.O_WRONLY, 0644)
	writer := bufio.NewWriter(file)
	writer.Write(json)
	writer.Flush()
	file.Close()
}

func (s *processServer) runSaveTraceDataToFile(ctx context.Context) {
	for {
		s.flushDataTofile(30, 20)
		select {
		case <-ctx.Done():
			s.flushDataTofile(0, 0)
			//所有的agent都发送完了.计算checksum并上报
			log.Printf("agent done ,flush all  data to file")
			s.checksum()
			return
		case <-time.After(time.Second):
			//log.Printf("wait 1 second to flush data to file")
		}
	}
}
func insertKeyWithOrderd(keylist []int64, key int64) []int64 {
	//key 是startTime,从过滤服务接收过来的数据大致是有序的，所以采用尾部比较插入
	//去重
	l := 0
	r := len(keylist) - 1
	for l <= r {
		mid := (l + r) >> 1
		if keylist[mid] == key {
			return keylist
			r = mid - 1
		} else {
			l = mid + 1
		}
	}
	keylist = append(keylist, key)
	i := len(keylist) - 1
	for i > 0 {
		if key < keylist[i-1] {
			keylist[i] = keylist[i-1]
			i--
		} else {
			break
		}
	}
	keylist[i] = key
	return keylist
}
func (s *processServer) runProcessData() {
	n := 0
	log.Printf("start runProcessData")
	//ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		//	cancel() // cancel when we are finished consuming integers
		log.Printf("end runProcessData,total span :%d", n)
		s.checksum()
		s.flushDataTofile(0, 0)
	}()
	//go s.runSaveTraceDataToFile(ctx)
	var span []byte
	var ok bool
	for {
		select {
		case span, ok = <-s.revChan:
			n++
			s.handleSpan(span)
			continue
		default:
		}
		select {
		case span, ok = <-s.revChan:
			n++
			s.handleSpan(span)
			continue
		case _, ok = <-s.agentDone:
			if !ok {
				log.Printf("agent done")
			}
		}
		break
	}
}
func (s *processServer) handleSpan(span []byte) {
	fields := bytes.SplitN(span, []byte("|"), 3) //traceid:fields[0],startTime:fields[1]
	if len(fields) < 3 {
		log.Printf("unexpact: %s", string(span))
		return
	}
	startTime, _ := strconv.ParseInt(string(fields[1]), 10, 64)
	key, _ := strconv.ParseUint(string(fields[0]), 16, 64)
	tdDesc, ok := s.traceDataMap.Load(key)
	if !ok {
		tdDesc = &traceDataDesc{0, []int64{}, make(map[int64][]byte)}
	}
	startTimeKeys := tdDesc.(*traceDataDesc).startTimeList
	//累计数据总数sum 大于 .startTimeList的数量，说明已经有数据写入文件，
	// startTime 比第一个元素小，说明需要在文件中插入，需要额外处理
	/*	if tdDesc.(*traceDataDesc).num > len(startTimeKeys) && (startTime < startTimeKeys[0]) {
			log.Printf("need insert to file")
			//TODO
		} else {
			tdDesc.(*traceDataDesc).num++
			tdDesc.(*traceDataDesc).traceData[startTime] = span
			tdDesc.(*traceDataDesc).startTimeList = insertKeyWithOrderd(startTimeKeys, startTime)
		}*/
	tdDesc.(*traceDataDesc).num++
	tdDesc.(*traceDataDesc).traceData[startTime] = span
	tdDesc.(*traceDataDesc).startTimeList = insertKeyWithOrderd(startTimeKeys, startTime)

	s.traceDataMap.Store(key, tdDesc)
}

func (s *processServer) SendTraceData(gs pb.ProcessService_SendTraceDataServer) error {
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
		s.revChan <- in.Tracedata
	}

	return nil
}

func (s *processServer) SendTargetIds(gs pb.ProcessService_SendTargetIdsServer) error {
	var nodifyAddr string
	if md, ok := metadata.FromIncomingContext(gs.Context()); ok {
		for _, v := range s.agentAddrs {
			if v != md["addr"][0] {
				nodifyAddr = v
				break
			}
		}
	}
	nodifyAgentCli := NewAgentdCli(nodifyAddr)
	go nodifyAgentCli.RunNodifyTraceId()
	for {
		in, err := gs.Recv()
		if err == io.EOF {
			gs.SendAndClose(&pb.Reply{Reply: []byte("ok")})
			break
		}
		if err != nil {
			log.Printf("failed to recv: %v", err)
			//	break
		}
		nodifyAgentCli.NodifyTraceIdChan <- &peerInfo{in.Checkcur, in.Traceid}
	}
	nodifyAgentCli.Close()
	return nil
}

func (s *processServer) StartGrpcServer() {
	lis, err := net.Listen("tcp", ":"+s.opt.grpcPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	var kasp = keepalive.ServerParameters{
		Time:    5 * time.Second,  // Ping the client if it is idle for 5 seconds to ensure the connection is still active
		Timeout: 10 * time.Second, // Wait 1 second for the ping ack before assuming the connection is dead
	}
	srv := grpc.NewServer(grpc.KeepaliveParams(kasp))
	pb.RegisterProcessServiceServer(srv, s)
	if err := srv.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
func (s *processServer) StartHttpServer() {
	http.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {})
	http.HandleFunc("/setParameter", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		port := r.Form.Get("port")
		if port != "" {
			s.port = port
		}
	})
	log.Fatal(http.ListenAndServe(":"+s.opt.port, nil))

}
func (s *processServer) SignalHandle() {
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
