package profiles

import (
	"testing"
)

func TestGetProfile(t *testing.T) {
	tests := []struct {
		name     string
		expected int
	}{
		{"quick", 121},
		{"web", 40},
		{"database", 32},
		{"full", 65535},
		{"udp-common", 27},
		{"gateway", 22},
		{"voip", 13},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ports := GetProfile(tt.name)
			if len(ports) != tt.expected {
				t.Errorf("GetProfile(%s) = %d ports, want %d", tt.name, len(ports), tt.expected)
			}
			if tt.name == "full" {
				// Check first and last ports for full profile
				if ports[0] != 1 {
					t.Errorf("First port should be 1, got %d", ports[0])
				}
				if ports[65534] != 65535 {
					t.Errorf("Last port should be 65535, got %d", ports[65534])
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
