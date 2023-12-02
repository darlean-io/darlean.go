package backoff

import (
	"math"
	"math/rand"
	"time"
)

type exponentialBackOff struct {
	baseDuration time.Duration
	factor       float32
	hysteresis   float32
	count        int
}

type exponentialBackOffSession struct {
	exponentialBackOff
	nextDuration time.Duration
	remaining    int
}

func Exponential(baseDuration time.Duration, count int, factor float32, hysteresis float32) BackOff {
	return exponentialBackOff{
		baseDuration: baseDuration,
		factor:       factor,
		hysteresis:   hysteresis,
		count:        count,
	}
}

func (backoff exponentialBackOff) Begin() BackOffSession {
	return &exponentialBackOffSession{
		exponentialBackOff: backoff,
		nextDuration:       backoff.baseDuration,
		remaining:          backoff.count,
	}
}

func (session *exponentialBackOffSession) BackOff() bool {
	if session.remaining <= 0 {
		return false
	}
	deviation := float64(session.hysteresis) * rand.Float64()

	sleepTime := time.Duration(math.Round(float64(int64(session.nextDuration)) * (1.0 - deviation)))
	time.Sleep(sleepTime)
	session.nextDuration = time.Duration(session.factor * float32(session.nextDuration))
	session.remaining--
	return (session.remaining > 0)
}
