package main

import (
	"bufio"
	"bytes"
	"io"
	"io/ioutil"
	"log"
	"os"
	"sync"
)

type traceInfo struct {
	isTarget  bool     //是否命中上报规则,为true 表示已经命中
	firstTime int64    // 第一条span的starttime
	sum       int      //当前traceid累计记录数量
	cur       int      //已发送游标
	sendData  [][]byte //待向后端发送的数据
	hasFile   bool
	m         sync.Mutex
}

type filter struct {
	//traceMap  map[int64]*traceInfo //traceid数据映射表
	traceMap  sync.Map
	opt       *option
	cli       *processdCli
	curBufNum int //当前缓冲区内的未命中traceid数量
	m         sync.Mutex
}

func newFilter(opt *option) *filter {
	return &filter{
		sync.Map{},
		opt,
		newProcessdCli(opt),
		0,
		sync.Mutex{}}
}

//向tracemap中添加新traceid key
func (f *filter) addTraceidKey(traceid []byte, firstTime int64) *traceInfo {
	t := &traceInfo{
		false,
		firstTime,
		0,
		0,
		[][]byte{},
		false,
		sync.Mutex{}}
	info, _ := f.traceMap.LoadOrStore(traceidToUint64(traceid), t)
	return info.(*traceInfo)
}

func (t *traceInfo) sendAllCurrentDataToBuf(cli *processdCli) (b bool) {
	t.m.Lock()
	defer t.m.Unlock()
	b = true
	n := len(t.sendData)
	if n > 0 {
		for _, v := range t.sendData {
			if !cli.sendToBuffer(v) {
				b = false
				break
			}
			t.cur++
		}
		if t.cur > 0 {
			//		copy(t.sendData, t.sendData[t.cur:])
			//		t.sendData = t.sendData[:n-t.cur]
			t.sendData = t.sendData[t.cur:]
			t.cur = 0
		}
	}
	return
}

func (f *filter) handleSpan(span []byte, fields [][]byte, lruTraceidList []*Queue, rangeNum int) {
	// traceId :fields[0]  tags : fields[8]
	key := traceidToUint64(fields[0])
	info, exist := f.traceMap.Load(key)
	if !exist {
		info = f.addTraceidKey(fields[0], startTimeToInt64(fields[1]))
		f.curBufNum++
		lruTraceidList[rangeNum].Push(key)
		if f.curBufNum > f.opt.bufNum {
			useRangeNum := 0
			for {
				e := lruTraceidList[useRangeNum].Pop()
				if e == nil {
					useRangeNum++
					continue
				}
				f.curBufNum--
				tmpinfo, _ := f.traceMap.Load(e)
				if tmpinfo.(*traceInfo).isTarget {
					continue
				} else {
					info.(*traceInfo).sendData = tmpinfo.(*traceInfo).sendData[:0]
					tmpinfo.(*traceInfo).sendData = nil
					break
				}
			}
		}
	}
	info.(*traceInfo).sum++
	// if traceid is target ,send
	if !info.(*traceInfo).isTarget && checkIsTarget(fields[8]) {
		info.(*traceInfo).isTarget = true
		//traceid first target
		go f.cli.setTargetTraceidToProcessd(fields[0])
		/*if info.(*traceInfo).hasFile {
			log.Printf("unexpact,first target but need to load from file")
			info.(*traceInfo).loadFromFile(fields[0])
			info.(*traceInfo).hasFile = false
		}*/
	}
	if info.(*traceInfo).isTarget {
		if info.(*traceInfo).sendAllCurrentDataToBuf(f.cli) && f.cli.sendToBuffer(span) {
			return
		} else {
			info.(*traceInfo).sendData = append(info.(*traceInfo).sendData, span)
			log.Printf("sendChan blocked")
		}
	} else {
		/*if info.(*traceInfo).hasFile {
			log.Printf("unexpact ,traceid is not target ,need to write file")
			//TODO
		} else {*/
		if info.(*traceInfo).sendData != nil {
			info.(*traceInfo).sendData = append(info.(*traceInfo).sendData, span)
		}
		//	}
	}
}

//load to memory
func (t *traceInfo) loadFromFile(traceid []byte) {
	log.Printf("loadFromFile")
	b, err := ioutil.ReadFile("./tracedata/" + string(traceid) + ".tmp")
	if err != nil {
		log.Println(err)
	}
	buf := bytes.NewBuffer(b)
	for {
		span, err := buf.ReadBytes('\n')
		t.sendData = append(t.sendData, span)
		if err == io.EOF {
			return
		}
		if err != nil {
			log.Println(err)
			return
		}
	}
}

//flush to file
func (t *traceInfo) flushToTmpFile(traceid []byte) {
	if len(t.sendData) == 0 {
		return
	}
	log.Printf("flushToTmpFile,spannum:%d", len(t.sendData))
	file, err := os.OpenFile("./tracedata/"+string(traceid)+".tmp", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	defer file.Close()
	if err != nil {
		log.Printf("open file failed err:%v", err)
		file = nil
	}
	writer := bufio.NewWriter(file)
	for _, v := range t.sendData {
		writer.Write(v)
	}
	writer.Flush()
	t.sendData = t.sendData[:0]
	t.cur = 0
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

func traceidToUint64(a []byte) (u uint64) {
	var d byte
	for _, c := range a {
		switch {
		case '0' <= c && c <= '9':
			d = c - '0'
		case 'a' <= c && c <= 'z':
			d = c - 'a' + 10
		}
		u *= uint64(16)
		u += uint64(d)
	}
	return u
}

func startTimeToInt64(a []byte) (u int64) {
	for _, c := range a {
		u *= int64(16)
		u += int64(c - '0')
	}
	return u

}
