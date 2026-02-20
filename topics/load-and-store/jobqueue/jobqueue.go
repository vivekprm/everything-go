package main

import (
	"log"

	ants "github.com/panjf2000/ants/v2"
)

type JobQueue struct {
	Pool *ants.Pool
}

// JobQueueOption is used to configure option on JobQueue.
type JobQueueOption func(c *JobQueue)

func NewJobQueue(poolSize int, opts ...JobQueueOption) *JobQueue {
	pool, err := ants.NewPool(poolSize)
	if err != nil {
		log.Fatalf(err.Error())
	}

	log.Printf("created job pool with size %d", poolSize)

	jq := &JobQueue{
		Pool: pool,
	}
	for _, opt := range opts {
		opt(jq)
	}

	return jq
}

func (jq *JobQueue) NewJob(t func()) error {
	err := jq.Pool.Submit(t)
	if err != nil {
		log.Printf("failed to submit job to job pool: %v", err)

		return err
	}

	log.Printf("submitted job to job pool, pool_size=%d, running_jobs=%d, free_workers=%d",
		jq.Pool.Cap(),
		jq.Pool.Running(),
		jq.Pool.Free())

	return nil
}

func (jq *JobQueue) Stop() {
	jq.Pool.Release()
}
