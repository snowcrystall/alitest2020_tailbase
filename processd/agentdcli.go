package main

import (
	"alitest2020_tailbase/pb"
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"log"
	"sync"
	"time"
)

type agentdCli struct {
	conn              *grpc.ClientConn
	addr              string
	NodifyTraceIdChan chan *peerInfo
	wg                *sync.WaitGroup
}

type peerInfo struct {
	checkCur int64
	traceId  []byte
}

func NewAgentdCli(addr string) *agentdCli {
	return &agentdCli{nil, addr, make(chan *peerInfo, 500), &sync.WaitGroup{}}
}
func (c *agentdCli) Connect() {
	if c.conn != nil {
		return
	}
	// Set up a connection to the server.
	var kacp = keepalive.ClientParameters{
		Time:                5 * time.Second,  // send pings every 10 seconds if there is no activity
		Timeout:             10 * time.Second, // wait 1 second for ping ack before considering the connection dead
		PermitWithoutStream: true,             // send pings even without active streams
	}
	conn, err := grpc.Dial(c.addr, grpc.WithInsecure(), grpc.WithKeepaliveParams(kacp))
	if err != nil {
		log.Fatalf("did not connect: %s : %v ", c.addr, err)
	}
	c.conn = conn
}

func (c *agentdCli) RunNodifyTraceId() {
	c.wg.Add(1)
	defer c.wg.Done()
	c.Connect()
	client := pb.NewAgentServiceClient(c.conn)
	stream, err := client.NotifyTargetTraceids(context.Background())
	if err != nil {
		panic(err)
	}

	for {
		traceIdInfo, ok := <-c.NodifyTraceIdChan
		if !ok {
			break
		}
		stream.Send(&pb.TargetInfo{Traceid: traceIdInfo.traceId, Checkcur: traceIdInfo.checkCur})
	}
	_, err = stream.CloseAndRecv()
	if err != nil {
		log.Panicf("%v.CloseAndRecv() got error %v, want %v", stream, err, nil)
	}

}

func (c *agentdCli) Close() {
	close(c.NodifyTraceIdChan)
	c.wg.Wait()
	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}
}
