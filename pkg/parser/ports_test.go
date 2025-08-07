package parser

import (
	"reflect"
	"testing"
)

func TestParsePorts(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    []uint16
		wantErr bool
	}{
		{
			name:  "single port",
			input: "80",
			want:  []uint16{80},
		},
		{
			name:  "multiple ports",
			input: "80,443,8080",
			want:  []uint16{80, 443, 8080},
		},
		{
			name:  "port range",
			input: "80-82",
			want:  []uint16{80, 81, 82},
		},
		{
			name:  "mixed ports and ranges",
			input: "22,80-82,443",
			want:  []uint16{22, 80, 81, 82, 443},
		},
		{
			name:  "with spaces",
			input: " 22 , 80 - 82 , 443 ",
			want:  []uint16{22, 80, 81, 82, 443},
		},
		{
			name:  "duplicate ports removed",
			input: "80,80,443",
			want:  []uint16{80, 443},
		},
		{
			name:  "overlapping ranges",
			input: "80-82,81-83",
			want:  []uint16{80, 81, 82, 83},
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
		{
			name:    "only spaces",
			input:   "   ",
			wantErr: true,
		},
		{
			name:    "invalid port - too low",
			input:   "0",
			wantErr: true,
		},
		{
			name:    "invalid port - too high",
			input:   "65536",
			wantErr: true,
		},
		{
			name:    "invalid port - not a number",
			input:   "abc",
			wantErr: true,
		},
		{
			name:    "invalid range - start > end",
			input:   "82-80",
			wantErr: true,
		},
		{
			name:    "invalid range - multiple dashes",
			input:   "80-81-82",
			wantErr: true,
		},
		{
			name:    "invalid range - no start",
			input:   "-80",
			wantErr: true,
		},
		{
			name:    "invalid range - no end",
			input:   "80-",
			wantErr: true,
		},
		{
			name:  "common ports",
			input: "21,22,23,25,53,80,110,143,443,445",
			want:  []uint16{21, 22, 23, 25, 53, 80, 110, 143, 443, 445},
		},
		{
			name:  "high port range",
			input: "65533-65535",
			want:  []uint16{65533, 65534, 65535},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParsePorts(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParsePorts() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParsePorts() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParsePortsPerformance(t *testing.T) {
	// Test parsing a large range
	ports, err := ParsePorts("1-1000")
	if err != nil {
		t.Fatalf("ParsePorts() error = %v", err)
	}
	if len(ports) != 1000 {
		t.Errorf("Expected 1000 ports, got %d", len(ports))
	}
}
