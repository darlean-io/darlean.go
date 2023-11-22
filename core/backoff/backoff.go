package backoff

type BackOffSession interface {
	BackOff() bool
}

type BackOff interface {
	Begin() BackOffSession
}
