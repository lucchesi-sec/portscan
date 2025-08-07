package ui

import "testing"

func TestGetServiceName(t *testing.T) {
	tests := []struct {
		name string
		port uint16
		want string
	}{
		{
			name: "FTP port",
			port: 21,
			want: "FTP",
		},
		{
			name: "SSH port",
			port: 22,
			want: "SSH",
		},
		{
			name: "HTTP port",
			port: 80,
			want: "HTTP",
		},
		{
			name: "HTTPS port",
			port: 443,
			want: "HTTPS",
		},
		{
			name: "MySQL port",
			port: 3306,
			want: "MySQL",
		},
		{
			name: "PostgreSQL port",
			port: 5432,
			want: "PostgreSQL",
		},
		{
			name: "MongoDB port",
			port: 27017,
			want: "MongoDB",
		},
		{
			name: "unknown port",
			port: 9999,
			want: "Unknown",
		},
		{
			name: "port 1",
			port: 1,
			want: "Unknown",
		},
		{
			name: "high port",
			port: 65535,
			want: "Unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getServiceName(tt.port)
			if got != tt.want {
				t.Errorf("getServiceName(%d) = %q, want %q", tt.port, got, tt.want)
			}
		})
	}
}
