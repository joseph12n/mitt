// Package pair builds LAN pairing values for the dashboard QR.
package pair

import (
	"errors"
	"net"
	"net/url"
	"strings"
)

// errNoLAN is returned when no suitable LAN IPv4 address exists.
var errNoLAN = errors.New("no LAN IPv4 address available")

// LanIPv4 returns the first suitable LAN IPv4 address from addrs.
// It skips loopback, link-local, unspecified and non-IPv4 entries and
// requires a global-unicast address. Interface down state is not visible
// from net.Addr, so callers pass only addresses from up interfaces when
// that filtering matters.
func LanIPv4(addrs []net.Addr) (string, error) {
	for _, a := range addrs {
		var ip net.IP
		switch v := a.(type) {
		case *net.IPNet:
			ip = v.IP
		case *net.IPAddr:
			ip = v.IP
		default:
			continue
		}
		ip4 := ip.To4()
		if ip4 == nil {
			continue
		}
		if ip4.IsLoopback() || ip4.IsLinkLocalUnicast() || ip4.IsUnspecified() {
			continue
		}
		if !ip4.IsGlobalUnicast() {
			continue
		}
		return ip4.String(), nil
	}
	return "", errNoLAN
}

// BaseURL resolves the public base URL shown in the QR.
// A non-empty advertise value wins verbatim (trimmed). Otherwise the empty
// or wildcard host in addr (":8080" or "0.0.0.0:8080") is replaced with ip.
func BaseURL(advertise, addr, ip string) string {
	if a := strings.TrimSpace(advertise); a != "" {
		return a
	}
	addr = strings.TrimSpace(addr)
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		// Addr without host:port (e.g. "8080" or "") falls back to ip:port.
		port = strings.TrimPrefix(addr, ":")
		if port == "" {
			port = "8080"
		}
		return "http://" + net.JoinHostPort(ip, port)
	}
	if host == "" || host == "0.0.0.0" || host == "::" {
		host = ip
	}
	if port == "" {
		port = "8080"
	}
	return "http://" + net.JoinHostPort(host, port)
}

// Code builds the mitt:// pairing payload embedding the base URL and token.
func Code(baseURL, token string) string {
	return "mitt://pair?url=" + url.QueryEscape(baseURL) + "&token=" + url.QueryEscape(token)
}
