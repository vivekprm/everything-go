package main

import (
	"fmt"
	"time"
)

type Item struct {
	Title   string
	Channel string
	GUID    string
}

type Fetcher interface {
	Fetch() (items []Item, next time.Time, err error)
}

func Fetch(domain string) Fetcher {
	// Fetches items from domain
	f := &fetcherImpl{
		domain: domain,
	}
	return f
}

type Subscription interface {
	Updates() <-chan Item // Stream of items
	Close() error         // Shuts down the stream
}

func Subscribe(fetcher Fetcher) Subscription { // Converts Fetches to a stream
	s := &subscriptionImpl{
		fetcher: fetcher,
		updates: make(chan Item), // for updates
	}
	go s.loop()
	return s
}

func Merge(subs ...Subscription) Subscription { // Merges several streams
	return &subscriptionImpl{}
}

func main() {
	// Subscribe to some feeds, and create a merged update stream.
	merged := Merge(
		Subscribe(Fetch("blog.golang.org")),
		Subscribe(Fetch("googleblog.blogspot.com")),
		Subscribe(Fetch("googledevelopers.blogspot.com")),
	)

	// Close the subscriptions after some time.
	time.AfterFunc(3*time.Second, func() {
		fmt.Println("closed:", merged.Close())
	})

	// Print the stream.
	for it := range merged.Updates() {
		fmt.Println(it.Channel, it.Title)
	}

	panic("show me the stacks")
}
