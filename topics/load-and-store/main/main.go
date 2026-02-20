package main

import (
	"context"
	"log"
	"sync"
	"time"
	"load-and-store/jobqueue"
)

type task struct {
	id      int
	running bool
}

type DiscoveryJob struct {
	isJobRunning sync.Map
	stop         chan bool
	tasks        []*task
	jq           *jobqueue.JobQueue
}

func (j *DiscoveryJob) DiscoverJob(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	j.startDiscoveryJob(ctx)

	for {
		select {
		case <-j.stop:
			return
		case <-ticker.C:
			j.startDiscoveryJob(ctx)
		}
	}
}

func (j *DiscoveryJob) stopDiscoveryJob(taskID int) {
	j.isJobRunning.Store(taskID, false)
}

func (j *DiscoveryJob) startDiscoveryJob(ctx context.Context) {
	for _, task := range j.tasks {
		if _, loaded := j.isJobRunning.LoadOrStore(task.id, true); loaded {
			log.Printf("Job is already running ")

			continue
		}

		f := func() {
			start := time.Now()

			defer func() {
				j.stopDiscoveryJob(task.id)

				finishTime := time.Since(start).Seconds()
				log.Printf("finished discovery job for task %d, duration: %f seconds", task.id, finishTime)
			}()

			log.Printf("starting discovery job for task %d", task.id)

			doWork()
		}

		j.jq.NewJob(f)
	}
}

func (j *DiscoveryJob) Start(ctx context.Context) error {
	go j.DiscoverJob(ctx)

	return nil
}

func doWork() {
	log.Println("Sleeping")
	time.Sleep(10 * time.Second)
}

func main() {
	tasks := []*task{}
	for i := 0; i < 100; i++ {
		t := &task{
			id:      i,
			running: false,
		}
		tasks = append(tasks, t)
	}

	discoveryJob := &DiscoveryJob{
		isJobRunning: sync.Map{},
		stop:         make(chan bool),
		tasks:        tasks,
		jq:           NewJobQueue(10),
	}

	ctx := context.Background()
	discoveryJob.Start(ctx)
}
