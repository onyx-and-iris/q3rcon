package conn

import (
	"bytes"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/onyx-and-iris/q3rcon/internal/packet"
	log "github.com/sirupsen/logrus"
)

type UDPConn struct {
	conn     *net.UDPConn
	response packet.Response
}

func New(host string, port int) (UDPConn, error) {
	udpAddr, err := net.ResolveUDPAddr("udp4", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		return UDPConn{}, err
	}
	conn, err := net.DialUDP("udp4", nil, udpAddr)
	if err != nil {
		return UDPConn{}, err
	}
	log.Infof("Outgoing address %s", conn.RemoteAddr())

	return UDPConn{
		conn:     conn,
		response: packet.NewResponse(),
	}, nil
}

func (c UDPConn) Write(buf []byte) (int, error) {
	n, err := c.conn.Write(buf)
	if err != nil {
		return 0, err
	}

	return n, nil
}

func (c UDPConn) Listen(timeout time.Duration, resp chan<- string, errChan chan<- error) {
	c.conn.SetReadDeadline(time.Now().Add(timeout))
	done := make(chan struct{})
	var sb strings.Builder
	buf := make([]byte, 2048)

	for {
		select {
		case <-done:
			resp <- sb.String()
			return
		default:
			rlen, _, err := c.conn.ReadFromUDP(buf)
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
			if rlen < len(c.response.Header()) {
				continue
			}

			if bytes.HasPrefix(buf, c.response.Header()) {
				sb.Write(buf[len(c.response.Header()):rlen])
			}
		}
	}
}

func (c UDPConn) Close() error {
	err := c.conn.Close()
	if err != nil {
		return err
	}
	return nil
}
