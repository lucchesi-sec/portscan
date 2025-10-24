package services

var services = map[uint16]string{
	// Common ports (TCP and UDP)
	53:  "dns", // DNS
	445: "smb", // SMB

	// TCP ports
	21:    "ftp",
	22:    "ssh",
	23:    "telnet",
	25:    "smtp",
	80:    "http",
	110:   "pop3",
	143:   "imap",
	443:   "https",
	3306:  "mysql",
	3389:  "rdp",
	5432:  "postgresql",
	6379:  "redis",
	8080:  "http-alt",
	8443:  "https-alt",
	27017: "mongodb",

	// UDP-specific ports
	67:    "dhcp",        // BOOTP/DHCP Server
	68:    "dhcp",        // BOOTP/DHCP Client
	69:    "tftp",        // TFTP
	123:   "ntp",         // NTP
	137:   "netbios-ns",  // NetBIOS Name Service
	138:   "netbios-dgm", // NetBIOS Datagram Service
	139:   "netbios-ssn", // NetBIOS Session Service
	161:   "snmp",        // SNMP
	162:   "snmptrap",    // SNMP Trap
	500:   "isakmp",      // ISAKMP/IKEd (IPSec)
	514:   "syslog",      // Syslog
	520:   "rip",         // RIP
	1194:  "openvpn",     // OpenVPN
	1701:  "l2tp",        // L2TP
	1812:  "radius",      // RADIUS Authentication
	1813:  "radius-acct", // RADIUS Accounting
	1900:  "ssdp",        // UPnP SSDP
	3478:  "stun",        // STUN/TURN
	4500:  "ipsec-nat",   // IPSec NAT-Traversal
	5060:  "sip",         // SIP
	5061:  "sips",        // SIP over TLS
	5353:  "mdns",        // mDNS
	5355:  "llmnr",       // LLMNR
	10000: "webmin",      // Webmin or ndmp
	51820: "wireguard",   // WireGuard
}

// GetName returns a human-friendly service name for a well-known port.
// Falls back to "unknown" if the port is not in the map.
func GetName(port uint16) string {
	if name, ok := services[port]; ok {
		return name
	}
	return "unknown"
}
