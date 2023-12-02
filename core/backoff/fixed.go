package backoff

import (
	"math"
	"math/rand"
	"time"
)

type fixedBackOff struct {
	duration   time.Duration
	count      int
	hysteresis float32
}

type fixedBackOffSession struct {
	fixedBackOff
	remaining int
}

func Fixed(duration time.Duration, count int, hysteresis float32) BackOff {
	return fixedBackOff{
		duration:   duration,
		count:      count,
		hysteresis: hysteresis,
	}
}

func (backoff fixedBackOff) Begin() BackOffSession {
	return &fixedBackOffSession{
		fixedBackOff: backoff,
		remaining:    backoff.count,
	}
}

func (session *fixedBackOffSession) BackOff() bool {
	if session.remaining <= 0 {
		return false
	}
	deviation := float64(session.hysteresis) * rand.Float64()
	sleepTime := time.Duration(math.Round(float64(int64(session.duration)) * (1.0 - deviation)))

	time.Sleep(sleepTime)
	session.remaining--
	return (session.remaining > 0)
}
