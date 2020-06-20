package main

import (
	"flag"
)

type option struct {
	port         string
	grpcPort     string
	dataFilename string
	processdAddr string
	downn        int
	sendn        int
}

func main() {
	opt := new(option)
	flag.StringVar(&opt.port, "port", "", "the port of http server")
	flag.StringVar(&opt.grpcPort, "rpcport", "50000", "the port of grpc server")
	flag.StringVar(&opt.dataFilename, "filename", "trace1.data", "data file name")
	//flag.StringVar(&opt.processdAddr, "processdAddr", "localhost:50002", "")
	opt.processdAddr = "localhost:50002"
	flag.IntVar(&opt.downn, "downn", 1, "")
	flag.IntVar(&opt.sendn, "sendn", 1, "")
	//flag.Int64Var(&opt.bufTime, "bufsec", 30, "")
	flag.Parse()

	var srv agentServer
	srv.initServer(opt)
	go srv.StartGrpcServer()
	go srv.StartHttpServer()
	srv.SignalHandle()
}
