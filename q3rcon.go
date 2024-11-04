package q3rcon

import (
	"errors"
	"strings"
	"time"

	"github.com/onyx-and-iris/q3rcon/internal/conn"
	"github.com/onyx-and-iris/q3rcon/internal/packet"
)

// Option is a functional option type that allows us to configure the VbanTxt.
type Option func(*Rcon)

func WithDefaultTimeout(timeout time.Duration) Option {
	return func(rcon *Rcon) {
		rcon.defaultTimeout = timeout
	}
}

// WithTimeouts is a functional option to set the timeouts for responses
func WithTimeouts(timeouts map[string]time.Duration) Option {
	return func(rcon *Rcon) {
		rcon.timeouts = timeouts
	}
}

type Rcon struct {
	conn     conn.UDPConn
	request  packet.Request
	response packet.Response

	defaultTimeout time.Duration
	timeouts       map[string]time.Duration

	resp chan string
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
		conn:           conn,
		request:        packet.NewRequest(password),
		resp:           make(chan string),
		defaultTimeout: 20 * time.Millisecond,
		timeouts:       make(map[string]time.Duration),
	}

	for _, o := range options {
		o(r)
	}

	err = r.Login()
	if err != nil {
		return nil, err
	}

	return r, nil
}

func (r Rcon) Login() error {
	timeout := time.After(2 * time.Second)
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

	e := make(chan error)
	go r.conn.Listen(timeout, r.resp, e)
	_, err := r.conn.Write(r.request.Encode(cmd))
	if err != nil {
		return "", err
	}

	select {
	case err := <-e:
		return "", err
	case resp := <-r.resp:
		return strings.TrimPrefix(resp, string(r.response.Header())), nil
	}
}

func (r Rcon) Close() {
	r.conn.Close()
}
