package session

import (
	"fmt"

	netbridge "github.com/netbridge/netbridge"
)

func BuildEnvVars(session *netbridge.Session) map[string]string {
	if session == nil {
		return nil
	}

	vars := make(map[string]string)

	switch session.Mode {
	case netbridge.ModeSOCKS:
		vars["http_proxy"] = fmt.Sprintf("socks5://%s", session.LocalAddr)
		vars["https_proxy"] = fmt.Sprintf("socks5://%s", session.LocalAddr)
		vars["all_proxy"] = fmt.Sprintf("socks5://%s", session.LocalAddr)
	case netbridge.ModeHTTP:
		vars["http_proxy"] = fmt.Sprintf("http://%s", session.LocalAddr)
		vars["https_proxy"] = fmt.Sprintf("http://%s", session.LocalAddr)
	}

	vars["no_proxy"] = "localhost,127.0.0.1,::1"
	return vars
}
