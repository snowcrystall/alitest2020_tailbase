package main

import (
	"flag"
)

type option struct {
	port         string
	grpcPort     string
	dataFilename string
	processdAddr string
	bufn         int
	fn           int
	//bufTime      int64
}

func main() {
	opt := new(option)
	flag.StringVar(&opt.port, "port", "", "the port of http server")
	flag.StringVar(&opt.grpcPort, "rpcport", "50000", "the port of grpc server")
	flag.StringVar(&opt.dataFilename, "filename", "trace1.data", "data file name")
	//flag.StringVar(&opt.processdAddr, "processdAddr", "localhost:50002", "")
	opt.processdAddr = "localhost:50002"
	flag.IntVar(&opt.bufn, "bufn", 20000, "")
	flag.IntVar(&opt.fn, "fn", 2, "")
	//flag.Int64Var(&opt.bufTime, "bufsec", 30, "")
	flag.Parse()

	var srv agentServer
	srv.initServer(opt)
	go srv.StartGrpcServer()
	go srv.StartHttpServer()
	srv.SignalHandle()
}
