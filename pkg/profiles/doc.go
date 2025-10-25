// Package profiles provides predefined port scanning profiles for common use cases.
//
// Profiles are curated sets of commonly-used ports optimized for specific
// scanning scenarios. They eliminate the need to manually specify port lists
// and encode security best practices.
//
// Available Profiles:
//
// TCP Profiles:
//   - quick: Top 100 most common ports (fast reconnaissance)
//   - web: HTTP, HTTPS, and web application ports
//   - database: Database servers (MySQL, PostgreSQL, MongoDB, Redis, etc.)
//   - full: All 65,535 ports (comprehensive but slow)
//
// UDP Profiles:
//   - udp-common: Common UDP services (DNS, DHCP, NTP, SNMP, etc.)
//   - voip: VoIP/SIP services (SIP, RTP, IAX2, etc.)
//
// Mixed Protocol Profiles:
//   - gateway: Router/gateway services (TCP and UDP)
//
// Example usage:
//
//	// Get ports for a profile
//	ports := profiles.GetProfile("web")
//	fmt.Printf("Scanning %d web ports\n", len(ports))
//
//	// List all available profiles
//	for _, name := range profiles.ListProfiles() {
//	    fmt.Println(name)
//	}
//
// Profile Details:
//
// quick Profile (100 ports):
//   Top 100 most commonly open ports across all services. Ideal for fast
//   reconnaissance when you need quick results. Includes: SSH (22), HTTP (80),
//   HTTPS (443), SMB (445), RDP (3389), MySQL (3306), etc.
//
// web Profile (30+ ports):
//   HTTP/HTTPS and web application ports including: 80, 443, 8080, 8443,
//   3000-3005 (Node.js), 4200 (Angular), 5000 (Flask), 8000 (Django),
//   9090 (Prometheus), WebSocket, RTMP streaming, and API gateways.
//
// database Profile (25+ ports):
//   Relational databases (MySQL, PostgreSQL, MSSQL, Oracle), NoSQL
//   (MongoDB, Redis, Cassandra, CouchDB), search engines (Elasticsearch),
//   and message queues (RabbitMQ, Kafka).
//
// udp-common Profile (20+ ports):
//   Essential UDP services: DNS (53), DHCP (67/68), NTP (123), SNMP (161/162),
//   VPN protocols (OpenVPN 1194, WireGuard 51820, IPSec 500/4500),
//   NetBIOS, mDNS, and LLMNR.
//
// gateway Profile (20+ ports):
//   Services commonly found on routers and gateways: SSH, Telnet, HTTP/HTTPS,
//   DNS, DHCP, SNMP, VPN endpoints, SIP, and router management interfaces
//   (MikroTik, Webmin).
//
// voip Profile (15+ ports):
//   VoIP and telephony services: SIP (5060/5061), RTP media ports,
//   STUN/TURN, IAX2, Cisco SCCP, H.323, and Asterisk Manager.
//
// full Profile (65,535 ports):
//   Every possible TCP/UDP port. Use with caution as this takes significant
//   time even at high scan rates (e.g., ~9 seconds at 7,500 pps).
//
// Port Deduplication:
//
// All profiles automatically deduplicate ports, so overlapping port
// definitions are safe and won't cause repeated scans.
package profiles
