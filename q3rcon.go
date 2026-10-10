package q3rcon

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"time"

	"charm.land/log/v2"
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

	log.Infof("Rcon outgoing address %s", conn.conn.RemoteAddr())

	r := &Rcon{
		conn:           conn,
		request:        newRequest(password),
		response:       newResponse(),
		loginTimeout:   5 * time.Second,
		defaultTimeout: 20 * time.Millisecond,
		timeouts:       make(map[string]time.Duration),
	}

	for _, o := range options {
		o(r)
	}

	if err := r.login(context.Background()); err != nil {
		return nil, fmt.Errorf("error logging in: %w", err)
	}

	return r, nil
}

func (r *Rcon) login(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, r.loginTimeout)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			resp, err := r.Exec(ctx, "login")
			if err != nil {
				return fmt.Errorf("login send: %w", err)
			}
			if resp == "" {
				continue
			}
			if strings.Contains(resp, "Bad rcon") {
				return errors.New("bad rcon password")
			}
			return nil
		}
	}
}

func (r Rcon) Exec(ctx context.Context, cmdWithArgs string) (string, error) {
	cmd, _, _ := strings.Cut(cmdWithArgs, " ")
	timeout := r.defaultTimeout
	if t, ok := r.timeouts[cmd]; ok {
		timeout = t
	}

	respChan := make(chan string, 1)
	errChan := make(chan error, 1)

	go r.listen(timeout, respChan, errChan)

	encodedCmd, err := r.request.encode(cmdWithArgs)
	if err != nil {
		return "", fmt.Errorf("error encoding command: %w", err)
	}

	if _, err = r.conn.Write(encodedCmd); err != nil {
		return "", fmt.Errorf("error writing command to connection: %w", err)
	}

	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case err := <-errChan:
		return "", err
	case resp := <-respChan:
		return resp, nil
	}
}

func (r Rcon) ExecBackground(cmd string) (string, error) {
	return r.Exec(context.Background(), cmd)
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

			if err := c.conn.SetReadDeadline(time.Now().Add(timeout)); err != nil {
				errChan <- fmt.Errorf("error setting read deadline: %w", err)
				return
			}

			rlen, err := r.conn.Read(respBuf)
			if err != nil {
				if e, ok := err.(net.Error); ok {
					if e.Timeout() {
						close(done)
						continue
					}
					errChan <- e
					return
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
