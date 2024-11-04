package packet

import "fmt"

type Request struct {
	magic    []byte
	password string
}

func NewRequest(password string) Request {
	return Request{
		magic:    []byte{'\xff', '\xff', '\xff', '\xff'},
		password: password,
	}
}

func (r Request) Header() []byte {
	return append(r.magic, []byte("rcon")...)
}

func (r Request) Encode(cmd string) []byte {
	datagram := r.Header()
	datagram = append(datagram, fmt.Sprintf(" %s %s", r.password, cmd)...)
	return datagram
}
