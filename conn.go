package q3rcon

import (
	"fmt"
	"net"

	"github.com/charmbracelet/log"
)

type UDPConn struct {
	conn *net.UDPConn
}

func newUDPConn(host string, port int) (*UDPConn, error) {
	udpAddr, err := net.ResolveUDPAddr("udp4", net.JoinHostPort(host, fmt.Sprintf("%d", port)))
	if err != nil {
		return nil, err
	}
	conn, err := net.DialUDP("udp4", nil, udpAddr)
	if err != nil {
		return nil, err
	}
	log.Infof("Outgoing address %s", conn.RemoteAddr())

	return &UDPConn{
		conn: conn,
	}, nil
}

func (c *UDPConn) Write(buf []byte) (int, error) {
	n, err := c.conn.Write(buf)
	if err != nil {
		return 0, err
	}

	return n, nil
}

func (c *UDPConn) Read(buf []byte) (int, error) {
	rlen, _, err := c.conn.ReadFromUDP(buf)
	if err != nil {
		return 0, err
	}
	return rlen, nil
}

func (c *UDPConn) Close() error {
	err := c.conn.Close()
	if err != nil {
		return err
	}
	return nil
}
