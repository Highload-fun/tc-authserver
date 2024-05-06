package geo

import (
	"net"
	"testing"
)

func TestTree_Find(t *testing.T) {
	tree := NewTree()

	cidrs := []string{
		"192.168.0.0/24",
		"192.168.0.0/16",
		"10.0.0.0/8",
		"172.1.2.3/32",
	}

	for _, cidr := range cidrs {
		_, ipNet, err := net.ParseCIDR(cidr)
		if err != nil {
			t.Fatal(err)
		}
		tree.Insert(ipNet, &City{Country: cidr})
	}

	for _, tc := range []struct {
		ip     string
		exists bool
		cidr   string
	}{
		{"192.168.0.10", true, "192.168.0.0/24"},
		{"192.168.1.10", true, "192.168.0.0/16"},
		{"192.169.1.10", false, ""},
		{"172.1.2.3", true, "172.1.2.3/32"},
	} {
		city := tree.Find(net.ParseIP(tc.ip))
		if city == nil && tc.exists {
			t.Fatalf("cidr for ip %s wasn't found, expected %s", tc.ip, tc.cidr)
		}

		if city != nil && !tc.exists {
			t.Fatalf("cidr for ip %s was found (%s), expected nil", tc.ip, tc.cidr)
		}

		if city != nil && city.Country != tc.cidr {
			t.Fatalf("cidr for ip %s is %s, expected %s", tc.ip, city.Country, tc.cidr)
		}
	}
}
