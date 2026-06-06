package utilities

import (
	"time"
)

func Retry(numOfRetry int, backoff time.Duration, callback func() error) error {
	var err error

	for i := 0; i < numOfRetry; i++ {
		if i > 0 {
			time.Sleep(backoff)
		}

		err = callback()
		if err == nil {
			return nil
		}
	}

	return err
}
