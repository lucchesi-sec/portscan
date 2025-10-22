package profiles

// Predefined scan profiles for common use cases
var profiles = map[string][]uint16{
	"quick": {
		// Top 100 most common ports
		21, 22, 23, 25, 53, 80, 110, 111, 135, 139, 143, 443, 445, 993, 995,
		1723, 3306, 3389, 5900, 8080, 8443, 8888,
		// Additional common ports
		20, 26, 37, 79, 81, 88, 106, 113, 119, 161, 162, 179, 194, 199,
		264, 280, 301, 306, 311, 340, 366, 389, 406, 407, 416, 417, 425,
		427, 443, 444, 458, 464, 465, 481, 497, 500, 512, 513, 514, 515,
		524, 541, 543, 544, 545, 548, 554, 555, 563, 587, 593, 616, 617,
		625, 631, 636, 646, 648, 666, 667, 668, 683, 687, 691, 700, 705,
		711, 714, 720, 722, 726, 749, 765, 777, 783, 787, 800, 801, 808,
		843, 873, 880, 888, 898, 900, 901, 902, 903, 911, 912, 981, 987,
		990, 992, 995, 999, 1000, 1001, 1002,
	},
	"web": {
		// Common web service ports
		80, 443, // HTTP/HTTPS
		8080, 8443, // Alternative HTTP/HTTPS
		3000, 3001, // Node.js common
		4200, 4443, // Angular dev
		5000, 5001, // Flask/ASP.NET
		7000, 7001, // Cassandra web
		8000, 8001, // Django/HTTP alt
		8081, 8082, // Additional HTTP
		8888, 8889, // Jupyter/misc
		9000, 9001, // PHP-FPM/misc
		9090, 9091, // Prometheus/misc
		10000, 10001, // Webmin
		// API and microservices
		3003, 3004, 3005, // Microservices
		4000, 4001, 4002, // API servers
		5555, 5556, // API gateways
		// Proxy and cache
		3128, 8123, // Squid proxy
		11211, // Memcached
		// WebSocket and streaming
		8081, 8082, 8083, // WebSocket
		1935, 8554, // RTMP/RTSP
	},
	"database": {
		// Relational databases
		3306,       // MySQL/MariaDB
		5432,       // PostgreSQL
		1433, 1434, // MSSQL
		1521, 1830, // Oracle
		50000, // DB2

		// NoSQL databases
		27017, 27018, 27019, // MongoDB
		6379, 6380, // Redis
		9042, 9160, // Cassandra
		5984, 6984, // CouchDB
		8086, 8088, // InfluxDB
		7000, 7001, // Cassandra inter-node

		// Search and analytics
		9200, 9300, // Elasticsearch
		8983, // Solr

		// Message queues
		5672, 15672, // RabbitMQ
		9092,  // Kafka
		11211, // Memcached
		2181,  // Zookeeper

		// Cache
		11211, 11212, // Memcached
		8091, 8092, // Couchbase
	},
	"full": {
		// This is handled specially - returns 1-65535
	},
	"udp-common": {
		// Common UDP services
		53,     // DNS
		67, 68, // DHCP
		69,            // TFTP
		123,           // NTP
		137, 138, 139, // NetBIOS
		161, 162, // SNMP
		445,        // SMB
		500,        // IPSec/IKE
		514,        // Syslog
		520,        // RIP
		1194,       // OpenVPN
		1701,       // L2TP
		1812, 1813, // RADIUS
		1900,       // UPnP
		3478,       // STUN/TURN
		4500,       // IPSec NAT-T
		5060, 5061, // SIP
		5353,  // mDNS
		5355,  // LLMNR
		10000, // Webmin
		51820, // WireGuard
	},
	"gateway": {
		// Common gateway/router services (TCP and UDP)
		22,     // SSH
		23,     // Telnet
		53,     // DNS
		67, 68, // DHCP
		80,    // HTTP
		161,   // SNMP
		443,   // HTTPS
		500,   // IPSec/IKE
		1194,  // OpenVPN
		1701,  // L2TP
		1723,  // PPTP
		4500,  // IPSec NAT-T
		5060,  // SIP
		8080,  // HTTP-alt
		8443,  // HTTPS-alt
		10000, // Webmin
		51820, // WireGuard
		// Router management interfaces
		8291, // MikroTik Winbox
		2000, // MikroTik bandwidth test
		8728, // MikroTik API
		8729, // MikroTik API-SSL
	},
	"voip": {
		// VoIP/SIP services
		5060, 5061, // SIP
		5004, 5005, // RTP
		3478, 3479, // STUN
		10000, 20000, // RTP range start/end markers
		4569, // IAX2
		2000, // Cisco SCCP
		1719, // H.323 Gatekeeper
		1720, // H.323 Call Signaling
		5038, // Asterisk Manager
	},
}

// GetProfile returns the ports for a given profile name
func GetProfile(name string) []uint16 {
	if name == "full" {
		// Generate 1-65535
		ports := make([]uint16, 65535)
		for i := uint16(0); i < 65535; i++ {
			ports[i] = i + 1
		}
		return ports
	}

	if ports, ok := profiles[name]; ok {
		return dedupePorts(ports)
	}

	return nil
}

// ListProfiles returns all available profile names
func ListProfiles() []string {
	names := make([]string, 0, len(profiles))
	for name := range profiles {
		names = append(names, name)
	}
	return names
}

func dedupePorts(ports []uint16) []uint16 {
	seen := make(map[uint16]struct{}, len(ports))
	result := make([]uint16, 0, len(ports))
	for _, port := range ports {
		if _, exists := seen[port]; exists {
			continue
		}
		seen[port] = struct{}{}
		result = append(result, port)
	}
	return result
}
