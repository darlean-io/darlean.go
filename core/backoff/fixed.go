package backoff

import "time"

type fixedBackOff struct {
	duration time.Duration
	count    int
}

type fixedBackOffSession struct {
	fixedBackOff
	remaining int
}

func Fixed(duration time.Duration, count int, histeresis float32) BackOff {
	return fixedBackOff{
		duration: duration,
		count:    count,
	}
}

func (backoff fixedBackOff) Begin() BackOffSession {
	return fixedBackOffSession{
		fixedBackOff: backoff,
		remaining:    backoff.count,
	}
}

func (session fixedBackOffSession) BackOff() bool {
	if session.remaining <= 0 {
		return false
	}
	time.Sleep(session.duration)
	session.remaining--
	return (session.remaining > 0)
}
