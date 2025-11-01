package ui

import (
	"github.com/charmbracelet/bubbles/table"
)

type columnSpec struct {
	title  string
	weight int
	min    int
}

var defaultColumnSpecs = []columnSpec{
	{"Host", ColumnWeightHost, ColumnMinWidthHost},
	{"Port", ColumnWeightPort, ColumnMinWidthPort},
	{"Protocol", ColumnWeightProtocol, ColumnMinWidthProtocol},
	{"State", ColumnWeightState, ColumnMinWidthState},
	{"Service", ColumnWeightService, ColumnMinWidthService},
	{"Banner", ColumnWeightBanner, ColumnMinWidthBanner},
	{"Latency", ColumnWeightLatency, ColumnMinWidthLatency},
}

var columnPriorityOrder = []int{0, 5, 4, 6, 1, 2, 3}

const tableHorizontalFrame = 4

func (m *ScanUI) applyTableGeometry() {
	if m == nil {
		return
	}
	if m.width <= 0 || m.height <= 0 {
		return
	}

	contentWidth := m.tableViewportWidth()
	columns := calculateColumnWidths(contentWidth)
	m.table.SetColumns(columns)
	m.table.SetWidth(contentWidth)

	overhead := tableOverheadLines(m.scanning, m.indicatorsVisible())
	availableRows := max(MinTableHeight, m.height-overhead)
	m.table.SetHeight(availableRows)
}

// calculateColumnWidths returns table columns sized according to the configured
// weights while respecting minimum widths. The total width will never be less
// than the sum of minimum widths, ensuring the table stays legible on narrow
// terminals.
func calculateColumnWidths(totalWidth int) []table.Column {
	if totalWidth <= 0 {
		totalWidth = sumMinWidths()
	}

	minWidth := sumMinWidths()
	if totalWidth < minWidth {
		totalWidth = minWidth
	}

	remaining := totalWidth - minWidth
	weightSum := sumWeights()
	columnWidths := make([]int, len(defaultColumnSpecs))
	extraAssigned := 0

	for i, spec := range defaultColumnSpecs {
		columnWidths[i] = spec.min
		if remaining <= 0 || weightSum == 0 {
			continue
		}

		extra := (remaining * spec.weight) / weightSum
		columnWidths[i] += extra
		extraAssigned += extra
	}

	leftover := remaining - extraAssigned
	for _, idx := range columnPriorityOrder {
		if leftover <= 0 {
			break
		}
		columnWidths[idx]++
		leftover--
	}

	columns := make([]table.Column, len(defaultColumnSpecs))
	for i, spec := range defaultColumnSpecs {
		columns[i] = table.Column{Title: spec.title, Width: columnWidths[i]}
	}

	return columns
}

func sumMinWidths() int {
	total := 0
	for _, spec := range defaultColumnSpecs {
		total += spec.min
	}
	return total
}

func sumWeights() int {
	total := 0
	for _, spec := range defaultColumnSpecs {
		total += spec.weight
	}
	return total
}

// tableOverheadLines returns the number of non-table lines that occupy the
// viewport. The indicators flag should reflect whether sort or filter badges
// will be shown.
func tableOverheadLines(scanning bool, showIndicators bool) int {
	overhead := HeightBreadcrumb + HeightHeader + HeightStatus + HeightSpacing + HeightFooter
	if scanning {
		overhead += HeightProgress
	}
	if showIndicators {
		overhead += HeightIndicators
	}
	return overhead
}
