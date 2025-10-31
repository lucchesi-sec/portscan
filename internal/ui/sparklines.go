package ui

import (
	"math"
	"strings"
	"time"
)

// TimeSeriesData represents a time-series data point
type TimeSeriesData struct {
	Timestamp time.Time
	Value     float64
}

// SparklineData holds time-series data for different metrics
type SparklineData struct {
	ScanRate    []TimeSeriesData
	DiscoveryRate []TimeSeriesData
	ErrorRate   []TimeSeriesData
	MaxPoints   int
}

// NewSparklineData creates a new sparkline data collector
func NewSparklineData() *SparklineData {
	return &SparklineData{
		ScanRate:      make([]TimeSeriesData, 0),
		DiscoveryRate: make([]TimeSeriesData, 0),
		ErrorRate:     make([]TimeSeriesData, 0),
		MaxPoints:     60, // Keep last 60 seconds of data
	}
}

// AddScanRate adds a scan rate data point
func (s *SparklineData) AddScanRate(rate float64) {
	s.addDataPoint(&s.ScanRate, rate)
}

// AddDiscoveryRate adds a discovery rate data point
func (s *SparklineData) AddDiscoveryRate(rate float64) {
	s.addDataPoint(&s.DiscoveryRate, rate)
}

// AddErrorRate adds an error rate data point
func (s *SparklineData) AddErrorRate(rate float64) {
	s.addDataPoint(&s.ErrorRate, rate)
}

// addDataPoint adds a data point to the series, maintaining max size
func (s *SparklineData) addDataPoint(series *[]TimeSeriesData, value float64) {
	now := time.Now()
	point := TimeSeriesData{
		Timestamp: now,
		Value:     value,
	}

	*series = append(*series, point)

	// Remove old data points to maintain size
	if len(*series) > s.MaxPoints {
		*series = (*series)[len(*series)-s.MaxPoints:]
	}
}

// RenderSparkline renders an ASCII sparkline chart
func (s *SparklineData) RenderSparkline(series []TimeSeriesData, width int) string {
	if len(series) == 0 {
		return strings.Repeat(" ", width)
	}

	// Extract just the values
	values := make([]float64, len(series))
	for i, data := range series {
		values[i] = data.Value
	}

	return renderSparklineValues(values, width)
}

// renderSparklineValues renders a sparkline from float values
func renderSparklineValues(values []float64, width int) string {
	if len(values) == 0 || width <= 0 {
		return ""
	}

	// Find min and max values
	min := values[0]
	max := values[0]
	for _, v := range values {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}

	// Handle case where all values are the same
	if math.Abs(max-min) < 0.0001 {
		max = min + 1
	}

	// Sparkline characters from lowest to highest
	blocks := []rune{' ', '▂', '▃', '▄', '▅', '▆', '▇', '█'}

	// Determine the sampling interval
	dataLen := len(values)
	interval := 1
	if dataLen > width {
		interval = dataLen / width
	}

	var result strings.Builder

	for i := 0; i < width; i++ {
		// Sample the appropriate data point
		dataIndex := i * interval
		if dataIndex >= dataLen {
			dataIndex = dataLen - 1
		}

		value := values[dataIndex]
		
		// Normalize value to 0-1 range
		normalized := (value - min) / (max - min)
		if normalized < 0 {
			normalized = 0
		} else if normalized > 1 {
			normalized = 1
		}

		// Convert to block character
		blockIndex := int(normalized * float64(len(blocks)-1))
		result.WriteRune(blocks[blockIndex])
	}

	return result.String()
}

// GetMetricSummary returns a summary of recent metrics
func (s *SparklineData) GetMetricSummary() MetricSummary {
	summary := MetricSummary{}

	if len(s.ScanRate) > 0 {
		summary.CurrentScanRate = s.ScanRate[len(s.ScanRate)-1].Value
		summary.AverageScanRate = calculateAverage(s.ScanRate)
		summary.PeakScanRate = calculateMax(s.ScanRate)
	}

	if len(s.DiscoveryRate) > 0 {
		summary.CurrentDiscoveryRate = s.DiscoveryRate[len(s.DiscoveryRate)-1].Value
		summary.AverageDiscoveryRate = calculateAverage(s.DiscoveryRate)
	}

	if len(s.ErrorRate) > 0 {
		summary.CurrentErrorRate = s.ErrorRate[len(s.ErrorRate)-1].Value
		summary.AverageErrorRate = calculateAverage(s.ErrorRate)
	}

	return summary
}

// MetricSummary provides aggregated metrics
type MetricSummary struct {
	CurrentScanRate      float64
	AverageScanRate      float64
	PeakScanRate         float64
	CurrentDiscoveryRate float64
	AverageDiscoveryRate float64
	CurrentErrorRate    float64
	AverageErrorRate     float64
}

// Helper functions for calculations
func calculateAverage(series []TimeSeriesData) float64 {
	if len(series) == 0 {
		return 0
	}

	sum := 0.0
	for _, data := range series {
		sum += data.Value
	}

	return sum / float64(len(series))
}

func calculateMax(series []TimeSeriesData) float64 {
	if len(series) == 0 {
		return 0
	}

	max := series[0].Value
	for _, data := range series {
		if data.Value > max {
			max = data.Value
		}
	}

	return max
}
