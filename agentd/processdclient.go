package main

import (
	"alitest2020_tailbase/pb"
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/metadata"
	"log"
	"sync"
	"time"
)

type processdCli struct {
	conn *grpc.ClientConn
	addr string //who to connect
	opt  *option

	SendTargetTraceIdChan chan int64
	wg                    sync.WaitGroup
}

func NewProcessdCli(opt *option) *processdCli {
	p := &processdCli{}
	p.conn = nil
	p.addr = opt.processdAddr
	p.opt = opt
	p.SendTargetTraceIdChan = make(chan int64, 500)
	p.wg = sync.WaitGroup{}
	return p
}

func (c *processdCli) Connect() {
	if c.conn != nil {
		return
	}
	// Set up a connection to the server.
	var kacp = keepalive.ClientParameters{
		Time:                5 * time.Second,
		Timeout:             10 * time.Second,
		PermitWithoutStream: true,
	}

	retry := 10
	for {
		conn, err := grpc.Dial(
			c.addr, grpc.WithInsecure(),
			grpc.WithKeepaliveParams(kacp))
		if err != nil {
			log.Printf(err.Error())
			if retry < 0 {
				panic(err)
			}
			retry--
		}
		c.conn = conn
		break
	}
}
func (c *processdCli) GetSendDataStream() pb.ProcessService_SendTraceDataClient {
	c.Connect()
	client := pb.NewProcessServiceClient(c.conn)
	stream, err := client.SendTraceData(context.Background())
	if err != nil {
		panic(err)
	}
	return stream
}

func (c *processdCli) GetSendTargetIdStream() pb.ProcessService_SendTargetIdsClient {
	c.Connect()
	client := pb.NewProcessServiceClient(c.conn)
	//ctx := peer.NewContext(context.Background(), &peer.Peer{"localhost:" + c.opt.grpcPort, nil})
	ctx := metadata.NewOutgoingContext(context.Background(), metadata.MD{"addr": []string{"localhost:" + c.opt.grpcPort}})
	stream, err := client.SendTargetIds(ctx)
	if err != nil {
		panic(err)
	}
	return stream
}

func (c *processdCli) SendTargetTraceId() {
	/*c.wg.Add(1)
	defer c.wg.Done()
	for {
		traceId, ok := <-c.SendTargetTraceIdChan
		if !ok {
			break
		}
		c.SetTargetTraceidToProcessd(traceId)
	}*/
}

func (c *processdCli) SetTargetTraceidToProcessd(traceid []byte) {
	c.Connect()
	client := pb.NewProcessServiceClient(c.conn)
	_, err := client.SetTargetTraceid(context.Background(), &pb.TraceidRequest{Traceid: traceid})
	if err != nil {
		panic(err)
	}
}

func (c *processdCli) NotifySendOver() {
	c.Connect()
	client := pb.NewProcessServiceClient(c.conn)
	_, err := client.NotifySendOver(context.Background(), &pb.Addr{Addr: "localhost:" + c.opt.grpcPort})
	if err != nil {
		panic(err)
	}

}

func (c *processdCli) Close() {
	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}
}
