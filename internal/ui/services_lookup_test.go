package ui

import "testing"

func TestGetServiceName(t *testing.T) {
	tests := []struct {
		name     string
		port     uint16
		expected string
	}{
		{"FTP", 21, "FTP"},
		{"SSH", 22, "SSH"},
		{"Telnet", 23, "Telnet"},
		{"SMTP", 25, "SMTP"},
		{"DNS", 53, "DNS"},
		{"HTTP", 80, "HTTP"},
		{"POP3", 110, "POP3"},
		{"IMAP", 143, "IMAP"},
		{"HTTPS", 443, "HTTPS"},
		{"SMB", 445, "SMB"},
		{"MySQL", 3306, "MySQL"},
		{"RDP", 3389, "RDP"},
		{"PostgreSQL", 5432, "PostgreSQL"},
		{"Redis", 6379, "Redis"},
		{"HTTP-Alt", 8080, "HTTP-Alt"},
		{"HTTPS-Alt", 8443, "HTTPS-Alt"},
		{"MongoDB", 27017, "MongoDB"},
		{"unknown port", 12345, "Unknown"},
		{"another unknown", 9999, "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getServiceName(tt.port)
			if result != tt.expected {
				t.Errorf("getServiceName(%d) = %s; want %s", tt.port, result, tt.expected)
			}
		})
	}
}

func TestGetServiceName_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		port     uint16
		expected string
	}{
		{"port 0", 0, "Unknown"},
		{"port 1", 1, "Unknown"},
		{"max port", 65535, "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getServiceName(tt.port)
			if result != tt.expected {
				t.Errorf("getServiceName(%d) = %s; want %s", tt.port, result, tt.expected)
			}
		})
	}
}
