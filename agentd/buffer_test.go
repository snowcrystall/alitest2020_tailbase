package main

import (
	"io"
	"log"
	"testing"
)

func TestCheckIsTarget(t *testing.T) {
	a := []byte("123234|435df|akdkdjfk|sdf\nasdf|sdfsdf|error=1sdfsd\n324534|asdfsdf|asdfasdf|sdfhttp.status_code=200\n234sdkfj|http.status_code=502\nsadfsdf234|http.status_code=400\n")
	buffer := NewBuffer(a)

	for {
		line, err := buffer.ReadLine('\n')
		log.Printf("%s,%t", line, buffer.CheckIsTarget(line))
		if err == io.EOF {
			break
		}
	}
}

func TestReadLineWithTraceId(t *testing.T) {
	a := []byte("123234|435df|akdkdjfk|sdf\nasdf|sdfsdf|error=1sdfsd\n324534|asdfsdf|asdfasdf|sdfhttp.status_code=200\n234sdkfj|http.status_code=502\nsadfsdf234|http.status_code=400\n")
	buffer := NewBuffer(a)

	for {
		line, traceid, err := buffer.ReadLineWithTraceId()
		if err == io.EOF {
			break
		}
		log.Printf("%s,%s", line, traceid)

	}
}
