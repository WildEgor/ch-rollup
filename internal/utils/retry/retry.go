package retry

import "time"

func Retry(retries int64, timeout time.Duration, f func(attemptNum int64) error) error {
	var (
		err     error
		retried int64
	)
	for {
		err = f(retried + 1)
		if err == nil {
			break
		}
		retried++
		if retried > retries {
			break
		}
		time.Sleep(timeout)
	}

	return err
}
