package core

import (
	"context"
	"math/rand"
	"net"
	"strconv"
	"sync"
	"time"
)

type Scanner struct {
	config           *Config
	results          chan Event
	rateTicker       *time.Ticker
	wg               sync.WaitGroup
	progressReporter *ProgressReporter
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
		cfg.Workers = DefaultWorkerCount
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = DefaultTimeoutMs * time.Millisecond
	}
	// Set default UDP read timeout if not specified
	if cfg.UDPReadTimeout <= 0 {
		// Default to same as TCP timeout for consistency
		cfg.UDPReadTimeout = cfg.Timeout
	}
	// Set default UDP buffer size if not specified
	if cfg.UDPBufferSize <= 0 {
		cfg.UDPBufferSize = DefaultUDPBufferSize
	}
	// Set default UDP jitter if not specified
	if cfg.UDPJitterMaxMs <= 0 {
		cfg.UDPJitterMaxMs = DefaultUDPJitterMaxMs
	}
	if cfg.RateLimit < 0 {
		cfg.RateLimit = 0
	}
	// Set default UDP worker ratio if not specified
	if cfg.UDPWorkerRatio <= 0 {
		cfg.UDPWorkerRatio = DefaultUDPWorkerRatio
	}

	var ticker *time.Ticker
	if cfg.RateLimit > 0 {
		interval := time.Second / time.Duration(cfg.RateLimit)
		ticker = time.NewTicker(interval)
	}

	resultsChan := make(chan Event, ResultChannelBufferSize)
	return &Scanner{
		config:           cfg,
		results:          resultsChan,
		rateTicker:       ticker,
		progressReporter: NewProgressReporter(resultsChan),
	}
}

func (s *Scanner) jobBufferSize(total int) int {
	if total <= 0 {
		return 0
	}
	buffer := total
	if maxBuffer := s.config.Workers * 4; maxBuffer > 0 && buffer > maxBuffer {
		buffer = maxBuffer
	}
	if buffer < 1 {
		buffer = 1
	}
	return buffer
}

func (s *Scanner) Results() <-chan Event { return s.results }

func (s *Scanner) ScanRange(ctx context.Context, host string, ports []uint16) {
	s.ScanTargets(ctx, []ScanTarget{{Host: host, Ports: ports}})
}

func (s *Scanner) ScanTargets(ctx context.Context, targets []ScanTarget) {
	totalPorts := totalPortCount(targets)
	if totalPorts == 0 {
		if s.rateTicker != nil {
			s.rateTicker.Stop()
		}
		close(s.results)
		return
	}

	s.progressReporter.SetCompleted(0)

	jobs := make(chan scanJob, s.jobBufferSize(totalPorts))
	progressDone := s.progressReporter.StartReporting(ctx, totalPorts)

	s.startWorkers(ctx, jobs)

	go s.feedJobs(ctx, jobs, targets)

	s.wg.Wait()

	s.finishScan(ctx, progressDone)
}

func (s *Scanner) startWorkers(ctx context.Context, jobs <-chan scanJob) {
	for i := 0; i < s.config.Workers; i++ {
		s.wg.Add(1)
		go s.worker(ctx, jobs)
	}
}

func (s *Scanner) feedJobs(ctx context.Context, jobs chan<- scanJob, targets []ScanTarget) {
	defer close(jobs)
	for _, target := range targets {
		host := target.Host
		for _, port := range target.Ports {
			select {
			case <-ctx.Done():
				return
			case jobs <- scanJob{host: host, port: port}:
			}
		}
	}
}

func (s *Scanner) finishScan(ctx context.Context, progressDone <-chan struct{}) {
	select {
	case <-progressDone:
	case <-ctx.Done():
	}

	if s.rateTicker != nil {
		s.rateTicker.Stop()
	}

	close(s.results)
}

func (s *Scanner) worker(ctx context.Context, jobs <-chan scanJob) {
	defer s.wg.Done()

	dialer := &net.Dialer{Timeout: s.config.Timeout}

	for job := range jobs {
		// Check context cancellation
		if ctx.Err() != nil {
			return
		}

		// Rate limiting at worker level
		if !s.waitForRate(ctx) {
			return
		}

		// Scan port inline
		result := s.performDial(ctx, dialer, job)
		if result != nil {
			s.emitResult(ctx, *result)
		}
	}
}

func (s *Scanner) performDial(ctx context.Context, dialer *net.Dialer, job scanJob) *ResultEvent {
	address := net.JoinHostPort(job.host, strconv.Itoa(int(job.port)))
	maxAttempts := s.config.MaxRetries + 1
	if maxAttempts <= 0 {
		maxAttempts = 1
	}

	var lastResult ResultEvent
	for attempt := 0; attempt < maxAttempts; attempt++ {
		start := time.Now()
		conn, err := dialer.DialContext(ctx, "tcp", address)
		duration := time.Since(start)

		result := ResultEvent{
			Host:     job.host,
			Port:     job.port,
			Duration: duration,
			Protocol: "tcp",
		}

		if err != nil {
			if ctx.Err() != nil {
				return nil
			}

			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				result.State = StateFiltered
				lastResult = result
				if attempt < maxAttempts-1 {
					if !s.sleepWithJitter(ctx, attempt) {
						return nil
					}
					continue
				}
			} else {
				result.State = StateClosed
				lastResult = result
				break
			}
		} else {
			result.State = StateOpen
			if s.config.BannerGrab {
				result.Banner = s.grabBanner(conn)
			}
			_ = conn.Close()
			return &result
		}
	}

	return &lastResult
}

func (s *Scanner) waitForRate(ctx context.Context) bool {
	if s.rateTicker == nil {
		return true
	}

	select {
	case <-ctx.Done():
		return false
	case <-s.rateTicker.C:
		return true
	}
}

func (s *Scanner) sleepWithJitter(ctx context.Context, attempt int) bool {
	wait := s.retryBackoff(attempt)
	if wait <= 0 {
		return true
	}

	timer := time.NewTimer(wait)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}

func (s *Scanner) retryBackoff(attempt int) time.Duration {
	if attempt < 0 {
		attempt = 0
	}

	base := time.Duration(attempt+1) * RetryBackoffBase
	if base > s.config.Timeout {
		base = s.config.Timeout
	}

	jitter := time.Duration(rand.Intn(RetryJitterRangeMs)+RetryJitterMinMs) * time.Millisecond
	return base + jitter
}

func (s *Scanner) emitResult(ctx context.Context, result ResultEvent) {
	evt := NewResultEvent(result)
	select {
	case s.results <- evt:
		s.progressReporter.IncrementCompleted()
	case <-ctx.Done():
	}
}

func (s *Scanner) grabBanner(conn net.Conn) string {
	_ = conn.SetReadDeadline(time.Now().Add(BannerGrabTimeout))
	buffer := make([]byte, BannerBufferSize)
	n, err := conn.Read(buffer)
	if err != nil || n == 0 {
		return ""
	}
	return string(buffer[:n])
}
