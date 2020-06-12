package main

import (
	"flag"
)

type option struct {
	port         string
	grpcPort     string
	dataFilename string
}

func main() {
	opt := new(option)
	flag.StringVar(&opt.port, "port", "", "the port of http server")
	flag.StringVar(&opt.grpcPort, "rpcport", "50000", "the port of grpc server")
	flag.StringVar(&opt.dataFilename, "filename", "trace1.data", "data file name")
	flag.Parse()

	var srv agentServer
	srv.initServer(opt)
	go srv.StartGrpcServer()
	go srv.StartHttpServer()
	srv.SignalHandle()
}
