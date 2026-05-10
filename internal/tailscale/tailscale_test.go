package tailscale

import (
	"net"
	"testing"
)

func mustCIDR(s string) net.Addr {
	ip, n, err := net.ParseCIDR(s)
	if err != nil {
		panic(err)
	}
	n.IP = ip
	return n
}

func TestCandidateIPExactCIDRAndInterfaceFlags(t *testing.T) {
	ifs := []Interface{
		{Name: "down", Flags: 0, Addrs: []net.Addr{mustCIDR("100.64.0.2/32")}},
		{Name: "lo", Flags: net.FlagUp | net.FlagLoopback, Addrs: []net.Addr{mustCIDR("100.64.0.3/32")}},
		{Name: "bad", Flags: net.FlagUp, Addrs: []net.Addr{mustCIDR("100.128.0.1/32")}},
		{Name: "ts0", Flags: net.FlagUp, Addrs: []net.Addr{mustCIDR("100.127.255.254/32")}},
	}
	got, err := CandidateIP(ifs)
	if err != nil || got != "100.127.255.254" {
		t.Fatalf("got %q err %v", got, err)
	}
}

func TestCandidateIPRejectsOutsideRange(t *testing.T) {
	_, err := CandidateIP([]Interface{{Name: "bad", Flags: net.FlagUp, Addrs: []net.Addr{mustCIDR("100.63.255.255/32"), mustCIDR("127.0.0.1/32")}}})
	if err == nil {
		t.Fatal("expected no Tailscale IP")
	}
}
