package profiles

import (
	"testing"
)

func TestGetProfile(t *testing.T) {
	tests := []struct {
		name string
	}{
		{"quick"},
		{"web"},
		{"database"},
		{"full"},
		{"udp-common"},
		{"gateway"},
		{"voip"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ports := GetProfile(tt.name)
			var expected int
			if tt.name == "full" {
				expected = 65535
			} else {
				expected = len(dedupePorts(profiles[tt.name]))
			}

			if len(ports) != expected {
				t.Fatalf("GetProfile(%s) = %d ports, want %d", tt.name, len(ports), expected)
			}

			seen := make(map[uint16]struct{}, len(ports))
			for _, p := range ports {
				if _, exists := seen[p]; exists {
					t.Fatalf("duplicate port %d found in profile %s", p, tt.name)
				}
				seen[p] = struct{}{}
			}

			if tt.name == "full" {
				if ports[0] != 1 {
					t.Errorf("First port should be 1, got %d", ports[0])
				}
				if ports[len(ports)-1] != 65535 {
					t.Errorf("Last port should be 65535, got %d", ports[len(ports)-1])
				}
			}
		})
	}
}

func TestListProfiles(t *testing.T) {
	profiles := ListProfiles()
	if len(profiles) != 7 {
		t.Errorf("Expected 7 profiles, got %d", len(profiles))
	}

	expected := map[string]bool{
		"quick":      true,
		"web":        true,
		"database":   true,
		"full":       true,
		"udp-common": true,
		"gateway":    true,
		"voip":       true,
	}

	for _, profile := range profiles {
		if !expected[profile] {
			t.Errorf("Unexpected profile: %s", profile)
		}
	}
}
