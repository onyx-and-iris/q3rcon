package q3rcon

import "bytes"

const (
	responseHeader = "\xff\xff\xff\xffprint\n"
)

type response struct{}

func newResponse() response {
	return response{}
}

func (r response) IsValid(buf []byte) bool {
	return len(buf) > len(responseHeader) && bytes.HasPrefix(buf, []byte(responseHeader))
}

func (r response) Decode(buf []byte) string {
	return string(buf[len(responseHeader):])
}
