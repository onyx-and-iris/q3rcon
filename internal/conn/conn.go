package conn

import (
	"fmt"
	"net"
	"time"

	log "github.com/sirupsen/logrus"
)

type UDPConn struct {
	conn *net.UDPConn
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
		conn: conn,
	}, nil
}

func (c UDPConn) Write(buf []byte) (int, error) {
	n, err := c.conn.Write(buf)
	if err != nil {
		return 0, err
	}

	return n, nil
}

func (c UDPConn) ReadUntil(timeout time.Time, buf []byte) (int, error) {
	c.conn.SetReadDeadline(timeout)
	rlen, _, err := c.conn.ReadFromUDP(buf)
	if err != nil {
		return 0, err
	}
	return rlen, nil
}

func (c UDPConn) Close() error {
	err := c.conn.Close()
	if err != nil {
		return err
	}
	return nil
}
