package core

import "time"

// Scanner configuration defaults
const (
	// DefaultWorkerCount is the default number of concurrent workers
	DefaultWorkerCount = 100

	// DefaultTimeoutMs is the default connection timeout in milliseconds
	DefaultTimeoutMs = 200

	// DefaultUDPBufferSize is the buffer size for UDP responses (1KB)
	DefaultUDPBufferSize = 1024

	// DefaultUDPJitterMaxMs is the maximum jitter in milliseconds for UDP scanning
	DefaultUDPJitterMaxMs = 10

	// DefaultUDPWorkerRatio is the default ratio of workers for UDP (half of TCP workers)
	DefaultUDPWorkerRatio = 0.5

	// DefaultMaxRetries is the default number of retry attempts for failed connections
	DefaultMaxRetries = 2
)

// Channel buffer sizes
const (
	// ResultChannelBufferSize is the buffer size for the results channel
	ResultChannelBufferSize = 1000
)

// Banner grabbing configuration
const (
	// BannerGrabTimeout is the timeout for reading service banners
	BannerGrabTimeout = 1 * time.Second

	// BannerBufferSize is the buffer size for reading service banners
	BannerBufferSize = 512
)

// Progress reporting configuration
const (
	// ProgressReportInterval is how often to report progress updates
	ProgressReportInterval = 100 * time.Millisecond
)

// Retry backoff configuration
const (
	// RetryBackoffBase is the base duration for retry backoff
	RetryBackoffBase = 50 * time.Millisecond

	// RetryJitterMinMs is the minimum jitter in milliseconds
	RetryJitterMinMs = 10

	// RetryJitterMaxMs is the maximum jitter in milliseconds
	RetryJitterMaxMs = 50

	// RetryJitterRangeMs is the range for jitter calculation (max - min + 1)
	RetryJitterRangeMs = 41
)

// Rate limiting configuration
const (
	// MaxSafeRateLimit is the maximum safe rate limit in packets per second
	MaxSafeRateLimit = 15000
)
