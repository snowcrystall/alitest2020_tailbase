package main

import (
	"alitest2020/processd/pb"
	"context"
	"google.golang.org/grpc"
	"log"
	"strings"
	"time"
)

type agentdCli struct {
	addrs []string
}

func newAgentdCli(addr string) *agentdCli {
	return &agentdCli{strings.Split(addr, ",")}
}

func (c *agentdCli) broadcastNotifyTargetTraceid(traceid string) {
	for _, addr := range c.addrs {
		// Set up a connection to the server.
		conn, err := grpc.Dial(addr, grpc.WithInsecure())
		if err != nil {
			log.Fatalf("did not connect: %s : %v ", addr, err)
			break
		}
		client := pb.NewAgentServiceClient(conn)
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_, err = client.NotifyTargetTraceid(ctx, &pb.TraceidRequest{Traceid: traceid})
		if err != nil {
			log.Fatalf("could not greet: %v", err)
		}
	}
}
