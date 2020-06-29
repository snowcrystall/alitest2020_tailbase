package main

import (
	"bytes"
	"log"
	"os"
	"runtime"
	"runtime/pprof"
)

func bytesToInt64(a []byte) (u int64) {
	for _, c := range a {
		u *= int64(16)
		switch {
		case '0' <= c && c <= '9':
			u += int64(c - '0')
		case 'a' <= c && c <= 'z':
			u += int64(c - 'a' + 10)
		}
	}
	return u
}

/*func checkIsTarget(tag []byte) bool {
	// 找到所有tags中存在 http.status_code 不为 200
	n := len(tag)
	if bytes.Equal(tag[n-21:n-4], []byte("http.status_code=")) {
		if !(tag[n-4] == 50 && tag[n-3] == 48 && tag[n-2] == 48) {
			return true
		}
	}

	//判断error 等于1的调用链路
	if bytes.Equal(tag[n-8:n-1], []byte("error=1")) {
		return true
	}
	return false
}*/

func checkIsTarget(tag []byte) bool {
	//判断error 等于1的调用链路
	if bytes.Contains(tag, []byte("error=1")) {
		return true
	}
	// 找到所有tags中存在 http.status_code 不为 200
	if index := bytes.Index(tag, []byte("http.status_code=")); index != -1 {
		if !(tag[index+17] == 50 && tag[index+18] == 48 && tag[index+19] == 48) {
			return true
		}
	}
	return false
}
func pprofMemory() {
	f, err := os.Create("memory.ppof")
	if err != nil {
		log.Fatal("could not create memory profile: ", err)
	}
	defer f.Close() // error handling omitted for example
	//runtime.GC()    // get up-to-date statistics
	if err := pprof.WriteHeapProfile(f); err != nil {
		log.Fatal("could not write memory profile: ", err)
	}

}

func PrintMemUsage() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	log.Printf("\tTotalAlloc = %v MiB", bToMb(m.TotalAlloc))
	log.Printf("\tSys = %v MiB", bToMb(m.Sys))
	log.Printf("\tNumGC = %v\n", m.NumGC)
}

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}
