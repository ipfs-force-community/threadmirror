package xscraper

import (
	"fmt"
	"math/rand"
	"time"
)

// ScraperPool holds a shuffled copy of scrapers for reuse
type ScraperPool struct {
	scrapers []*XScraper
}

// NewScraperPool creates a new pool with shuffled scrapers
func NewScraperPool(scrapers []*XScraper) *ScraperPool {
	shuffled := make([]*XScraper, len(scrapers))
	copy(shuffled, scrapers)
	rand.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})

	return &ScraperPool{
		scrapers: shuffled,
	}
}

// TryWithResult tries an operation and returns result
func TryWithResult[T any](pool *ScraperPool, op func(*XScraper) (T, error)) (T, error) {
	return TryWithDelay(pool, op, time.Millisecond*300, time.Second*2)
}

// TryWithDelay tries an operation with delay between attempts
func TryWithDelay[T any](pool *ScraperPool, op func(*XScraper) (T, error), minDelay, maxDelay time.Duration) (T, error) {
	var zero T
	if len(pool.scrapers) == 0 {
		return zero, fmt.Errorf("no scrapers available")
	}

	var lastErr error
	for _, scraper := range pool.scrapers {
		// Add delay if configured
		if maxDelay > 0 {
			delay := minDelay
			if maxDelay > minDelay {
				jitter := time.Duration(rand.Int63n(int64(maxDelay - minDelay)))
				delay += jitter
			}
			time.Sleep(delay)
		}

		result, err := op(scraper)
		if err == nil {
			return result, nil
		}
		lastErr = err
	}

	return zero, fmt.Errorf("all scrapers failed, last error: %w", lastErr)
}
