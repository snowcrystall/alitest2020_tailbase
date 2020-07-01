package main

import (
	"flag"
	"log"
	"os"
	"runtime/pprof"
)

type option struct {
	port         string
	grpcPort     string
	dataFilename string
	processdAddr string
	downn        int
	sendn        int
	debug        int
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
	flag.IntVar(&opt.debug, "debug", 0, "")
	flag.Parse()

	if opt.debug == 1 {
		f := startDebug()
		defer stopDebug(f)
	}
	srv := NewAgentServer(opt)
	go srv.StartGrpcServer()
	go srv.StartHttpServer()

	srv.SignalHandle()
}

func startDebug() *os.File {
	f, err := os.Create("cpu.ppof")
	if err != nil {
		log.Fatal("could not create CPU profile: ", err)
	}
	if err := pprof.StartCPUProfile(f); err != nil {
		log.Fatal("could not start CPU profile: ", err)
	}
	return f

}
func stopDebug(cpuf *os.File) {

	pprof.StopCPUProfile()
	PrintMemUsage()
	cpuf.Close()
}
