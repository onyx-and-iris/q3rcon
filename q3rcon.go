package q3rcon

import (
	"bytes"
	"errors"
	"net"
	"strings"
	"time"

	"github.com/onyx-and-iris/q3rcon/internal/conn"
	"github.com/onyx-and-iris/q3rcon/internal/packet"
)

const respBufSiz = 2048

type Rcon struct {
	conn     conn.UDPConn
	request  packet.Request
	response packet.Response

	loginTimeout   time.Duration
	defaultTimeout time.Duration
	timeouts       map[string]time.Duration
}

func New(host string, port int, password string, options ...Option) (*Rcon, error) {
	if password == "" {
		return nil, errors.New("no password provided")
	}

	conn, err := conn.New(host, port)
	if err != nil {
		return nil, err
	}

	r := &Rcon{
		conn:     conn,
		request:  packet.NewRequest(password),
		response: packet.NewResponse(),

		loginTimeout:   5 * time.Second,
		defaultTimeout: 20 * time.Millisecond,
		timeouts:       make(map[string]time.Duration),
	}

	for _, o := range options {
		o(r)
	}

	if err = r.Login(); err != nil {
		return nil, err
	}

	return r, nil
}

func (r Rcon) Login() error {
	timeout := time.After(r.loginTimeout)
	for {
		select {
		case <-timeout:
			return errors.New("timeout logging in")
		default:
			resp, err := r.Send("login")
			if err != nil {
				return err
			}
			if resp == "" {
				continue
			}

			if strings.Contains(resp, "Bad rcon") {
				return errors.New("bad rcon password provided")
			} else {
				return nil
			}
		}
	}
}

func (r Rcon) Send(cmd string) (string, error) {
	timeout, ok := r.timeouts[cmd]
	if !ok {
		timeout = r.defaultTimeout
	}

	respChan := make(chan string)
	errChan := make(chan error)

	go r.listen(timeout, respChan, errChan)

	_, err := r.conn.Write(r.request.Encode(cmd))
	if err != nil {
		return "", err
	}

	select {
	case err := <-errChan:
		return "", err
	case resp := <-respChan:
		return strings.TrimPrefix(resp, string(r.response.Header())), nil
	}
}

func (r Rcon) listen(timeout time.Duration, respChan chan<- string, errChan chan<- error) {
	done := make(chan struct{})
	respBuf := make([]byte, respBufSiz)
	var sb strings.Builder

	for {
		select {
		case <-done:
			respChan <- sb.String()
			return
		default:
			rlen, err := r.conn.ReadUntil(time.Now().Add(timeout), respBuf)
			if err != nil {
				e, ok := err.(net.Error)
				if ok {
					if e.Timeout() {
						close(done)
					} else {
						errChan <- e
						return
					}
				}
			}

			if rlen > len(r.response.Header()) {
				if bytes.HasPrefix(respBuf, r.response.Header()) {
					sb.Write(respBuf[len(r.response.Header()):rlen])
				}
			}
		}
	}
}

func (r Rcon) Close() {
	r.conn.Close()
}
