package main

import (
	"bytes"
	"io"
)

type Buffer struct {
	buf []byte // contents are the bytes buf[off : len(buf)]
	off int    // read at &buf[off], write at &buf[len(buf)]
}

var target1 []byte = []byte("error=1")
var target2 []byte = []byte("http.status_code=")

func NewBuffer(buf []byte) *Buffer { return &Buffer{buf: buf} }

func (b *Buffer) ReadLine(delim byte) (line []byte, err error) {
	i := bytes.IndexByte(b.buf[b.off:], delim)
	end := b.off + i + 1
	if i < 0 {
		end = len(b.buf)
		err = io.EOF
	}
	line = b.buf[b.off:end]
	b.off = end
	return line, err
}

func (b *Buffer) ReadLineWithTraceId() (line []byte, traceId []byte, err error) {
	i := bytes.IndexByte(b.buf[b.off:], '|')
	if i < 0 {
		err = io.EOF
		return line, traceId, err
	}

	end := b.off + i
	traceId = b.buf[b.off:end]

	i = bytes.IndexByte(b.buf[end:], '\n')
	line = b.buf[b.off : end+i+1]
	b.off = end + i + 1
	return line, traceId, nil
}

/*func (b *Buffer) ReadLineWithCheck() (line []byte, traceId []byte, isTarget bool, err error) {
	i := bytes.IndexByte(b.buf[b.off:], '|')
	if i < 0 {
		err = io.EOF
		return line, traceId, err
	}

	end := b.off + i
	traceId = b.buf[b.off:end]

	i = bytes.IndexByte(b.buf[end+1:], '|')
	end = end + 1 + i

	i = bytes.IndexByte(b.buf[end:], '\n')
	line = b.buf[b.off : end+i+1]
	b.off = end + i + 1
	return line, traceId, nil
}
*/
/*func (b *Buffer) CheckIsTarget(d []byte) bool {
	i:=len(d)-1

	for  {
		if t1i < len(target1) && x == target1[t1i] {
			t1i++
		} else if t1i < len(target1) {
			t1i = 0
		} else {
			return true
		}
		if t2i < len(target2) && x == target2[t2i] {
			t2i++
		} else if t2i < len(target2) {
			t2i = 0
		} else {
			if !(x == 50 && d[j+1] == 48 && d[j+2] == 48) {
				return true
			} else {
				t2i = 0
			}
		}

	}
	return false
}*/
