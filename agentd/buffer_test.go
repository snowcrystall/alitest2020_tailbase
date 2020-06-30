package main

import (
	"io"
	"log"
	"testing"
)

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

func TestReadLineWithTag(t *testing.T) {
	a := []byte("1|2|3|4|5|6|7|8|esdfrror=1\n2|2|3|4|5|6|7|8|http.status_code=200\n3|2|3|4|5|6|7|8|error=1asdf\n4|2|3|4|5|6|7|8|http.status_code=400\n5|2|3|4|5|6|7|8|error=1\n6|2|3|4|5|6|7|8|asdfhttp.status_code=500\n")
	buffer := NewBuffer(a)

	for {
		line, traceid, tag, err := buffer.ReadLineWithTag()
		if err == io.EOF {
			break
		}
		log.Printf("%s,%s,%s", line, traceid, tag)

	}
}
