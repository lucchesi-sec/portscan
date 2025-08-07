package core

import (
	"context"
	"fmt"
	"net"
	"sync"
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
}

type ProgressEvent struct {
	Total     int
	Completed int
	Rate      float64
}

type Scanner struct {
	config      *Config
	results     chan interface{}
	workerPool  chan struct{}
	rateLimiter <-chan time.Time
	wg          sync.WaitGroup
}

type Config struct {
	Workers     int
	Timeout     time.Duration
	RateLimit   int
	BannerGrab  bool
	MaxRetries  int
}

func NewScanner(cfg *Config) *Scanner {
	if cfg.Workers <= 0 {
		cfg.Workers = 100
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = 200 * time.Millisecond
	}
	if cfg.RateLimit <= 0 {
		cfg.RateLimit = 7500
	}

	var rateLimiter <-chan time.Time
	if cfg.RateLimit > 0 {
		interval := time.Second / time.Duration(cfg.RateLimit)
		rateLimiter = time.Tick(interval)
	}

	return &Scanner{
		config:      cfg,
		results:     make(chan interface{}, 1000),
		workerPool:  make(chan struct{}, cfg.Workers),
		rateLimiter: rateLimiter,
	}
}

func (s *Scanner) Results() <-chan interface{} {
	return s.results
}

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
	
	// Now safe to close results
	close(s.results)
}

func (s *Scanner) worker(ctx context.Context, host string, jobs <-chan uint16) {
	defer s.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case port, ok := <-jobs:
			if !ok {
				return
			}
			
			select {
			case <-ctx.Done():
				return
			case s.workerPool <- struct{}{}:
				if s.rateLimiter != nil {
					select {
					case <-ctx.Done():
						<-s.workerPool
						return
					case <-s.rateLimiter:
					}
				}
				
				s.wg.Add(1)
				go func(p uint16) {
					defer s.wg.Done()
					defer func() { <-s.workerPool }()
					s.scanPort(ctx, host, p)
				}(port)
			}
		}
	}
}

func (s *Scanner) scanPort(ctx context.Context, host string, port uint16) {
	start := time.Now()
	address := fmt.Sprintf("%s:%d", host, port)
	
	dialer := &net.Dialer{
		Timeout: s.config.Timeout,
	}
	
	conn, err := dialer.DialContext(ctx, "tcp", address)
	duration := time.Since(start)
	
	result := ResultEvent{
		Host:     host,
		Port:     port,
		Duration: duration,
	}
	
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
		
		conn.Close()
	}
	
	select {
	case s.results <- result:
	case <-ctx.Done():
	}
}

func (s *Scanner) grabBanner(conn net.Conn) string {
	conn.SetReadDeadline(time.Now().Add(time.Second))
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
			elapsed := time.Since(startTime).Seconds()
			if elapsed == 0 {
				elapsed = 0.1
			}
			
			// This is a simplified progress - in production you'd track actual completions
			completed := int(elapsed * float64(s.config.RateLimit))
			if completed > total {
				completed = total
			}
			
			rate := float64(completed) / elapsed
			
			progress := ProgressEvent{
				Total:     total,
				Completed: completed,
				Rate:      rate,
			}
			
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