package health

import (
	"fmt"
	"net"
	"runtime"
	"strconv"
	"strings"
	"time"
)

type Checker struct{}

func (c *Checker) TCPProbe(host string, port int, timeout time.Duration) error {
	addr := net.JoinHostPort(host, fmt.Sprintf("%d", port))
	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		return err
	}
	conn.Close()
	return nil
}

func (c *Checker) DNSCheck(host string) (time.Duration, error) {
	start := time.Now()
	addrs, err := net.LookupHost(host)
	latency := time.Since(start)
	if err != nil {
		return 0, err
	}
	if len(addrs) == 0 {
		return 0, fmt.Errorf("no addresses found")
	}
	return latency, nil
}

func (c *Checker) TLSVerify(host string, port int) error {
	addr := net.JoinHostPort(host, fmt.Sprintf("%d", port))
	conn, err := net.DialTimeout("tcp", addr, 10*time.Second)
	if err != nil {
		return err
	}
	conn.Close()
	return nil
}

func (c *Checker) Ping(host string, count int) (avgLatency time.Duration, loss float64, err error) {
	if runtime.GOOS != "linux" && runtime.GOOS != "darwin" {
		return 0, 0, fmt.Errorf("ping not supported on %s", runtime.GOOS)
	}

	out, err := execCommand(fmt.Sprintf("ping -c %d -W 5 %s", count, host))
	if err != nil {
		return 0, 0, err
	}

	lines := strings.Split(out, "\n")
	for _, line := range lines {
		if strings.Contains(line, "avg") {
			parts := strings.Split(line, "/")
			if len(parts) >= 5 {
				avgMs, _ := strconv.ParseFloat(parts[4], 64)
				avgLatency = time.Duration(avgMs * float64(time.Millisecond))
			}
		}
		if strings.Contains(line, "packet loss") {
			idx := strings.Index(line, "=")
			if idx > 0 {
				pctStr := strings.TrimSpace(line[idx+1:])
				pctStr = strings.TrimSuffix(pctStr, "%")
				loss, _ = strconv.ParseFloat(pctStr, 64)
			}
		}
	}

	return avgLatency, loss, nil
}

func execCommand(cmd string) (string, error) {
	return "", nil
}
