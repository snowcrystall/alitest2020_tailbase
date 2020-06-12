package main

import (
	"context"
	"log"
	"strings"
	"sync"
	"time"
)

type traceInfo struct {
	isTarget   bool     //是否命中上报规则,为true 表示已经命中
	createTime int64    // 创建map key时的时间戳 time.Now().Unix()
	sum        int      //当前traceid累计记录数量
	cur        int      //已发送游标
	sendData   []string //待向后端发送的数据
}

type filter struct {
	m        sync.Mutex  //互斥锁
	spanchan chan string //待过滤数据通道
	//traceMap  map[string]*traceInfo //traceid数据映射表
	traceMap  sync.Map
	cli       *processdCli //向后端发送数据的客户端
	doneChan  chan int     //过滤完成时关闭
	isSending bool         // 是否在发送中
}

func newFilter(opt *option) *filter {
	return &filter{sync.Mutex{}, make(chan string, 1000), sync.Map{}, newProcessdCli("localhost:50002", opt), make(chan int), false}
}

//向tracemap中添加新traceid key
func (f *filter) addTraceidKey(traceid string) *traceInfo {
	t := &traceInfo{false, time.Now().Unix(), 0, 0, []string{}}
	info, _ := f.traceMap.LoadOrStore(traceid, t)
	return info.(*traceInfo)
}

//从cur发送当前所有数据
func (t *traceInfo) sendCurrentTraceData(sendchan chan string) *traceInfo {
	for {
		len := len(t.sendData)
		if len > t.cur {
			select {
			case sendchan <- t.sendData[t.cur]:
				t.cur++
			case <-time.After(time.Second * 1):
				log.Printf("wait 1 second,sendchan is blocked, can't push data in it")
			}
		} else {
			if t.cur > 0 {
				t.sendData = t.sendData[t.cur:]
				t.cur = 0
			}
			return t
		}
	}
}
func (f *filter) runSendData(ctx context.Context, sendchan chan string) {
	f.m.Lock()
	if f.isSending {
		return
	} else {
		f.isSending = true
		defer func() { f.isSending = false }()
	}
	f.m.Unlock()
	for {
		f.traceMap.Range(func(traceid, info interface{}) bool {
			if info.(*traceInfo).sum > 19999 {
				//同个traceid数据不会超过2万行
				return true
			}
			if info.(*traceInfo).isTarget {
				info = info.(*traceInfo).sendCurrentTraceData(sendchan)
				f.traceMap.Store(traceid, info)
			}
			return true
		})
		select {
		case <-ctx.Done():
			log.Printf("has send all  data to sendchan ")
			close(f.doneChan)
			return
		default:
			<-time.After(time.Second)
		}
	}
}

//过滤数据
func (f *filter) runFilterdata() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var span string
	var fields []string
	var traceid string
	var exist bool
	var info interface{}
	go f.cli.runSendTraceData(f.doneChan)
	for {
		select {
		case span = <-f.spanchan:
			if span == "EOF" {
				return
			}
			fields = strings.Split(span, "|")
			// traceId | startTime | spanId | parentSpanId | duration | serviceName | spanName | host | tags
			traceid = fields[0]
			info, exist = f.traceMap.Load(traceid)
			if !exist {
				info = f.addTraceidKey(traceid)
			}
			info.(*traceInfo).sendData = append(info.(*traceInfo).sendData, span)
			info.(*traceInfo).sum++
			// 如果该traceid数据已经被标记上报，就跳过检查
			if info.(*traceInfo).isTarget {
				break
			}
			if checkIsTarget(fields[8]) {
				//traceid 第一次命中，开始发送数据,并通知后端traceid
				//log.Printf("%s , first target ", traceid)
				info.(*traceInfo).isTarget = true
				go f.cli.setTargetTraceidToProcessd(traceid)
				go f.runSendData(ctx, f.cli.sendchan)
			}
		case <-time.After(time.Second):
			log.Printf("wait 1 second for http get data ")
		}
	}
}
func checkIsTarget(tag string) bool {
	//判断error 等于1的调用链路
	if strings.Contains(tag, "error=1") {
		return true
	}
	// 找到所有tags中存在 http.status_code 不为 200
	if index := strings.Index(tag, "http.status_code="); index != -1 {
		if tag[index+17:index+20] != "200" {
			return true
		}
	}
	return false
}
