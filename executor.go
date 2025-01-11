package goutils

import (
	"sync"

	"github.com/rs/zerolog/log"
)

type LatestTaskExecutor struct {
	nextTask func()
	cond     *sync.Cond
	mu       sync.Mutex
}

func NewLatestTaskExecutor() *LatestTaskExecutor {
	return &LatestTaskExecutor{
		cond: sync.NewCond(&sync.Mutex{}),
	}
}

func (d *LatestTaskExecutor) AddTask(f func()) {
	if f == nil {
		log.Warn().Msg("Task is nil")
		return
	}

	d.setTask(f)

	d.cond.L.Lock()
	d.cond.Signal()
	d.cond.L.Unlock()
}

func (d *LatestTaskExecutor) Start() {
	go func() {
		for {
			task := d.getTask()
			if task == nil {
				d.cond.L.Lock()
				d.cond.Wait()
				d.cond.L.Unlock()
			}
			task = d.getTask()
			d.setTask(nil)
			task()
		}
	}()
}

func (d *LatestTaskExecutor) getTask() func() {
	d.mu.Lock()
	defer d.mu.Unlock()

	return d.nextTask
}

func (d *LatestTaskExecutor) setTask(f func()) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.nextTask = f
}
