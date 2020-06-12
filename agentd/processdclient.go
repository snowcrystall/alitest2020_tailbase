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
	addr     string
	sendchan chan string
	opt      *option
}

func newProcessdCli(addr string, opt *option) *processdCli {
	return &processdCli{nil, addr, make(chan string, 2000), opt}
}

func (c *processdCli) connect() {
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
func (c *processdCli) runSendTraceData(doneChan chan int) {
	c.connect()
	client := pb.NewProcessServiceClient(c.conn)
	stream, err := client.SendTraceData(context.Background())
	if err != nil {
		log.Fatalf("failed to call: %v", err)
	}

	for {
		select {
		case span := <-c.sendchan:
			stream.Send(&pb.TraceData{Tracedata: span})
		default:
			select {
			case _, ok := <-doneChan:
				if !ok {
					log.Printf("exit , send finish")
					c.notifySendOver()
					return
				}
			default:
				//				log.Printf("wait 1 second, sendchan empty ")
				<-time.After(time.Second * 1)
			}

		}
	}
}
func (c *processdCli) setTargetTraceidToProcessd(traceid string) {
	c.connect()
	client := pb.NewProcessServiceClient(c.conn)
	_, err := client.SetTargetTraceid(context.Background(), &pb.TraceidRequest{Traceid: traceid})
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
