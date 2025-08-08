package services

// GetName returns a human-friendly service name for a well-known TCP port.
// Falls back to "unknown" if the port is not in the map.
func GetName(port uint16) string {
	services := map[uint16]string{
		21:    "ftp",
		22:    "ssh",
		23:    "telnet",
		25:    "smtp",
		53:    "dns",
		80:    "http",
		110:   "pop3",
		143:   "imap",
		443:   "https",
		445:   "smb",
		3306:  "mysql",
		3389:  "rdp",
		5432:  "postgresql",
		6379:  "redis",
		8080:  "http-alt",
		8443:  "https-alt",
		27017: "mongodb",
	}
	if name, ok := services[port]; ok {
		return name
	}
	return "unknown"
}
