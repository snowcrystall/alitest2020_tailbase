package main

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"log"
	"time"
)

type agentdCli struct {
	conn *grpc.ClientConn
	addr string
}

func newAgentdCli(addr string) *agentdCli {
	return &agentdCli{nil, addr}
}
func (c *agentdCli) connect() {
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

func (c *agentdCli) close() {
	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}
}
