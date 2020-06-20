package main

import (
	"bytes"
	"io"
)

type Buffer struct {
	buf []byte // contents are the bytes buf[off : len(buf)]
	off int    // read at &buf[off], write at &buf[len(buf)]
}

func NewBuffer(buf []byte) *Buffer { return &Buffer{buf: buf} }

func (b *Buffer) ReadSlice(delim []byte) (line []byte, err error) {
	i := bytes.Index(b.buf[b.off:], delim)
	end := b.off + i + 1
	if i < 0 {
		end = len(b.buf)
		err = io.EOF
	}
	line = b.buf[b.off:end]
	b.off = end
	return line, err
}
