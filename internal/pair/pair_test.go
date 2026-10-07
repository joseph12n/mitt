package pair

import (
	"net"
	"testing"
)

// mustAddr parses a CIDR into a net.Addr for table tests.
func mustAddr(t *testing.T, cidr string) net.Addr {
	t.Helper()
	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		t.Fatalf("ParseCIDR(%q) = %v", cidr, err)
	}
	return &net.IPNet{IP: ip, Mask: ipnet.Mask}
}

func TestLanIPv4(t *testing.T) {
	loopback := mustAddr(t, "127.0.0.1/8")
	linkLocal := mustAddr(t, "169.254.10.20/16")
	lanA := mustAddr(t, "192.168.1.20/24")
	lanB := mustAddr(t, "10.0.0.5/8")
	ipv6 := mustAddr(t, "fe80::1/10")

	cases := []struct {
		name    string
		addrs   []net.Addr
		want    string
		wantErr bool
	}{
		{"loopback only errors", []net.Addr{loopback}, "", true},
		{"empty errors", nil, "", true},
		{"link local skipped", []net.Addr{linkLocal, lanA}, "192.168.1.20", false},
		{"loopback then LAN picks first suitable", []net.Addr{loopback, lanA, lanB}, "192.168.1.20", false},
		{"ipv6 skipped", []net.Addr{ipv6, lanB}, "10.0.0.5", false},
		{"link local only errors", []net.Addr{linkLocal}, "", true},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			got, err := LanIPv4(tt.addrs)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("LanIPv4() = %q, want error", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("LanIPv4() error = %v", err)
			}
			if got != tt.want {
				t.Fatalf("LanIPv4() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestBaseURL(t *testing.T) {
	cases := []struct {
		name      string
		advertise string
		addr      string
		ip        string
		want      string
	}{
		{"advertise override wins", "http://192.168.1.20:8080", ":8080", "10.0.0.9", "http://192.168.1.20:8080"},
		{"advertise trimmed", "  http://example:9090  ", ":8080", "10.0.0.9", "http://example:9090"},
		{"empty host gets ip", "", ":8080", "192.168.1.20", "http://192.168.1.20:8080"},
		{"wildcard host gets ip", "", "0.0.0.0:8080", "192.168.1.20", "http://192.168.1.20:8080"},
		{"concrete host kept", "", "192.168.5.5:9090", "10.0.0.9", "http://192.168.5.5:9090"},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			if got := BaseURL(tt.advertise, tt.addr, tt.ip); got != tt.want {
				t.Fatalf("BaseURL() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestCode(t *testing.T) {
	got := Code("http://192.168.1.20:8080", "abc 123&=?")
	want := "mitt://pair?url=http%3A%2F%2F192.168.1.20%3A8080&token=abc+123%26%3D%3F"
	if got != want {
		t.Fatalf("Code() = %q, want %q", got, want)
	}
}
