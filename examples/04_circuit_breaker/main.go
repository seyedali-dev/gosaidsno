// Package main - circuit_breaker demonstrates Around advice for fault tolerance
package main

import (
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/seyedali-dev/gosaidsno/aspect"
)

// -------------------------------------------- Circuit Breaker --------------------------------------------

type CircuitBreaker struct {
	mu              sync.RWMutex
	failures        int
	maxFailures     int
	state           string // CLOSED, OPEN, HALF_OPEN
	openedAt        time.Time
	resetTimeout    time.Duration
	successesNeeded int
	successes       int
}

func NewCircuitBreaker(maxFailures int, resetTimeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		maxFailures:     maxFailures,
		state:           "CLOSED",
		resetTimeout:    resetTimeout,
		successesNeeded: 2,
	}
}

func (cb *CircuitBreaker) Call(fn func() error) error {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	// Check if we should transition from OPEN to HALF_OPEN
	if cb.state == "OPEN" && time.Since(cb.openedAt) > cb.resetTimeout {
		log.Printf("[CIRCUIT] Transitioning OPEN -> HALF_OPEN")
		cb.state = "HALF_OPEN"
		cb.successes = 0
	}

	// Reject if circuit is OPEN
	if cb.state == "OPEN" {
		remaining := cb.resetTimeout - time.Since(cb.openedAt)
		return fmt.Errorf("circuit breaker OPEN, retry in %v", remaining.Round(time.Second))
	}

	// Execute function
	err := fn()

	if err != nil {
		cb.failures++
		log.Printf("[CIRCUIT] Failure recorded (%d/%d)", cb.failures, cb.maxFailures)

		if cb.failures >= cb.maxFailures {
			cb.state = "OPEN"
			cb.openedAt = time.Now()
			log.Printf("[CIRCUIT] Circuit OPENED due to failures")
		}
		return err
	}

	// Success
	if cb.state == "HALF_OPEN" {
		cb.successes++
		log.Printf("[CIRCUIT] Success in HALF_OPEN (%d/%d)", cb.successes, cb.successesNeeded)

		if cb.successes >= cb.successesNeeded {
			cb.state = "CLOSED"
			cb.failures = 0
			log.Printf("[CIRCUIT] Circuit CLOSED - system recovered")
		}
	} else if cb.state == "CLOSED" {
		cb.failures = 0 // Reset on success
	}

	return nil
}

var externalServiceCircuit = NewCircuitBreaker(3, 5*time.Second)

// -------------------------------------------- Setup --------------------------------------------

func setupAOP() {
	log.Println("=== Setting up Circuit Breaker AOP ===")

	aspect.MustRegister("CallExternalService")

	// Around advice: circuit breaker
	aspect.MustAddAdvice("CallExternalService", aspect.Advice{
		Type:     aspect.Around,
		Priority: 100,
		Handler: func(ctx *aspect.Context) error {
			// Check circuit state
			externalServiceCircuit.mu.RLock()
			state := externalServiceCircuit.state
			externalServiceCircuit.mu.RUnlock()

			log.Printf("[CIRCUIT] Current state: %s", state)

			if state == "OPEN" {
				remaining := externalServiceCircuit.resetTimeout - time.Since(externalServiceCircuit.openedAt)
				log.Printf("[CIRCUIT] Rejecting call - circuit OPEN")
				ctx.SetResult(0, "")
				ctx.Error = fmt.Errorf("circuit breaker OPEN, retry in %v", remaining.Round(time.Second))
				ctx.Skipped = true
				return nil
			}

			return nil // Allow execution
		},
	})

	// After advice: track failures
	aspect.MustAddAdvice("CallExternalService", aspect.Advice{
		Type:     aspect.After,
		Priority: 100,
		Handler: func(ctx *aspect.Context) error {
			if ctx.Error != nil && !ctx.Skipped {
				externalServiceCircuit.mu.Lock()
				externalServiceCircuit.failures++
				log.Printf("[CIRCUIT] Failure recorded (%d/%d)",
					externalServiceCircuit.failures, externalServiceCircuit.maxFailures)

				if externalServiceCircuit.failures >= externalServiceCircuit.maxFailures {
					externalServiceCircuit.state = "OPEN"
					externalServiceCircuit.openedAt = time.Now()
					log.Printf("[CIRCUIT] Circuit OPENED")
				}
				externalServiceCircuit.mu.Unlock()
			} else if ctx.Error == nil && !ctx.Skipped {
				// Success
				externalServiceCircuit.mu.Lock()
				if externalServiceCircuit.state == "HALF_OPEN" {
					externalServiceCircuit.successes++
					log.Printf("[CIRCUIT] Success in HALF_OPEN (%d/%d)",
						externalServiceCircuit.successes, externalServiceCircuit.successesNeeded)

					if externalServiceCircuit.successes >= externalServiceCircuit.successesNeeded {
						externalServiceCircuit.state = "CLOSED"
						externalServiceCircuit.failures = 0
						log.Printf("[CIRCUIT] Circuit CLOSED - recovered")
					}
				} else {
					externalServiceCircuit.failures = 0
				}
				externalServiceCircuit.mu.Unlock()
			}
			return nil
		},
	})

	log.Println("=== AOP Setup Complete ===\n")
}

// -------------------------------------------- Business Logic --------------------------------------------

var simulateFailure = true
var callCount = 0

func callExternalServiceImpl(endpoint string) (string, error) {
	callCount++
	log.Printf("[API] Calling external service: %s (call #%d)", endpoint, callCount)

	// Simulate flaky service
	time.Sleep(100 * time.Millisecond)

	if simulateFailure && callCount <= 3 {
		return "", errors.New("service unavailable (500)")
	}

	return fmt.Sprintf("Response from %s", endpoint), nil
}

// -------------------------------------------- Wrapped Functions --------------------------------------------

var CallExternalService = aspect.Wrap1RE("CallExternalService", callExternalServiceImpl)

// -------------------------------------------- Examples --------------------------------------------

func example1_CircuitOpenDueToFailures() {
	fmt.Println("\n========== Example 1: Circuit Opens Due to Failures ==========\n")

	simulateFailure = true
	callCount = 0

	for i := 1; i <= 5; i++ {
		fmt.Printf("\n--- Call %d ---\n", i)
		result, err := CallExternalService("/api/data")
		if err != nil {
			fmt.Printf("❌ Call failed: %v\n", err)
		} else {
			fmt.Printf("✅ Call succeeded: %s\n", result)
		}
		time.Sleep(200 * time.Millisecond)
	}
}

func example2_CircuitRecovery() {
	fmt.Println("\n========== Example 2: Circuit Recovery ==========\n")

	fmt.Println("--- Waiting for circuit to enter HALF_OPEN (5s) ---")
	time.Sleep(6 * time.Second)

	// Service is now healthy
	simulateFailure = false
	callCount = 0

	for i := 1; i <= 4; i++ {
		fmt.Printf("\n--- Recovery Call %d ---\n", i)
		result, err := CallExternalService("/api/data")
		if err != nil {
			fmt.Printf("❌ Call failed: %v\n", err)
		} else {
			fmt.Printf("✅ Call succeeded: %s\n", result)
		}
		time.Sleep(200 * time.Millisecond)
	}
}

func example3_RealWorldScenario() {
	fmt.Println("\n========== Example 3: Real World Scenario ==========\n")

	// Reset circuit
	externalServiceCircuit.mu.Lock()
	externalServiceCircuit.state = "CLOSED"
	externalServiceCircuit.failures = 0
	externalServiceCircuit.mu.Unlock()

	simulateFailure = false
	callCount = 0

	// Normal operation
	fmt.Println("--- Normal operation ---")
	for i := 1; i <= 2; i++ {
		result, _ := CallExternalService("/api/users")
		fmt.Printf("✅ %s\n", result)
		time.Sleep(100 * time.Millisecond)
	}

	// Service becomes unhealthy
	fmt.Println("\n--- Service degradation ---")
	simulateFailure = true
	for i := 1; i <= 4; i++ {
		_, err := CallExternalService("/api/users")
		if err != nil {
			fmt.Printf("❌ %v\n", err)
		}
		time.Sleep(100 * time.Millisecond)
	}

	// Service recovers
	fmt.Println("\n--- Waiting for recovery window (5s) ---")
	time.Sleep(6 * time.Second)

	simulateFailure = false
	callCount = 0
	fmt.Println("\n--- Service recovered ---")
	for i := 1; i <= 3; i++ {
		result, err := CallExternalService("/api/users")
		if err != nil {
			fmt.Printf("❌ %v\n", err)
		} else {
			fmt.Printf("✅ %s\n", result)
		}
		time.Sleep(100 * time.Millisecond)
	}
}

// -------------------------------------------- Main --------------------------------------------

func main() {
	setupAOP()

	example1_CircuitOpenDueToFailures()
	example2_CircuitRecovery()
	example3_RealWorldScenario()

	fmt.Println("\n========== Circuit Breaker Examples Complete ==========")
}
