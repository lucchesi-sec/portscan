package ui

import "time"

// UI buffer sizes
const (
	// DefaultResultBufferSize is the default size of the result buffer
	DefaultResultBufferSize = 10000
)

// Table dimensions
const (
	// TableDefaultHeight is the default height of the results table
	TableDefaultHeight = 15

	// MinTableHeight prevents the table from collapsing on very short viewports
	MinTableHeight = 5
)

// Table column weights in percentage points. They must sum to 100.
const (
	ColumnWeightHost     = 24
	ColumnWeightPort     = 8
	ColumnWeightProtocol = 8
	ColumnWeightState    = 10
	ColumnWeightService  = 18
	ColumnWeightBanner   = 24
	ColumnWeightLatency  = 8
)

// Table column minimum widths to keep data legible on narrow terminals.
const (
	ColumnMinWidthHost     = 16
	ColumnMinWidthPort     = 6
	ColumnMinWidthProtocol = 6
	ColumnMinWidthState    = 8
	ColumnMinWidthService  = 12
	ColumnMinWidthBanner   = 12
	ColumnMinWidthLatency  = 6
)

// Banner truncation
const (
	// BannerMaxDisplayLength is the maximum length to display for banners
	BannerMaxDisplayLength = 40

	// BannerTruncateLength is the length at which to truncate banners
	BannerTruncateLength = 37
)

// Navigation and scrolling
const (
	// PageScrollLines is the number of lines to scroll per page up/down
	PageScrollLines = 10
)

// Event polling
const (
	// ResultPollTimeout is the timeout for polling result events
	ResultPollTimeout = 100 * time.Millisecond
)

// Dashboard and UI layout
const (
	// DashboardMinWidth is the minimum width required to show the dashboard
	DashboardMinWidth = 120

	// StatusBarLabelWidth is the width of status bar labels
	StatusBarLabelWidth = 10
)

// Dashboard panel ratios and spacing.
const (
	DashboardLeftWidthPercent  = 0.65
	DashboardRightWidthPercent = 0.35
	DashboardGutterWidth       = 3
)

// Modal dialog dimensions and positioning
const (
	// ModalWidthPercent is the modal width as a percentage of screen width
	ModalWidthPercent = 0.6

	// ModalHeightPercent is the modal height as a percentage of screen height
	ModalHeightPercent = 0.4

	// ModalBorderPadding is the padding inside modal borders
	ModalBorderPadding = 2

	// ModalMinWidth is the minimum modal width in characters
	ModalMinWidth = 40

	// ModalMinHeight is the minimum modal height in characters
	ModalMinHeight = 10
)

// Fixed-height contributions used to compute the table viewport height.
const (
	HeightBreadcrumb = 1
	HeightHeader     = 1
	HeightProgress   = 1
	HeightStatus     = 2
	HeightIndicators = 1
	HeightSpacing    = 1
	HeightFooter     = 2
)
