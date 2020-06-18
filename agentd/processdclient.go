package main

import (
	"alitest2020/agentd/pb"
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"log"
	"time"
)

type processdCli struct {
	conn     *grpc.ClientConn
	addr     string //who to connect
	sendChan chan []byte
	opt      *option
}

func newProcessdCli(opt *option) *processdCli {
	return &processdCli{nil, opt.processdAddr, make(chan []byte, 2048), opt}
}

func (c *processdCli) connect() {
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
		log.Fatalf("did not connect: %s : %v ", c.addr, err)
	}
	c.conn = conn
}
func (c *processdCli) getStream() pb.ProcessService_SendTraceDataClient {
	c.connect()
	client := pb.NewProcessServiceClient(c.conn)
	stream, err := client.SendTraceData(context.Background())
	if err != nil {
		log.Fatalf("could not get stream: %v ", err)
	}
	return stream
}

func (c *processdCli) setTargetTraceidToProcessd(traceid []byte) {
	c.connect()
	client := pb.NewProcessServiceClient(c.conn)
	_, err := client.SetTargetTraceid(context.Background(), &pb.TraceidRequest{Traceid: traceid})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
}

func (c *processdCli) notifyFilterOver() {
	c.connect()
	client := pb.NewProcessServiceClient(c.conn)
	_, err := client.NotifyFilterOver(context.Background(), &pb.Addr{Addr: "localhost:" + c.opt.grpcPort})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
}

func (c *processdCli) notifySendOver() {
	c.connect()
	client := pb.NewProcessServiceClient(c.conn)
	_, err := client.NotifySendOver(context.Background(), &pb.Addr{Addr: "localhost:" + c.opt.grpcPort})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}

}

func (c *processdCli) close() {
	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}
}
