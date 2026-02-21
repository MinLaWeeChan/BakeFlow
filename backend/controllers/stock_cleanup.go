package controllers

import (
	"log"
	"sync"
	"time"

	"bakeflow/models"
)

// StockCleanupJob manages periodic cleanup of expired reservations
type StockCleanupJob struct {
	interval time.Duration
	stopChan chan struct{}
	running  bool
	mu       sync.Mutex
}

var stockCleanupJob *StockCleanupJob

// StartStockCleanupJob starts the background job to release expired reservations
func StartStockCleanupJob(interval time.Duration) {
	if interval <= 0 {
		interval = 1 * time.Minute
	}

	stockCleanupJob = &StockCleanupJob{
		interval: interval,
		stopChan: make(chan struct{}),
	}

	go stockCleanupJob.run()
	log.Printf("📦 Stock cleanup job started (interval: %v)", interval)
}

// StopStockCleanupJob stops the background cleanup job
func StopStockCleanupJob() {
	if stockCleanupJob == nil {
		return
	}

	stockCleanupJob.mu.Lock()
	defer stockCleanupJob.mu.Unlock()

	if stockCleanupJob.running {
		close(stockCleanupJob.stopChan)
		stockCleanupJob.running = false
		log.Println("📦 Stock cleanup job stopped")
	}
}

func (j *StockCleanupJob) run() {
	j.mu.Lock()
	j.running = true
	j.mu.Unlock()

	ticker := time.NewTicker(j.interval)
	defer ticker.Stop()

	// Run immediately on start
	j.cleanup()

	for {
		select {
		case <-ticker.C:
			j.cleanup()
		case <-j.stopChan:
			return
		}
	}
}

func (j *StockCleanupJob) cleanup() {
	count, err := models.CleanupExpiredReservations()
	if err != nil {
		log.Printf("⚠️ Stock cleanup error: %v", err)
		return
	}
	if count > 0 {
		log.Printf("📦 Released %d expired stock reservations", count)
	}
}

// ManualCleanup triggers an immediate cleanup (for testing/admin)
func ManualStockCleanup() (int, error) {
	return models.CleanupExpiredReservations()
}
