package netutil

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"time"
)

func TCPProbe(host string, port int, timeout time.Duration) error {
	addr := net.JoinHostPort(host, fmt.Sprintf("%d", port))
	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		return err
	}
	conn.Close()
	return nil
}

func TLSProbe(host string, port int, timeout time.Duration) error {
	addr := net.JoinHostPort(host, fmt.Sprintf("%d", port))
	dialer := &net.Dialer{Timeout: timeout}
	conn, err := tls.DialWithDialer(dialer, "tcp", addr, &tls.Config{
		InsecureSkipVerify: true,
	})
	if err != nil {
		return err
	}
	conn.Close()
	return nil
}

func DNSLookup(host string) ([]string, error) {
	return net.LookupHost(host)
}

func Ping(ctx context.Context, host string, count int) (time.Duration, error) {
	start := time.Now()
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(host, "80"), 5*time.Second)
	if err != nil {
		return 0, err
	}
	conn.Close()
	return time.Since(start), nil
}
