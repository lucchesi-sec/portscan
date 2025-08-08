package core

import (
	"context"
	"math"
	"net"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

type ScanState string

const (
	StateOpen     ScanState = "open"
	StateClosed   ScanState = "closed"
	StateFiltered ScanState = "filtered"
)

type ResultEvent struct {
	Host     string
	Port     uint16
	State    ScanState
	Banner   string
	Duration time.Duration
	Protocol string // "tcp" or "udp"
}

type ProgressEvent struct {
	Total     int
	Completed int
	Rate      float64
}

type Scanner struct {
	config     *Config
	results    chan interface{}
	rateTicker *time.Ticker
	wg         sync.WaitGroup
	completed  atomic.Uint64
}

type Config struct {
	Workers        int
	Timeout        time.Duration
	UDPReadTimeout time.Duration // Specific timeout for UDP read operations
	UDPBufferSize  int           // Buffer size for UDP responses
	UDPJitterMaxMs int           // Maximum jitter in milliseconds for UDP scanning
	RateLimit      int
	BannerGrab     bool
	MaxRetries     int
	UDPWorkerRatio float64 // Ratio of workers to use for UDP scanning (0.5 = half of TCP workers)
}

func NewScanner(cfg *Config) *Scanner {
	if cfg.Workers <= 0 {
		cfg.Workers = 100
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = 200 * time.Millisecond
	}
	// Set default UDP read timeout if not specified
	if cfg.UDPReadTimeout <= 0 {
		// Default to same as TCP timeout for consistency
		cfg.UDPReadTimeout = cfg.Timeout
	}
	// Set default UDP buffer size if not specified
	if cfg.UDPBufferSize <= 0 {
		cfg.UDPBufferSize = 1024 // 1KB buffer for UDP responses
	}
	// Set default UDP jitter if not specified
	if cfg.UDPJitterMaxMs <= 0 {
		cfg.UDPJitterMaxMs = 10 // 10ms max jitter for UDP scanning
	}
	if cfg.RateLimit < 0 {
		cfg.RateLimit = 0
	}
	// Set default UDP worker ratio if not specified
	if cfg.UDPWorkerRatio <= 0 {
		cfg.UDPWorkerRatio = 0.5 // Default to half of TCP workers
	}

	var ticker *time.Ticker
	if cfg.RateLimit > 0 {
		interval := time.Second / time.Duration(cfg.RateLimit)
		ticker = time.NewTicker(interval)
	}

	return &Scanner{
		config:     cfg,
		results:    make(chan interface{}, 1000),
		rateTicker: ticker,
	}
}

func (s *Scanner) Results() <-chan interface{} { return s.results }

func (s *Scanner) ScanRange(ctx context.Context, host string, ports []uint16) {
	jobs := make(chan uint16, len(ports))

	// Start progress reporter
	progressDone := make(chan struct{})
	go func() {
		s.progressReporter(ctx, len(ports))
		close(progressDone)
	}()

	// Start workers
	for i := 0; i < s.config.Workers; i++ {
		s.wg.Add(1)
		go s.worker(ctx, host, jobs)
	}

	// Feed jobs
	go func() {
		for _, port := range ports {
			select {
			case <-ctx.Done():
				close(jobs)
				return
			case jobs <- port:
			}
		}
		close(jobs)
	}()

	// Wait for all workers to finish
	s.wg.Wait()

	// Wait for progress reporter to finish
	<-progressDone

	// Stop rate ticker if used
	if s.rateTicker != nil {
		s.rateTicker.Stop()
	}

	// Now safe to close results
	close(s.results)
}

func (s *Scanner) worker(ctx context.Context, host string, jobs <-chan uint16) {
	defer s.wg.Done()

	dialer := &net.Dialer{Timeout: s.config.Timeout}

	for {
		select {
		case <-ctx.Done():
			return
		case port, ok := <-jobs:
			if !ok {
				return
			}

			// Rate limit if enabled
			if s.rateTicker != nil {
				select {
				case <-ctx.Done():
					return
				case <-s.rateTicker.C:
				}
			}

			s.scanPort(ctx, dialer, host, port)
		}
	}
}

func (s *Scanner) scanPort(ctx context.Context, dialer *net.Dialer, host string, port uint16) {
	start := time.Now()
	address := net.JoinHostPort(host, strconv.Itoa(int(port)))

	conn, err := dialer.DialContext(ctx, "tcp", address)
	duration := time.Since(start)

	result := ResultEvent{Host: host, Port: port, Duration: duration, Protocol: "tcp"}

	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			result.State = StateFiltered
		} else {
			result.State = StateClosed
		}
	} else {
		result.State = StateOpen
		if s.config.BannerGrab {
			result.Banner = s.grabBanner(conn)
		}
		_ = conn.Close()
	}

	select {
	case s.results <- result:
		// Count completed results when delivered
		s.completed.Add(1)
	case <-ctx.Done():
	}
}

func (s *Scanner) grabBanner(conn net.Conn) string {
	_ = conn.SetReadDeadline(time.Now().Add(time.Second))
	buffer := make([]byte, 512)
	n, err := conn.Read(buffer)
	if err != nil || n == 0 {
		return ""
	}
	return string(buffer[:n])
}

func (s *Scanner) progressReporter(ctx context.Context, total int) {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	startTime := time.Now()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			completedUint := s.completed.Load()
			// Safely convert to int with proper bounds checking
			var completed int
			if completedUint > math.MaxInt32 {
				completed = math.MaxInt32
			} else {
				completed = int(completedUint) // #nosec G115 - safe after bounds check
			}
			if completed > total {
				completed = total
			}
			elapsed := time.Since(startTime).Seconds()
			if elapsed <= 0 {
				elapsed = 0.001
			}
			rate := float64(completed) / elapsed

			progress := ProgressEvent{Total: total, Completed: completed, Rate: rate}
			select {
			case s.results <- progress:
			case <-ctx.Done():
				return
			}

			if completed >= total {
				return
			}
		}
	}
}
