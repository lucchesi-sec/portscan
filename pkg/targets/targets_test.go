package targets

import "testing"

func TestResolveHosts(t *testing.T) {
	inputs := []string{"example.com", "192.168.1.1", "example.com"}
	targets, err := Resolve(inputs, Options{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(targets) != 2 {
		t.Fatalf("expected 2 targets, got %d", len(targets))
	}

	if targets[0] != "example.com" || targets[1] != "192.168.1.1" {
		t.Errorf("unexpected targets: %#v", targets)
	}
}

func TestResolveCIDR(t *testing.T) {
	inputs := []string{"192.168.1.0/30"}
	targets, err := Resolve(inputs, Options{CIDRHostLimit: 16})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []string{"192.168.1.0", "192.168.1.1", "192.168.1.2", "192.168.1.3"}
	if len(targets) != len(expected) {
		t.Fatalf("expected %d hosts, got %d", len(expected), len(targets))
	}

	for i, host := range expected {
		if targets[i] != host {
			t.Errorf("expected %s at index %d, got %s", host, i, targets[i])
		}
	}
}

func TestResolveCIDRTooLarge(t *testing.T) {
	inputs := []string{"10.0.0.0/16"}
	_, err := Resolve(inputs, Options{CIDRHostLimit: 128})
	if err == nil {
		t.Fatalf("expected error for oversized CIDR")
	}
}

func TestResolveInvalidHostname(t *testing.T) {
	inputs := []string{"-badhost"}
	_, err := Resolve(inputs, Options{})
	if err == nil {
		t.Fatalf("expected validation error")
	}
}
