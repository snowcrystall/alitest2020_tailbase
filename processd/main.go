package main

import (
	"flag"
)

type option struct {
	port     string
	grpcPort string
}

func main() {
	opt := new(option)
	flag.StringVar(&opt.port, "port", "", "the port of http server")
	flag.StringVar(&opt.grpcPort, "rpcport", "50002", "the port of grpc server")
	flag.Parse()

	var srv processServer
	srv.initServer(opt)
	go srv.StartGrpcServer()
	go srv.StartHttpServer()
	go srv.runProcessData()
	srv.SignalHandle()
}
