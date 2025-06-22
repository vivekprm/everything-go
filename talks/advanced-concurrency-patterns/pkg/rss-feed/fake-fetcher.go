package main

import (
	"time"
)

type fetcherImpl struct {
	domain string
}

func (f fetcherImpl) Fetch() ([]Item, time.Time, error) {
	return []Item{
		{
			Title: f.domain,
		},
	}, time.Now(), nil
}
