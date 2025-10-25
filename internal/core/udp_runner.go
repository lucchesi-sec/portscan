package core

import "context"

// ScanTargets runs UDP scans across the provided host targets.
func (s *UDPScanner) ScanTargets(ctx context.Context, targets []ScanTarget) {
	totalPorts := totalPortCount(targets)
	if totalPorts == 0 {
		close(s.results)
		return
	}

	s.progressReporter.SetCompleted(0)

	jobs := make(chan scanJob, totalPorts)
	progressDone := s.progressReporter.StartReporting(ctx, totalPorts)

	s.startUDPWorkers(ctx, jobs)

	go s.feedJobs(ctx, jobs, targets)

	s.wg.Wait()

	s.finishScan(ctx, progressDone)
}

// startUDPWorkers launches a UDP-specific worker pool honouring the configured ratio.
func (s *UDPScanner) startUDPWorkers(ctx context.Context, jobs <-chan scanJob) {
	workerCount := s.computeUDPWorkerCount()
	for i := 0; i < workerCount; i++ {
		s.wg.Add(1)
		go s.udpWorker(ctx, jobs)
	}
}

// computeUDPWorkerCount determines the number of UDP workers to spawn.
func (s *UDPScanner) computeUDPWorkerCount() int {
	switch {
	case s.config.UDPWorkerRatio < 0:
		if s.config.Workers/2 < 1 {
			return 1
		}
		return s.config.Workers / 2
	case s.config.UDPWorkerRatio == 0:
		return 1
	default:
		count := int(float64(s.config.Workers) * s.config.UDPWorkerRatio)
		if count < 1 {
			count = 1
		}
		return count
	}
}

// helper shared with TCP scanner; re-export for UDP usage.
