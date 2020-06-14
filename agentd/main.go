package main

import (
	"flag"
)

type option struct {
	port         string
	grpcPort     string
	dataFilename string
	processdAddr string
	eachBufNum   int
	bufNum       int
	bufTime      int64
	debug        int
}

func main() {
	opt := new(option)
	flag.StringVar(&opt.port, "port", "", "the port of http server")
	flag.StringVar(&opt.grpcPort, "rpcport", "50000", "the port of grpc server")
	flag.StringVar(&opt.dataFilename, "filename", "trace1.data", "data file name")
	flag.StringVar(&opt.processdAddr, "processdAddr", "localhost:50002", "")
	//ebufn unuse
	flag.IntVar(&opt.eachBufNum, "ebufn", 128, "")
	flag.IntVar(&opt.bufNum, "bufn", 20000, "")
	flag.Int64Var(&opt.bufTime, "bufsec", 30, "")
	flag.IntVar(&opt.debug, "debug", 0, "")
	flag.Parse()

	var srv agentServer
	srv.initServer(opt)
	go srv.StartGrpcServer()
	go srv.StartHttpServer()
	srv.SignalHandle()
}
