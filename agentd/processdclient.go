package main

import (
	"alitest2020_tailbase/agentd/pb"
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"time"
)

type processdCli struct {
	conn *grpc.ClientConn
	addr string //who to connect
	opt  *option
}

func NewProcessdCli(opt *option) *processdCli {
	return &processdCli{nil, opt.processdAddr, opt}
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
	conn, err := grpc.Dial(
		c.addr, grpc.WithInsecure(),
		grpc.WithKeepaliveParams(kacp))
	if err != nil {
		panic(err)
	}
	c.conn = conn
}
func (c *processdCli) GetStream() pb.ProcessService_SendTraceDataClient {
	c.Connect()
	client := pb.NewProcessServiceClient(c.conn)
	stream, err := client.SendTraceData(context.Background())
	if err != nil {
		panic(err)
	}
	return stream
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
