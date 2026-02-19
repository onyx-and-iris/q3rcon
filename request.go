package q3rcon

import (
	"bytes"
	"errors"
	"fmt"
)

const (
	bufSz         = 1024
	requestHeader = "\xff\xff\xff\xffrcon"
)

type request struct {
	password string
	buf      *bytes.Buffer
}

func newRequest(password string) request {
	return request{
		password: password,
		buf:      bytes.NewBuffer(make([]byte, 0, bufSz)),
	}
}

func (r request) Encode(cmd string) ([]byte, error) {
	if cmd == "" {
		return nil, errors.New("command cannot be empty")
	}

	r.buf.Reset()
	r.buf.WriteString(requestHeader)
	r.buf.WriteString(fmt.Sprintf(" %s %s", r.password, cmd))
	return r.buf.Bytes(), nil
}
