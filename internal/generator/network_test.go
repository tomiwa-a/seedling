package generator

import (
	"context"
	"math/rand"
	"net"
	"strings"
	"testing"
)

func TestIPv4Generator(t *testing.T) {
	g := &IPv4Generator{}
	ctx := context.Background()
	row := testRow{}

	for i := 0; i < 100; i++ {
		v, err := g.Generate(ctx, row)
		if err != nil {
			t.Fatal(err)
		}
		ip := v.(string)
		parsed := net.ParseIP(ip)
		if parsed == nil {
			t.Errorf("invalid IPv4: %q", ip)
		}
		if parsed.To4() == nil {
			t.Errorf("not an IPv4 address: %q", ip)
		}
	}
}

func TestIPv6Generator(t *testing.T) {
	g := &IPv6Generator{}
	ctx := context.Background()
	row := testRow{}

	for i := 0; i < 50; i++ {
		v, err := g.Generate(ctx, row)
		if err != nil {
			t.Fatal(err)
		}
		ip := v.(string)
		if !strings.Contains(ip, ":") {
			t.Errorf("expected colon-separated IPv6, got %q", ip)
		}
		parts := strings.Split(ip, ":")
		if len(parts) != 8 {
			t.Errorf("expected 8 groups in IPv6, got %d: %q", len(parts), ip)
		}
	}
}

func TestMACGenerator(t *testing.T) {
	g := &MACGenerator{}
	ctx := context.Background()
	row := testRow{}

	for i := 0; i < 50; i++ {
		v, err := g.Generate(ctx, row)
		if err != nil {
			t.Fatal(err)
		}
		mac := v.(string)
		parts := strings.Split(mac, ":")
		if len(parts) != 6 {
			t.Errorf("expected 6 groups in MAC, got %d: %q", len(parts), mac)
		}
	}
}

func TestUserAgentGenerator(t *testing.T) {
	g := &UserAgentGenerator{rnd: rand.New(rand.NewSource(42))}
	ctx := context.Background()
	row := testRow{}

	seen := make(map[string]bool)
	for i := 0; i < 50; i++ {
		v, err := g.Generate(ctx, row)
		if err != nil {
			t.Fatal(err)
		}
		ua := v.(string)
		if len(ua) < 10 {
			t.Errorf("user agent too short: %q", ua)
		}
		seen[ua] = true
	}
	if len(seen) < 3 {
		t.Errorf("expected at least 3 different user agents, got %d", len(seen))
	}
}
