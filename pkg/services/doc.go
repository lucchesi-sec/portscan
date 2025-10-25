// Package services provides service name lookup for well-known ports.
//
// This package maintains a comprehensive mapping of port numbers to service names
// for both TCP and UDP protocols. It's used by the scanner to display human-readable
// service names in the UI and exports instead of just raw port numbers.
//
// Example usage:
//
//	// Lookup TCP service
//	service := services.LookupTCP(80)
//	fmt.Println(service) // Output: "http"
//
//	// Lookup UDP service
//	service := services.LookupUDP(53)
//	fmt.Println(service) // Output: "dns"
//
//	// Unknown ports return empty string
//	service := services.LookupTCP(54321)
//	fmt.Println(service) // Output: ""
//
// Service Database:
//
// The package includes mappings for:
//   - Common TCP services (HTTP, HTTPS, SSH, FTP, SMTP, etc.)
//   - Common UDP services (DNS, DHCP, NTP, SNMP, etc.)
//   - Database ports (MySQL, PostgreSQL, MongoDB, Redis, etc.)
//   - Application servers (Tomcat, WebLogic, JBoss, etc.)
//   - Message queues (RabbitMQ, Kafka, ActiveMQ, etc.)
//   - VPN protocols (OpenVPN, WireGuard, IPSec, etc.)
//   - VoIP services (SIP, RTP, IAX2, etc.)
//
// The service names follow IANA port assignment conventions where applicable,
// with additional entries for popular non-standard services.
//
// Protocol-Specific Lookups:
//
// Different protocols may assign the same port to different services:
//   - TCP 53: dns (DNS over TCP)
//   - UDP 53: dns (DNS queries)
//   - TCP 161: snmp (SNMP over TCP)
//   - UDP 161: snmp (SNMP traps)
//
// Performance:
//
// Lookups use Go maps for O(1) average-case performance. The service database
// is loaded once at package initialization.
package services
