package main

import (
	"time"
)

// Implements subscription interface
type subscriptionImpl struct {
	fetcher Fetcher   // fetches Item
	updates chan Item // delivers items to the user
	closed  bool
	err     error
}

func (s *subscriptionImpl) Updates() <-chan Item {
	return s.updates
}

func (s *subscriptionImpl) Close() error {
	s.closed = true
	return s.err
}

// loop fetches items using s.fetcher and sends them
// on s.updates. loop exits when s.Close is called.
func (s *subscriptionImpl) loop() {
	for {
		if s.closed {
			close(s.updates)
			return
		}
		items, next, err := s.fetcher.Fetch()
		if err != nil {
			s.err = err
			time.Sleep(10 * time.Second)
			continue
		}
		for _, item := range items {
			s.updates <- item
		}
		if now := time.Now(); next.After(now) {
			time.Sleep(next.Sub(now))
		}
	}
}
