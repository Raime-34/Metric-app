package utils

import "time"

var delays = []int{
	1,
	3,
	5,
}

func WithRetry(f func() error) error {
	var err error

	for i := 0; i <= len(delays); i++ {
		if err = f(); err == nil {
			return nil
		}

		if i == len(delays) {
			break
		}
		time.Sleep(time.Duration(delays[i]) * time.Second)
	}

	return err
}
