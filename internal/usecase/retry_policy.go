package usecase

// RetryPolicy drives the background poller that re-calls the gateway for stuck communications.
type RetryPolicy struct {
	Enabled                   bool
	IntervalSeconds           int
	RetryableResponseStatuses []string
	MaxPollAttempts           int
	BatchLimit                int
}
