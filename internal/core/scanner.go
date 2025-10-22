package core

import (
	"context"
	"math"
	"math/rand"
	"net"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

type Scanner struct {
	config     *Config
	results    chan Event
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
		results:    make(chan Event, 1000),
		rateTicker: ticker,
	}
}

func (s *Scanner) Results() <-chan Event { return s.results }

func (s *Scanner) ScanRange(ctx context.Context, host string, ports []uint16) {
	s.ScanTargets(ctx, []ScanTarget{{Host: host, Ports: ports}})
}

func (s *Scanner) ScanTargets(ctx context.Context, targets []ScanTarget) {
	totalPorts := totalPortCount(targets)
	if totalPorts == 0 {
		close(s.results)
		return
	}

	s.completed.Store(0)

	jobs := make(chan scanJob, totalPorts)
	progressDone := s.startProgressReporter(ctx, totalPorts)

	s.startWorkers(ctx, jobs)

	go s.feedJobs(ctx, jobs, targets)

	s.wg.Wait()

	s.finishScan(ctx, progressDone)
}

func (s *Scanner) startProgressReporter(ctx context.Context, total int) <-chan struct{} {
	done := make(chan struct{})
	go func() {
		s.progressReporter(ctx, total)
		close(done)
	}()
	return done
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

	for {
		select {
		case <-ctx.Done():
			return
		case job, ok := <-jobs:
			if !ok {
				return
			}

			s.scanPort(ctx, dialer, job)
		}
	}
}

func (s *Scanner) scanPort(ctx context.Context, dialer *net.Dialer, job scanJob) {
	result, ok := s.performDial(ctx, dialer, job)
	if !ok {
		return
	}
	s.emitResult(ctx, result)
}

func (s *Scanner) performDial(ctx context.Context, dialer *net.Dialer, job scanJob) (ResultEvent, bool) {
	address := net.JoinHostPort(job.host, strconv.Itoa(int(job.port)))
	maxAttempts := s.config.MaxRetries + 1
	if maxAttempts <= 0 {
		maxAttempts = 1
	}

	var lastResult ResultEvent
	for attempt := 0; attempt < maxAttempts; attempt++ {
		if !s.waitForRate(ctx) {
			return ResultEvent{}, false
		}

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
				return ResultEvent{}, false
			}

			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				result.State = StateFiltered
				lastResult = result
				if attempt < maxAttempts-1 {
					if !s.sleepWithJitter(ctx, attempt) {
						return ResultEvent{}, false
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
			return result, true
		}
	}

	return lastResult, true
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

	base := time.Duration(attempt+1) * 50 * time.Millisecond
	if base > s.config.Timeout {
		base = s.config.Timeout
	}

	jitter := time.Duration(rand.Intn(40)+10) * time.Millisecond
	return base + jitter
}

func (s *Scanner) emitResult(ctx context.Context, result ResultEvent) {
	evt := Event{Type: EventTypeResult, Result: result}
	select {
	case s.results <- evt:
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
			case s.results <- Event{Type: EventTypeProgress, Progress: progress}:
			case <-ctx.Done():
				return
			}

			if completed >= total {
				return
			}
		}
	}
}
