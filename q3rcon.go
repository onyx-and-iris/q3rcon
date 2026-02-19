package q3rcon

import (
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"time"

	"github.com/charmbracelet/log"
)

const respBufSiz = 2048

type encoder interface {
	encode(cmd string) ([]byte, error)
}

type decoder interface {
	isValid(buf []byte) bool
	decode(buf []byte) string
}

type Rcon struct {
	conn     io.ReadWriteCloser
	request  encoder
	response decoder

	loginTimeout   time.Duration
	defaultTimeout time.Duration
	timeouts       map[string]time.Duration
}

func New(host string, port int, password string, options ...Option) (*Rcon, error) {
	if password == "" {
		return nil, errors.New("no password provided")
	}

	conn, err := newUDPConn(host, port)
	if err != nil {
		return nil, fmt.Errorf("error creating UDP connection: %w", err)
	}

	r := &Rcon{
		conn:     conn,
		request:  newRequest(password),
		response: newResponse(),

		loginTimeout:   5 * time.Second,
		defaultTimeout: 20 * time.Millisecond,
		timeouts:       make(map[string]time.Duration),
	}

	for _, o := range options {
		o(r)
	}

	if err = r.login(); err != nil {
		return nil, fmt.Errorf("error logging in: %w", err)
	}

	return r, nil
}

func (r Rcon) login() error {
	timeout := time.After(r.loginTimeout)
	for {
		select {
		case <-timeout:
			return errors.New("timeout logging in")
		default:
			resp, err := r.Send("login")
			if err != nil {
				return fmt.Errorf("error sending login command: %w", err)
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

func (r Rcon) Send(cmdWithArgs string) (string, error) {
	cmd, _, _ := strings.Cut(string(cmdWithArgs), " ")
	timeout, ok := r.timeouts[cmd]
	if !ok {
		timeout = r.defaultTimeout
	} else {
		log.Debugf("%s in timeouts map, using timeout %v", cmd, timeout)
	}

	respChan := make(chan string)
	errChan := make(chan error)

	go r.listen(timeout, respChan, errChan)

	encodedCmd, err := r.request.encode(cmdWithArgs)
	if err != nil {
		return "", fmt.Errorf("error encoding command: %w", err)
	}

	_, err = r.conn.Write(encodedCmd)
	if err != nil {
		return "", fmt.Errorf("error writing command to connection: %w", err)
	}

	select {
	case err := <-errChan:
		return "", err
	case resp := <-respChan:
		return resp, nil
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
			c, ok := r.conn.(*UDPConn)
			if !ok {
				errChan <- errors.New("connection is not a UDPConn")
				return
			}
			err := c.conn.SetReadDeadline(time.Now().Add(timeout))
			if err != nil {
				errChan <- fmt.Errorf("error setting read deadline: %w", err)
				return
			}
			rlen, err := r.conn.Read(respBuf)
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

			if r.response.isValid(respBuf[:rlen]) {
				sb.WriteString(r.response.decode(respBuf[:rlen]))
			}
		}
	}
}

func (r Rcon) Close() error {
	return r.conn.Close()
}
