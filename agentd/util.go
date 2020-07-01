package main

import (
	"bytes"
	"log"
	"os"
	"regexp"
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

func checkIsTargetV2(tag []byte) bool {
	re := regexp.MustCompile(`error=1|http.status_code=\d{3}`)
	s := re.FindSubmatch(tag)
	if len(s) == 0 {
		return false
	} else {
		m := s[0]
		if len(m) == 7 {
			return true
		} else if bytes.Equal(m[len(m)-3:], []byte("200")) {
			if index := bytes.LastIndex(tag, []byte("error=1")); index != -1 {
				return true
			}

		} else {
			return true
		}
	}
	return false
}

func checkIsTarget(tag []byte) bool {
	//判断error 等于1的调用链路
	if index := bytes.Index(tag, []byte("error=1")); index != -1 {
		return true
	}
	// 找到所有tags中存在 http.status_code 不为 200
	if index := bytes.Index(tag, []byte("http.status_code=")); index != -1 {
		if !bytes.Equal(tag[index+17:index+20], []byte("200")) {
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
