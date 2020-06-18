package main

import (
	"alitest2020/agentd/pb"
	"bytes"
	"log"
	"sync"
)

type filter struct {
	traceMap sync.Map
	cli      *processdCli
}

type rangeDataFilter struct {
	buf          [][]byte
	checkCur     int
	checkOver    bool
	sendCur      int
	checkCurCond *sync.Cond
	sendCurCond  *sync.Cond
	f            *filter
	m            sync.Mutex
}

func newRangeFilter(opt *option, f *filter) *rangeDataFilter {
	return &rangeDataFilter{
		make([][]byte, 100*opt.bufn/opt.fn),
		0,
		false,
		0,
		sync.NewCond(&sync.Mutex{}),
		sync.NewCond(&sync.Mutex{}),
		f,
		sync.Mutex{}}
}

func (r *rangeDataFilter) checkSpan(span []byte, fields [][]byte) {
	// traceId :fields[0]  tags : fields[8]
	key := bytesToInt64(fields[0])
	if checkIsTarget(fields[8]) {
		r.f.traceMap.LoadOrStore(key, true)
		go r.f.cli.setTargetTraceidToProcessd(fields[0])
	}
	if r.checkCur != 0 && r.checkCur%len(r.buf) == r.sendCur%len(r.buf) {

		log.Println("checkCur wait ", r.checkCur, r.sendCur)
		r.sendCurCond.L.Lock()
		r.sendCurCond.Signal()
		r.sendCurCond.L.Unlock()

		r.checkCurCond.L.Lock()
		r.checkCurCond.Wait()
		r.checkCurCond.L.Unlock()
	}
	r.buf[r.checkCur%len(r.buf)] = span
	r.checkCur++
}

func (r *rangeDataFilter) runSendData(wgSend *sync.WaitGroup) {
	stream := r.f.cli.getStream()
	defer func() {
		stream.CloseSend()
		wgSend.Done()
	}()

	var span []byte
	var ok bool
	var traceid int64
	var spanNum int
	for {
		if !r.checkOver && r.checkCur-r.sendCur <= 1000000 {
			log.Println("sendCur wait ", r.checkCur, r.sendCur)
			r.checkCurCond.L.Lock()
			r.checkCurCond.Signal()
			r.checkCurCond.L.Unlock()

			r.sendCurCond.L.Lock()
			r.sendCurCond.Wait()
			r.sendCurCond.L.Unlock()
		}
		span = r.buf[r.sendCur%len(r.buf)]
		traceid = bytesToInt64(span[:bytes.Index(span, []byte{'|'})])
		_, ok = r.f.traceMap.Load(traceid)
		if ok {
			spanNum++
			stream.Send(&pb.TraceData{Tracedata: span})
		}
		r.sendCur++

		if r.checkOver && r.sendCur == r.checkCur {
			log.Println(" runSendData send over ", spanNum)
			break
		}
	}
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
