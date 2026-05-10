package tailscale

import (
	"fmt"
	"net"
)

var tailnet4 = net.IPNet{IP: net.IPv4(100, 64, 0, 0), Mask: net.CIDRMask(10, 32)}

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
