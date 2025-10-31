package core

import (
	"context"
	"errors"
	"math/rand"
	"net"
	"os"
	"strconv"
	"sync"
	"syscall"
	"time"
)

// UDPScanner handles UDP port scanning operations.
type UDPScanner struct {
	*Scanner
	serviceProbes map[uint16][]byte
	customProbes  map[uint16][]byte
	probeStats    map[uint16]ProbeStats
	probeMu       sync.RWMutex
}

// NewUDPScanner creates a new UDP scanner instance.
func NewUDPScanner(cfg *Config) *UDPScanner {
	return &UDPScanner{
		Scanner:       NewScanner(cfg),
		serviceProbes: initUDPProbes(),
		customProbes:  make(map[uint16][]byte),
		probeStats:    make(map[uint16]ProbeStats),
	}
}

// ScanRange implements the Scanner interface for UDP scanning.
func (s *UDPScanner) ScanRange(ctx context.Context, host string, ports []uint16) {
	s.ScanTargets(ctx, []ScanTarget{{Host: host, Ports: ports}})
}

func (s *UDPScanner) udpWorker(ctx context.Context, jobs <-chan scanJob) {
	defer s.wg.Done()

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	for {
		select {
		case <-ctx.Done():
			return
		case job, ok := <-jobs:
			if !ok {
				return
			}

			if s.rateTicker != nil {
				select {
				case <-ctx.Done():
					return
				case <-s.rateTicker.C:
					if s.config.UDPJitterMaxMs > 0 {
						jitter := time.Duration(rng.Intn(s.config.UDPJitterMaxMs)) * time.Millisecond
						if jitter > 0 {
							timer := time.NewTimer(jitter)
							select {
							case <-ctx.Done():
								timer.Stop()
								return
							case <-timer.C:
							}
						}
					}
				}
			}

			s.scanUDPPort(ctx, job.host, job.port)
		}
	}
}

func (s *UDPScanner) scanUDPPort(ctx context.Context, host string, port uint16) {
	start := time.Now()
	address := net.JoinHostPort(host, strconv.Itoa(int(port)))

	dialer := &net.Dialer{Timeout: s.config.Timeout}
	conn, err := dialer.DialContext(ctx, "udp", address)
	if err != nil {
		if ctx.Err() != nil {
			return
		}

		s.recordProbeAttempt(port, false)

		result := ResultEvent{
			Host:     host,
			Port:     port,
			State:    StateFiltered,
			Protocol: "udp",
			Duration: time.Since(start),
		}
		s.emitResult(ctx, result)
		return
	}
	defer func() { _ = conn.Close() }()

	_ = conn.SetReadDeadline(time.Now().Add(s.config.UDPReadTimeout))

	probe := s.getProbeForPort(port)
	if _, err = conn.Write(probe); err != nil {
		if ctx.Err() != nil {
			return
		}

		s.recordProbeAttempt(port, false)

		result := ResultEvent{
			Host:     host,
			Port:     port,
			State:    StateFiltered,
			Protocol: "udp",
			Duration: time.Since(start),
		}
		s.emitResult(ctx, result)
		return
	}

	buffer := make([]byte, s.config.UDPBufferSize)
	n, err := conn.Read(buffer)
	if ctx.Err() != nil {
		return
	}

	result := ResultEvent{
		Host:     host,
		Port:     port,
		Protocol: "udp",
		Duration: time.Since(start),
	}

	if err != nil {
		if ctx.Err() != nil {
			return
		}

		s.recordProbeAttempt(port, false)

		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			result.State = StateFiltered
		} else {
			var syscallErr *os.SyscallError
			if errors.As(err, &syscallErr) {
				switch syscallErr.Err {
				case syscall.ECONNREFUSED:
					result.State = StateClosed
				case syscall.EHOSTUNREACH, syscall.ENETUNREACH:
					result.State = StateFiltered
				default:
					result.State = StateFiltered
				}
			} else {
				result.State = StateClosed
			}
		}
	} else {
		s.recordProbeAttempt(port, true)
		result.State = StateOpen
		if n > 0 && s.config.BannerGrab {
			result.Banner = s.parseUDPResponse(port, buffer[:n])
		}
	}

	s.emitResult(ctx, result)
}
