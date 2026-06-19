package lowleveldesign

import (
	"container/heap"
	"context"
	"fmt"
	"sync"
	"time"
)

// =============================================================================
// LLD PROBLEM: TASK SCHEDULER (Priority Queue + Worker Pool)
// =============================================================================
// Design a task scheduler that:
// 1. Accepts tasks with priority levels
// 2. Executes highest-priority tasks first
// 3. Supports delayed/scheduled tasks
// 4. Retries failed tasks with exponential backoff
// 5. Limits concurrent execution
// 6. Supports task cancellation
// 7. Provides task status tracking
//
// Patterns Used:
// - Priority Queue (heap)
// - Worker Pool (bounded concurrency)
// - State Machine (task lifecycle)
// - Observer (status updates)
// - Strategy (retry policies)

// =============================================================================
// DOMAIN MODELS
// =============================================================================

type TaskStatus int

const (
	TaskPending TaskStatus = iota
	TaskScheduled
	TaskRunning
	TaskCompleted
	TaskFailed
	TaskCancelled
)

func (s TaskStatus) String() string {
	return [...]string{"PENDING", "SCHEDULED", "RUNNING", "COMPLETED", "FAILED", "CANCELLED"}[s]
}

type Priority int

const (
	PriorityLow      Priority = 0
	PriorityMedium   Priority = 5
	PriorityHigh     Priority = 10
	PriorityCritical Priority = 15
)

type Task struct {
	ID          string
	Name        string
	Priority    Priority
	ExecuteAt   time.Time // when to execute (for delayed tasks)
	MaxRetries  int
	Attempts    int
	Status      TaskStatus
	Handler     func(ctx context.Context) error
	CreatedAt   time.Time
	StartedAt   time.Time
	CompletedAt time.Time
	Error       error
	cancelFunc  context.CancelFunc
}

// =============================================================================
// PRIORITY QUEUE (Min-Heap by ExecuteAt, then by Priority desc)
// =============================================================================

type TaskQueue []*Task

func (q TaskQueue) Len() int { return len(q) }
func (q TaskQueue) Less(i, j int) bool {
	// First: earlier ExecuteAt wins
	if !q[i].ExecuteAt.Equal(q[j].ExecuteAt) {
		return q[i].ExecuteAt.Before(q[j].ExecuteAt)
	}
	// Then: higher priority wins
	return q[i].Priority > q[j].Priority
}
func (q TaskQueue) Swap(i, j int)       { q[i], q[j] = q[j], q[i] }
func (q *TaskQueue) Push(x interface{}) { *q = append(*q, x.(*Task)) }
func (q *TaskQueue) Pop() interface{} {
	old := *q
	n := len(old)
	task := old[n-1]
	old[n-1] = nil
	*q = old[:n-1]
	return task
}

// =============================================================================
// RETRY POLICY (Strategy Pattern)
// =============================================================================

type RetryPolicy interface {
	NextDelay(attempt int) time.Duration
	ShouldRetry(err error, attempt, maxRetries int) bool
}

type ExponentialBackoff struct {
	BaseDelay time.Duration
	MaxDelay  time.Duration
}

func (p *ExponentialBackoff) NextDelay(attempt int) time.Duration {
	delay := p.BaseDelay
	for i := 0; i < attempt; i++ {
		delay *= 2
		if delay > p.MaxDelay {
			return p.MaxDelay
		}
	}
	return delay
}

func (p *ExponentialBackoff) ShouldRetry(_ error, attempt, maxRetries int) bool {
	return attempt < maxRetries
}

// =============================================================================
// TASK SCHEDULER
// =============================================================================

type TaskScheduler struct {
	mu          sync.Mutex
	queue       TaskQueue
	tasks       map[string]*Task
	workers     int
	running     int
	retryPolicy RetryPolicy
	ctx         context.Context
	cancel      context.CancelFunc
	wakeup      chan struct{} // signal when new task is added
	statusCh    chan TaskStatusUpdate
}

type TaskStatusUpdate struct {
	TaskID string
	Status TaskStatus
	Error  error
}

func NewTaskScheduler(workers int, retryPolicy RetryPolicy) *TaskScheduler {
	ctx, cancel := context.WithCancel(context.Background())
	s := &TaskScheduler{
		tasks:       make(map[string]*Task),
		workers:     workers,
		retryPolicy: retryPolicy,
		ctx:         ctx,
		cancel:      cancel,
		wakeup:      make(chan struct{}, 1),
		statusCh:    make(chan TaskStatusUpdate, 100),
	}
	heap.Init(&s.queue)
	return s
}

// Submit adds a task to the scheduler
func (s *TaskScheduler) Submit(task *Task) string {
	s.mu.Lock()
	defer s.mu.Unlock()

	if task.ExecuteAt.IsZero() {
		task.ExecuteAt = time.Now()
	}
	task.CreatedAt = time.Now()
	task.Status = TaskPending
	s.tasks[task.ID] = task
	heap.Push(&s.queue, task)

	// Wake up the scheduler
	select {
	case s.wakeup <- struct{}{}:
	default:
	}

	fmt.Printf("[Scheduler] Task %q submitted (priority: %d, execute at: %v)\n",
		task.Name, task.Priority, task.ExecuteAt.Format("15:04:05"))
	return task.ID
}

// ScheduleAfter submits a task to run after a delay
func (s *TaskScheduler) ScheduleAfter(task *Task, delay time.Duration) string {
	task.ExecuteAt = time.Now().Add(delay)
	task.Status = TaskScheduled
	return s.Submit(task)
}

// Cancel cancels a pending or running task
func (s *TaskScheduler) Cancel(taskID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	task, exists := s.tasks[taskID]
	if !exists {
		return fmt.Errorf("task %s not found", taskID)
	}
	if task.Status == TaskCompleted || task.Status == TaskCancelled {
		return fmt.Errorf("task %s already in terminal state: %v", taskID, task.Status)
	}

	task.Status = TaskCancelled
	if task.cancelFunc != nil {
		task.cancelFunc()
	}
	return nil
}

// GetStatus returns current task status
func (s *TaskScheduler) GetStatus(taskID string) (*Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	task, exists := s.tasks[taskID]
	if !exists {
		return nil, fmt.Errorf("task %s not found", taskID)
	}
	return task, nil
}

// Start begins the scheduler loop
func (s *TaskScheduler) Start() {
	go s.schedulerLoop()
}

func (s *TaskScheduler) schedulerLoop() {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-s.wakeup:
			s.tryExecute()
		case <-ticker.C:
			s.tryExecute()
		}
	}
}

func (s *TaskScheduler) tryExecute() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for s.queue.Len() > 0 && s.running < s.workers {
		// Peek at the next task
		task := s.queue[0]

		// Not ready yet
		if task.ExecuteAt.After(now) {
			break
		}

		// Skip cancelled tasks
		if task.Status == TaskCancelled {
			heap.Pop(&s.queue)
			continue
		}

		// Execute
		heap.Pop(&s.queue)
		s.running++
		go s.executeTask(task)
	}
}

func (s *TaskScheduler) executeTask(task *Task) {
	// Create cancellable context for this task
	ctx, cancelFunc := context.WithCancel(s.ctx)
	task.cancelFunc = cancelFunc
	defer cancelFunc()

	// Update status
	s.mu.Lock()
	task.Status = TaskRunning
	task.StartedAt = time.Now()
	task.Attempts++
	s.mu.Unlock()

	fmt.Printf("[Scheduler] Executing task %q (attempt %d/%d)\n",
		task.Name, task.Attempts, task.MaxRetries+1)

	// Run the handler
	err := task.Handler(ctx)

	s.mu.Lock()
	defer s.mu.Unlock()
	s.running--

	if err != nil {
		task.Error = err
		// Check retry
		if s.retryPolicy.ShouldRetry(err, task.Attempts, task.MaxRetries) {
			delay := s.retryPolicy.NextDelay(task.Attempts)
			task.ExecuteAt = time.Now().Add(delay)
			task.Status = TaskPending
			heap.Push(&s.queue, task)
			fmt.Printf("[Scheduler] Task %q failed, retrying in %v (attempt %d)\n",
				task.Name, delay, task.Attempts)
		} else {
			task.Status = TaskFailed
			task.CompletedAt = time.Now()
			fmt.Printf("[Scheduler] Task %q FAILED permanently: %v\n", task.Name, err)
		}
	} else {
		task.Status = TaskCompleted
		task.CompletedAt = time.Now()
		fmt.Printf("[Scheduler] Task %q COMPLETED in %v\n",
			task.Name, task.CompletedAt.Sub(task.StartedAt))
	}

	// Signal for next task
	select {
	case s.wakeup <- struct{}{}:
	default:
	}
}

// Stop gracefully shuts down the scheduler
func (s *TaskScheduler) Stop() {
	s.cancel()
}

// Stats returns scheduler statistics
func (s *TaskScheduler) Stats() map[string]int {
	s.mu.Lock()
	defer s.mu.Unlock()

	stats := map[string]int{
		"queued":    s.queue.Len(),
		"running":   s.running,
		"total":     len(s.tasks),
		"completed": 0,
		"failed":    0,
		"cancelled": 0,
	}
	for _, task := range s.tasks {
		switch task.Status {
		case TaskCompleted:
			stats["completed"]++
		case TaskFailed:
			stats["failed"]++
		case TaskCancelled:
			stats["cancelled"]++
		}
	}
	return stats
}

// =============================================================================
// USAGE EXAMPLE
// =============================================================================

func ExampleTaskScheduler() {
	retryPolicy := &ExponentialBackoff{
		BaseDelay: 100 * time.Millisecond,
		MaxDelay:  5 * time.Second,
	}

	scheduler := NewTaskScheduler(3, retryPolicy) // 3 concurrent workers
	scheduler.Start()
	defer scheduler.Stop()

	// Submit tasks with different priorities
	scheduler.Submit(&Task{
		ID:         "task-1",
		Name:       "Send welcome email",
		Priority:   PriorityMedium,
		MaxRetries: 3,
		Handler: func(ctx context.Context) error {
			time.Sleep(50 * time.Millisecond)
			fmt.Println("    → Email sent!")
			return nil
		},
	})

	scheduler.Submit(&Task{
		ID:         "task-2",
		Name:       "Process payment",
		Priority:   PriorityCritical, // runs first!
		MaxRetries: 5,
		Handler: func(ctx context.Context) error {
			time.Sleep(30 * time.Millisecond)
			fmt.Println("    → Payment processed!")
			return nil
		},
	})

	scheduler.Submit(&Task{
		ID:         "task-3",
		Name:       "Generate report",
		Priority:   PriorityLow,
		MaxRetries: 1,
		Handler: func(ctx context.Context) error {
			time.Sleep(100 * time.Millisecond)
			fmt.Println("    → Report generated!")
			return nil
		},
	})

	// Schedule a delayed task
	scheduler.ScheduleAfter(&Task{
		ID:         "task-4",
		Name:       "Cleanup temp files",
		Priority:   PriorityLow,
		MaxRetries: 0,
		Handler: func(ctx context.Context) error {
			fmt.Println("    → Temp files cleaned!")
			return nil
		},
	}, 500*time.Millisecond)

	// Wait for tasks to complete
	time.Sleep(1 * time.Second)

	// Print stats
	stats := scheduler.Stats()
	fmt.Printf("\nScheduler Stats: %+v\n", stats)
}
