package packet

import (
	"bytes"
	"fmt"
)

const bufSz = 512

type Request struct {
	magic    []byte
	password string
	buf      *bytes.Buffer
}

func NewRequest(password string) Request {
	return Request{
		magic:    []byte{'\xff', '\xff', '\xff', '\xff'},
		password: password,
		buf:      bytes.NewBuffer(make([]byte, bufSz)),
	}
}

func (r Request) Header() []byte {
	return append(r.magic, []byte("rcon")...)
}

func (r Request) Encode(cmd string) []byte {
	r.buf.Reset()
	r.buf.Write(r.Header())
	r.buf.WriteString(fmt.Sprintf(" %s %s", r.password, cmd))
	return r.buf.Bytes()
}
