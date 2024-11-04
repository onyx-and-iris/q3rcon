package packet

type Response struct {
	magic []byte
}

func NewResponse() Response {
	return Response{magic: []byte{'\xff', '\xff', '\xff', '\xff'}}
}

func (r Response) Header() []byte {
	return append(r.magic, []byte("print\n")...)
}
