package goutils_test

import (
	"sync"
	"testing"
	"time"

	"github.com/117503445/goutils"
	"github.com/stretchr/testify/assert"
)

func TestLatestTaskExecutor_Basic(t *testing.T) {
	executor := goutils.NewLatestTaskExecutor()
	executor.Start()

	var executed bool
	task := func() { executed = true }

	executor.AddTask(task)
	time.Sleep(10 * time.Millisecond) // 等待足够长的时间以确保任务被执行

	assert.True(t, executed)
}

func TestLatestTaskExecutor_TaskOverride(t *testing.T) {
	executor := goutils.NewLatestTaskExecutor()
	executor.Start()

	var counter int
	task1 := func() { counter++ }
	task2 := func() { counter += 2 }

	executor.AddTask(task1)
	executor.AddTask(task2)
	time.Sleep(10 * time.Millisecond) // 确保第二个任务被执行

	assert.Equal(t, 2, counter)
}

func TestLatestTaskExecutor_ConcurrentSafety(t *testing.T) {
	executor := goutils.NewLatestTaskExecutor()
	executor.Start()

	var wg sync.WaitGroup
	var mu sync.Mutex
	var counter int

	addTask := func(i int) {
		defer wg.Done()
		task := func() {
			mu.Lock()
			defer mu.Unlock()
			counter += i
		}
		executor.AddTask(task)
	}

	for i := 0; i < 100; i++ {
		wg.Add(1)
		addTask(i)
	}

	wg.Wait()
	time.Sleep(10 * time.Millisecond) // 确保最后一个添加的任务被执行

	mu.Lock()
	defer mu.Unlock()
	assert.Equal(t, 99, counter) // 最后一个任务应该是i=99
}

func TestLatestTaskExecutor_NilTask(t *testing.T) {
	executor := goutils.NewLatestTaskExecutor()
	executor.Start()

	executor.AddTask(nil)
	time.Sleep(10 * time.Millisecond) // 确保没有任务被执行

	// 没有可验证的副作用，但应该不会panic
}

func TestLatestTaskExecutor_FastConsecutiveTasks(t *testing.T) {
	executor := goutils.NewLatestTaskExecutor()
	executor.Start()

	var lastExecuted int
	task := func(n int) {
		lastExecuted = n
	}

	for i := 0; i < 1000; i++ {
		executor.AddTask(func() { task(i) })
	}

	time.Sleep(10 * time.Millisecond) // 确保最后一个任务被执行

	assert.Equal(t, 999, lastExecuted)
}
