package utils

import (
	"errors"
	"log"
	"time"
)

func DoWithTries(fn func() error, attempts int, delay time.Duration) error {
	for attempts > 0 {
		err := errors.New("Failed to connect")
		if err = fn(); err != nil {
			log.Println("error while doing connection, err:", err)
			time.Sleep(delay)
			attempts--
			continue
		}

		return nil
	}

	return nil
}
