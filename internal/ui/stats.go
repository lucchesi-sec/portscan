package ui

import (
	"sort"
	"time"

	"github.com/lucchesi-sec/portscan/internal/core"
)

// StatsData holds computed statistics for the dashboard
type StatsData struct {
	TotalResults  int
	OpenCount     int
	ClosedCount   int
	FilteredCount int

	// Service statistics
	ServiceCounts map[string]int
	TopServices   []ServiceStat

	// Response time statistics
	MinResponseTime time.Duration
	MaxResponseTime time.Duration
	AvgResponseTime time.Duration
	P95ResponseTime time.Duration

	// Performance metrics
	CurrentRate float64
	AverageRate float64

	// Network statistics
	UniqueHosts   int
	HostsWithOpen int
}

// ServiceStat represents a service with its count
type ServiceStat struct {
	Name  string
	Count int
}

// computeStats calculates statistics from current results
func (m *ScanUI) computeStats() *StatsData {
	stats := &StatsData{
		ServiceCounts: make(map[string]int),
	}

	results := m.results.Items()
	if len(results) == 0 {
		return stats
	}

	stats.TotalResults = len(results)

	var totalDuration time.Duration
	minDuration := time.Hour
	maxDuration := time.Duration(0)
	var durations []time.Duration

	hostsMap := make(map[string]bool)
	hostsWithOpen := make(map[string]bool)

	// Collect statistics
	for _, result := range results {
		// Count states
		switch result.State {
		case core.StateOpen:
			stats.OpenCount++
			hostsWithOpen[result.Host] = true
		case core.StateClosed:
			stats.ClosedCount++
		case core.StateFiltered:
			stats.FilteredCount++
		}

		// Count services
		service := getServiceName(result.Port)
		if service != "" && service != "Unknown" {
			stats.ServiceCounts[service]++
		}

		// Response times
		if result.Duration > 0 {
			totalDuration += result.Duration
			durations = append(durations, result.Duration)

			if result.Duration < minDuration {
				minDuration = result.Duration
			}
			if result.Duration > maxDuration {
				maxDuration = result.Duration
			}
		}

		// Unique hosts
		hostsMap[result.Host] = true
	}

	// Response time stats
	if len(durations) > 0 {
		stats.MinResponseTime = minDuration
		stats.MaxResponseTime = maxDuration
		stats.AvgResponseTime = totalDuration / time.Duration(len(durations))

		// Calculate P95
		sort.Slice(durations, func(i, j int) bool {
			return durations[i] < durations[j]
		})
		p95Index := int(float64(len(durations)) * 0.95)
		if p95Index < len(durations) {
			stats.P95ResponseTime = durations[p95Index]
		}
	}

	// Top services
	type servicePair struct {
		name  string
		count int
	}
	var servicePairs []servicePair
	for name, count := range stats.ServiceCounts {
		servicePairs = append(servicePairs, servicePair{name, count})
	}
	sort.Slice(servicePairs, func(i, j int) bool {
		return servicePairs[i].count > servicePairs[j].count
	})

	// Take top 5
	maxServices := 5
	if len(servicePairs) < maxServices {
		maxServices = len(servicePairs)
	}
	for i := 0; i < maxServices; i++ {
		stats.TopServices = append(stats.TopServices, ServiceStat{
			Name:  servicePairs[i].name,
			Count: servicePairs[i].count,
		})
	}

	// Network stats
	stats.UniqueHosts = len(hostsMap)
	stats.HostsWithOpen = len(hostsWithOpen)

	// Performance
	stats.CurrentRate = m.currentRate
	stats.AverageRate = m.progressTrack.AverageRate

	return stats
}

// getPercentage calculates percentage
func getPercentage(part, total int) float64 {
	if total == 0 {
		return 0
	}
	return float64(part) / float64(total) * 100
}
