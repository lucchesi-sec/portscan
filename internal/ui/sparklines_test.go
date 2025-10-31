package ui

import (
	"math"
	"strings"
	"testing"
	"time"
)

func TestNewSparklineData(t *testing.T) {
	s := NewSparklineData()

	if s == nil {
		t.Fatal("NewSparklineData returned nil")
	}

	if s.MaxPoints != 60 {
		t.Errorf("MaxPoints = %d; want 60", s.MaxPoints)
	}

	if len(s.ScanRate) != 0 {
		t.Error("ScanRate should be empty initially")
	}
}

func TestSparklineData_AddDataPoint(t *testing.T) {
	s := NewSparklineData()

	// Add single point
	s.AddScanRate(100.0)
	if len(s.ScanRate) != 1 {
		t.Errorf("len(ScanRate) = %d; want 1", len(s.ScanRate))
	}

	if s.ScanRate[0].Value != 100.0 {
		t.Errorf("ScanRate[0].Value = %f; want 100.0", s.ScanRate[0].Value)
	}

	// Add multiple points
	for i := 0; i < 10; i++ {
		s.AddDiscoveryRate(float64(i))
	}

	if len(s.DiscoveryRate) != 10 {
		t.Errorf("len(DiscoveryRate) = %d; want 10", len(s.DiscoveryRate))
	}
}

func TestSparklineData_MaxPointsLimit(t *testing.T) {
	s := NewSparklineData()
	s.MaxPoints = 10

	// Add more than MaxPoints
	for i := 0; i < 20; i++ {
		s.AddScanRate(float64(i))
	}

	if len(s.ScanRate) != 10 {
		t.Errorf("len(ScanRate) = %d; want 10 (MaxPoints)", len(s.ScanRate))
	}

	// Verify oldest points were removed (should have 10-19)
	if s.ScanRate[0].Value != 10.0 {
		t.Errorf("ScanRate[0].Value = %f; want 10.0", s.ScanRate[0].Value)
	}

	if s.ScanRate[9].Value != 19.0 {
		t.Errorf("ScanRate[9].Value = %f; want 19.0", s.ScanRate[9].Value)
	}
}

func TestRenderSparkline_EmptySeries(t *testing.T) {
	s := NewSparklineData()
	result := s.RenderSparkline([]TimeSeriesData{}, 10)

	expected := strings.Repeat(" ", 10)
	if result != expected {
		t.Errorf("RenderSparkline(empty) = %q; want %q", result, expected)
	}
}

func TestRenderSparkline_SingleValue(t *testing.T) {
	s := NewSparklineData()
	data := []TimeSeriesData{
		{Timestamp: time.Now(), Value: 50.0},
	}

	result := s.RenderSparkline(data, 5)

	// Single value should render as middle block
	if len(result) != 5 {
		t.Errorf("len(result) = %d; want 5", len(result))
	}
}

func TestRenderSparkline_AllSameValues(t *testing.T) {
	s := NewSparklineData()
	data := make([]TimeSeriesData, 10)
	for i := range data {
		data[i] = TimeSeriesData{Timestamp: time.Now(), Value: 100.0}
	}

	result := s.RenderSparkline(data, 10)

	// All same values should render as all spaces (handled by adding 1 to max)
	if len(result) != 10 {
		t.Errorf("len(result) = %d; want 10", len(result))
	}
}

func TestRenderSparkline_IncreasingValues(t *testing.T) {
	s := NewSparklineData()
	data := make([]TimeSeriesData, 8)
	for i := range data {
		data[i] = TimeSeriesData{Timestamp: time.Now(), Value: float64(i)}
	}

	result := s.RenderSparkline(data, 8)
	runes := []rune(result)

	if len(runes) != 8 {
		t.Errorf("rune count = %d; want 8", len(runes))
	}

	// First character should be lowest block
	if runes[0] == '█' {
		t.Error("First character should not be full block")
	}

	// Last character should be highest block
	if runes[7] != '█' {
		t.Errorf("Last character = %c; want █", runes[7])
	}
}

func TestRenderSparkline_Sampling(t *testing.T) {
	s := NewSparklineData()

	// Create 100 data points
	data := make([]TimeSeriesData, 100)
	for i := range data {
		data[i] = TimeSeriesData{Timestamp: time.Now(), Value: float64(i)}
	}

	// Render to width of 10 (should sample every 10 points)
	result := s.RenderSparkline(data, 10)

	runeCount := len([]rune(result))
	if runeCount != 10 {
		t.Errorf("rune count = %d; want 10", runeCount)
	}
}

func TestRenderSparklineValues_ZeroWidth(t *testing.T) {
	result := renderSparklineValues([]float64{1, 2, 3}, 0)

	if result != "" {
		t.Errorf("renderSparklineValues(width=0) = %q; want empty string", result)
	}
}

func TestRenderSparklineValues_EmptyValues(t *testing.T) {
	result := renderSparklineValues([]float64{}, 10)

	if result != "" {
		t.Errorf("renderSparklineValues(empty) = %q; want empty string", result)
	}
}

func TestRenderSparklineValues_Normalization(t *testing.T) {
	// Test values from 0 to 100
	values := []float64{0, 25, 50, 75, 100}
	result := renderSparklineValues(values, 5)
	runes := []rune(result)

	if len(runes) != 5 {
		t.Errorf("rune count = %d; want 5", len(runes))
	}

	// First should be lowest
	if runes[0] == '█' {
		t.Error("First rune should not be full block")
	}

	// Last should be highest
	if runes[4] != '█' {
		t.Errorf("Last rune = %c; want █", runes[4])
	}
}

func TestGetMetricSummary_Empty(t *testing.T) {
	s := NewSparklineData()
	summary := s.GetMetricSummary()

	if summary.CurrentScanRate != 0 {
		t.Errorf("CurrentScanRate = %f; want 0", summary.CurrentScanRate)
	}

	if summary.AverageScanRate != 0 {
		t.Errorf("AverageScanRate = %f; want 0", summary.AverageScanRate)
	}

	if summary.PeakScanRate != 0 {
		t.Errorf("PeakScanRate = %f; want 0", summary.PeakScanRate)
	}
}

func TestGetMetricSummary_WithData(t *testing.T) {
	s := NewSparklineData()

	// Add scan rate data: 10, 20, 30, 40, 50
	rates := []float64{10, 20, 30, 40, 50}
	for _, rate := range rates {
		s.AddScanRate(rate)
	}

	summary := s.GetMetricSummary()

	if summary.CurrentScanRate != 50 {
		t.Errorf("CurrentScanRate = %f; want 50", summary.CurrentScanRate)
	}

	expectedAvg := 30.0
	if math.Abs(summary.AverageScanRate-expectedAvg) > 0.01 {
		t.Errorf("AverageScanRate = %f; want %f", summary.AverageScanRate, expectedAvg)
	}

	if summary.PeakScanRate != 50 {
		t.Errorf("PeakScanRate = %f; want 50", summary.PeakScanRate)
	}
}

func TestGetMetricSummary_AllMetrics(t *testing.T) {
	s := NewSparklineData()

	s.AddScanRate(100)
	s.AddScanRate(200)

	s.AddDiscoveryRate(5)
	s.AddDiscoveryRate(10)

	s.AddErrorRate(0.1)
	s.AddErrorRate(0.2)

	summary := s.GetMetricSummary()

	if summary.CurrentScanRate != 200 {
		t.Errorf("CurrentScanRate = %f; want 200", summary.CurrentScanRate)
	}

	if summary.CurrentDiscoveryRate != 10 {
		t.Errorf("CurrentDiscoveryRate = %f; want 10", summary.CurrentDiscoveryRate)
	}

	if summary.CurrentErrorRate != 0.2 {
		t.Errorf("CurrentErrorRate = %f; want 0.2", summary.CurrentErrorRate)
	}

	if math.Abs(summary.AverageScanRate-150) > 0.01 {
		t.Errorf("AverageScanRate = %f; want 150", summary.AverageScanRate)
	}

	if math.Abs(summary.AverageDiscoveryRate-7.5) > 0.01 {
		t.Errorf("AverageDiscoveryRate = %f; want 7.5", summary.AverageDiscoveryRate)
	}

	if math.Abs(summary.AverageErrorRate-0.15) > 0.01 {
		t.Errorf("AverageErrorRate = %f; want 0.15", summary.AverageErrorRate)
	}
}

func TestCalculateAverage_Empty(t *testing.T) {
	result := calculateAverage([]TimeSeriesData{})

	if result != 0 {
		t.Errorf("calculateAverage(empty) = %f; want 0", result)
	}
}

func TestCalculateAverage_Values(t *testing.T) {
	data := []TimeSeriesData{
		{Value: 10},
		{Value: 20},
		{Value: 30},
		{Value: 40},
	}

	result := calculateAverage(data)
	expected := 25.0

	if math.Abs(result-expected) > 0.01 {
		t.Errorf("calculateAverage = %f; want %f", result, expected)
	}
}

func TestCalculateMax_Empty(t *testing.T) {
	result := calculateMax([]TimeSeriesData{})

	if result != 0 {
		t.Errorf("calculateMax(empty) = %f; want 0", result)
	}
}

func TestCalculateMax_Values(t *testing.T) {
	data := []TimeSeriesData{
		{Value: 10},
		{Value: 50},
		{Value: 30},
		{Value: 20},
	}

	result := calculateMax(data)

	if result != 50 {
		t.Errorf("calculateMax = %f; want 50", result)
	}
}

func TestCalculateMax_SingleValue(t *testing.T) {
	data := []TimeSeriesData{{Value: 42}}

	result := calculateMax(data)

	if result != 42 {
		t.Errorf("calculateMax = %f; want 42", result)
	}
}

func TestSparklineData_TimestampOrdering(t *testing.T) {
	s := NewSparklineData()

	now := time.Now()
	s.AddScanRate(100)
	time.Sleep(1 * time.Millisecond)
	s.AddScanRate(200)

	if len(s.ScanRate) != 2 {
		t.Fatalf("len(ScanRate) = %d; want 2", len(s.ScanRate))
	}

	// First timestamp should be before second
	if !s.ScanRate[0].Timestamp.Before(s.ScanRate[1].Timestamp) {
		t.Error("Timestamps should be ordered")
	}

	// Both should be after now
	if s.ScanRate[0].Timestamp.Before(now) {
		t.Error("First timestamp should be after test start")
	}
}

func TestRenderSparkline_NegativeValues(t *testing.T) {
	s := NewSparklineData()
	data := []TimeSeriesData{
		{Value: -10},
		{Value: -5},
		{Value: 0},
		{Value: 5},
		{Value: 10},
	}

	result := s.RenderSparkline(data, 5)
	runes := []rune(result)

	if len(runes) != 5 {
		t.Errorf("rune count = %d; want 5", len(runes))
	}

	// Should handle negative values correctly
	if runes[0] == runes[4] {
		t.Error("First and last characters should differ for increasing values")
	}
}

func TestRenderSparkline_VeryLargeWidth(t *testing.T) {
	s := NewSparklineData()
	data := []TimeSeriesData{
		{Value: 1},
		{Value: 2},
		{Value: 3},
	}

	// Request width larger than data
	result := s.RenderSparkline(data, 100)

	runeCount := len([]rune(result))
	if runeCount != 100 {
		t.Errorf("rune count = %d; want 100", runeCount)
	}
}
