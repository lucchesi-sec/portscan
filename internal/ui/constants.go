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

	// ColumnWidthHost is the width of the host column
	ColumnWidthHost = 20

	// ColumnWidthPort is the width of the port column
	ColumnWidthPort = 8

	// ColumnWidthProtocol is the width of the protocol column
	ColumnWidthProtocol = 8

	// ColumnWidthState is the width of the state column
	ColumnWidthState = 10

	// ColumnWidthService is the width of the service column
	ColumnWidthService = 15

	// ColumnWidthBanner is the width of the banner column
	ColumnWidthBanner = 35

	// ColumnWidthLatency is the width of the latency column
	ColumnWidthLatency = 10
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
