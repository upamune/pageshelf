package tailscale

import (
	"encoding/json"
	"fmt"
	"net"
	"os/exec"
	"strings"
)

var tailnet4 = net.IPNet{IP: net.IPv4(100, 64, 0, 0), Mask: net.CIDRMask(10, 32)}

var commandOutput = func(name string, args ...string) ([]byte, error) {
	return exec.Command(name, args...).Output()
}

type Interface struct {
	Name  string
	Flags net.Flags
	Addrs []net.Addr
}

func CandidateIP(ifs []Interface) (string, error) {
	for _, ifi := range ifs {
		if ifi.Flags&net.FlagUp == 0 || ifi.Flags&net.FlagLoopback != 0 {
			continue
		}
		for _, a := range ifi.Addrs {
			var ip net.IP
			switch v := a.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip4 := ip.To4(); ip4 != nil && tailnet4.Contains(ip4) {
				return ip4.String(), nil
			}
		}
	}
	return "", fmt.Errorf("no Tailscale 100.64.0.0/10 address found")
}

func IP() (string, error) {
	nifs, err := net.Interfaces()
	if err != nil {
		return "", err
	}
	ifs := make([]Interface, 0, len(nifs))
	for _, ifi := range nifs {
		addrs, err := ifi.Addrs()
		if err != nil {
			continue
		}
		ifs = append(ifs, Interface{Name: ifi.Name, Flags: ifi.Flags, Addrs: addrs})
	}
	return CandidateIP(ifs)
}

func validMagicDNSName(s string) (string, bool) {
	s = strings.TrimSuffix(strings.TrimSpace(s), ".")
	if s == "" || net.ParseIP(s) != nil || !strings.HasSuffix(strings.ToLower(s), ".ts.net") {
		return "", false
	}
	return s, true
}

// MagicDNSName returns the local Tailscale MagicDNS hostname when it can be
// discovered locally. It avoids depending on Tailscale APIs: first try the
// system FQDN, then the local tailscale CLI's self DNSName.
func MagicDNSName() (string, error) {
	if out, err := commandOutput("hostname", "-f"); err == nil {
		if name, ok := validMagicDNSName(string(out)); ok {
			return name, nil
		}
	}

	if out, err := commandOutput("tailscale", "status", "--json"); err == nil {
		var status struct {
			Self struct {
				DNSName string
			} `json:"Self"`
		}
		if json.Unmarshal(out, &status) == nil {
			if name, ok := validMagicDNSName(status.Self.DNSName); ok {
				return name, nil
			}
		}
	}

	return "", fmt.Errorf("no Tailscale MagicDNS .ts.net hostname found")
}
