package outbox

import "time"

type WorkerConfig struct {
	Interval       time.Duration
	BatchSize      int
	MaxAttempts    int
	PublishTimeout time.Duration
	StuckAfter     time.Duration
}

func DefaultWorkerConfig() WorkerConfig {
	return WorkerConfig{
		Interval:       5 * time.Second,
		BatchSize:      100,
		MaxAttempts:    5,
		PublishTimeout: 10 * time.Second,
		StuckAfter:     5 * time.Minute,
	}
}
