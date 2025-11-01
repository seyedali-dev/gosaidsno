// Package main - caching_pattern demonstrates Around advice for caching expensive operations
package main

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/seyedali-dev/gosaidsno/aspect"
)

// -------------------------------------------- Simple Cache --------------------------------------------

type Cache struct {
	mu   sync.RWMutex
	data map[string]interface{}
	hits int
	miss int
}

func NewCache() *Cache {
	return &Cache{data: make(map[string]interface{})}
}

func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	val, ok := c.data[key]
	if ok {
		c.hits++
	} else {
		c.miss++
	}
	return val, ok
}

func (c *Cache) Set(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[key] = value
}

func (c *Cache) Stats() (hits, miss int) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.hits, c.miss
}

var userCache = NewCache()

// -------------------------------------------- Setup --------------------------------------------

func setupAOP() {
	log.Println("=== Setting up Caching AOP ===")

	aspect.MustRegister("FetchUserProfile")
	aspect.MustRegister("CalculateRecommendations")

	// Around advice for caching
	aspect.MustAddAdvice("FetchUserProfile", aspect.Advice{
		Type:     aspect.Around,
		Priority: 100,
		Handler: func(ctx *aspect.Context) error {
			userID := ctx.Args[0].(string)
			cacheKey := "user:" + userID

			// Check cache
			if cached, ok := userCache.Get(cacheKey); ok {
				log.Printf("[CACHE HIT] FetchUserProfile(%s)", userID)
				ctx.SetResult(0, cached)
				ctx.Skipped = true // Skip expensive DB call
				return nil
			}

			log.Printf("[CACHE MISS] FetchUserProfile(%s) - will fetch from DB", userID)
			return nil // Allow function to execute
		},
	})

	// AfterReturning to populate cache
	aspect.MustAddAdvice("FetchUserProfile", aspect.Advice{
		Type:     aspect.AfterReturning,
		Priority: 100,
		Handler: func(ctx *aspect.Context) error {
			userID := ctx.Args[0].(string)
			profile := ctx.Results[0].(*UserProfile)
			cacheKey := "user:" + userID

			userCache.Set(cacheKey, profile)
			log.Printf("[CACHE SET] Stored user profile for %s", userID)
			return nil
		},
	})

	// Time-based cache with TTL
	cacheStore := make(map[string]cacheEntry)
	var cacheMu sync.RWMutex

	aspect.MustAddAdvice("CalculateRecommendations", aspect.Advice{
		Type:     aspect.Around,
		Priority: 100,
		Handler: func(ctx *aspect.Context) error {
			userID := ctx.Args[0].(string)

			cacheMu.RLock()
			entry, exists := cacheStore[userID]
			cacheMu.RUnlock()

			// Check if cache is valid (not expired)
			if exists && time.Since(entry.timestamp) < 5*time.Second {
				log.Printf("[CACHE HIT] CalculateRecommendations(%s) - using cached", userID)
				ctx.SetResult(0, entry.value)
				ctx.Skipped = true
				return nil
			}

			if exists {
				log.Printf("[CACHE EXPIRED] CalculateRecommendations(%s) - recalculating", userID)
			} else {
				log.Printf("[CACHE MISS] CalculateRecommendations(%s) - calculating", userID)
			}
			return nil
		},
	})

	aspect.MustAddAdvice("CalculateRecommendations", aspect.Advice{
		Type:     aspect.AfterReturning,
		Priority: 100,
		Handler: func(ctx *aspect.Context) error {
			userID := ctx.Args[0].(string)
			recommendations := ctx.Results[0].([]string)

			cacheMu.Lock()
			cacheStore[userID] = cacheEntry{
				value:     recommendations,
				timestamp: time.Now(),
			}
			cacheMu.Unlock()

			log.Printf("[CACHE SET] Stored recommendations for %s (TTL: 5s)", userID)
			return nil
		},
	})

	log.Println("=== AOP Setup Complete ===\n")
}

type cacheEntry struct {
	value     interface{}
	timestamp time.Time
}

// -------------------------------------------- Domain Models --------------------------------------------

type UserProfile struct {
	ID        string
	Name      string
	Interests []string
}

// -------------------------------------------- Business Logic --------------------------------------------

func fetchUserProfileImpl(userID string) (*UserProfile, error) {
	// Simulate expensive database query
	log.Printf("[DB] Executing expensive query for user %s...", userID)
	time.Sleep(200 * time.Millisecond)

	return &UserProfile{
		ID:        userID,
		Name:      "User " + userID,
		Interests: []string{"coding", "music", "travel"},
	}, nil
}

func calculateRecommendationsImpl(userID string) ([]string, error) {
	// Simulate expensive ML calculation
	log.Printf("[ML] Running recommendation engine for user %s...", userID)
	time.Sleep(300 * time.Millisecond)

	return []string{
		"Product A",
		"Product B",
		"Product C",
	}, nil
}

// -------------------------------------------- Wrapped Functions --------------------------------------------

var (
	FetchUserProfile         = aspect.Wrap1RE("FetchUserProfile", fetchUserProfileImpl)
	CalculateRecommendations = aspect.Wrap1RE("CalculateRecommendations", calculateRecommendationsImpl)
)

// -------------------------------------------- Examples --------------------------------------------

func example1_BasicCaching() {
	fmt.Println("\n========== Example 1: Basic Caching ==========\n")

	userID := "user_123"

	// First call - cache miss, executes DB query
	start := time.Now()
	profile1, _ := FetchUserProfile(userID)
	duration1 := time.Since(start)
	fmt.Printf("First call: %s, took %v\n", profile1.Name, duration1)

	// Second call - cache hit, skips DB query
	start = time.Now()
	profile2, _ := FetchUserProfile(userID)
	duration2 := time.Since(start)
	fmt.Printf("Second call: %s, took %v (%.1fx faster)\n\n",
		profile2.Name, duration2, float64(duration1)/float64(duration2))

	hits, miss := userCache.Stats()
	fmt.Printf("Cache Stats: %d hits, %d miss\n", hits, miss)
}

func example2_TTLCache() {
	fmt.Println("\n========== Example 2: TTL-Based Caching ==========\n")

	userID := "user_456"

	// First call - miss
	fmt.Println("--- First call ---")
	recs1, _ := CalculateRecommendations(userID)
	fmt.Printf("Recommendations: %v\n\n", recs1)

	// Second call within TTL - hit
	fmt.Println("--- Second call (within 5s TTL) ---")
	recs2, _ := CalculateRecommendations(userID)
	fmt.Printf("Recommendations: %v\n\n", recs2)

	// Wait for cache to expire
	fmt.Println("--- Waiting for cache to expire (5s) ---")
	time.Sleep(6 * time.Second)

	// Third call after TTL - expired, recalculates
	fmt.Println("--- Third call (after TTL expired) ---")
	recs3, _ := CalculateRecommendations(userID)
	fmt.Printf("Recommendations: %v\n", recs3)
}

func example3_PerformanceComparison() {
	fmt.Println("\n========== Example 3: Performance Impact ==========\n")

	const iterations = 5

	fmt.Printf("Fetching user profile %d times...\n", iterations)

	totalWithCache := time.Duration(0)
	for i := 0; i < iterations; i++ {
		start := time.Now()
		_, _ = FetchUserProfile("user_789")
		totalWithCache += time.Since(start)
	}

	hits, miss := userCache.Stats()
	fmt.Printf("\nTotal time: %v\n", totalWithCache)
	fmt.Printf("Average per call: %v\n", totalWithCache/iterations)
	fmt.Printf("Cache efficiency: %d hits, %d miss (%.1f%% hit rate)\n",
		hits, miss, float64(hits)/float64(hits+miss)*100)
}

// -------------------------------------------- Main --------------------------------------------

func main() {
	setupAOP()

	example1_BasicCaching()
	example2_TTLCache()
	example3_PerformanceComparison()

	fmt.Println("\n========== Caching Examples Complete ==========")
}
