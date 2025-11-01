// Package main - retry_pattern demonstrates wrapper pattern for automatic retries
package main

import (
	"errors"
	"fmt"
	"log"
	"math"
	"time"

	"github.com/seyedali-dev/gosaidsno/aspect"
)

// -------------------------------------------- Retry Wrapper --------------------------------------------

// WithRetry wraps a function with retry logic
func WithRetry[T any](fn func() (T, error), maxRetries int, baseDelay time.Duration) func() (T, error) {
	return func() (T, error) {
		var result T
		var err error

		for attempt := 0; attempt <= maxRetries; attempt++ {
			if attempt > 0 {
				delay := time.Duration(math.Pow(2, float64(attempt-1))) * baseDelay
				log.Printf("[RETRY] Attempt %d/%d failed, retrying in %v...", attempt, maxRetries, delay)
				time.Sleep(delay)
			}

			result, err = fn()
			if err == nil {
				if attempt > 0 {
					log.Printf("[RETRY] Success on attempt %d", attempt+1)
				}
				return result, nil
			}
		}

		log.Printf("[RETRY] Exhausted all %d retries", maxRetries)
		return result, err
	}
}

// -------------------------------------------- Setup --------------------------------------------

func setupAOP() {
	log.Println("=== Setting up Monitoring AOP ===")

	aspect.MustRegister("SendEmail")
	aspect.MustRegister("ProcessPayment")

	// Add timing for monitoring
	for _, fn := range []string{"SendEmail", "ProcessPayment"} {
		aspect.MustAddAdvice(fn, aspect.Advice{
			Type:     aspect.Before,
			Priority: 100,
			Handler: func(ctx *aspect.Context) error {
				ctx.Metadata["start"] = time.Now()
				return nil
			},
		})

		aspect.MustAddAdvice(fn, aspect.Advice{
			Type:     aspect.After,
			Priority: 100,
			Handler: func(ctx *aspect.Context) error {
				start := ctx.Metadata["start"].(time.Time)
				duration := time.Since(start)
				status := "SUCCESS"
				if ctx.Error != nil {
					status = "FAILED"
				}
				log.Printf("[MONITOR] %s %s in %v", ctx.FunctionName, status, duration)
				return nil
			},
		})
	}

	log.Println("=== AOP Setup Complete ===\n")
}

// -------------------------------------------- Business Logic --------------------------------------------

var emailAttempts = 0

func sendEmailImpl(to, subject string) error {
	emailAttempts++
	log.Printf("[EMAIL] Attempt %d - Sending to %s: %s", emailAttempts, to, subject)

	// Simulate transient failures
	if emailAttempts <= 2 {
		return errors.New("SMTP connection timeout")
	}

	log.Printf("[EMAIL] Successfully sent!")
	return nil
}

var paymentAttempts = 0

func processPaymentImpl(amount float64, cardToken string) (string, error) {
	paymentAttempts++
	log.Printf("[PAYMENT] Attempt %d - Processing $%.2f", paymentAttempts, amount)

	// Simulate various failure scenarios
	if paymentAttempts == 1 {
		return "", errors.New("network timeout")
	}
	if paymentAttempts == 2 {
		return "", errors.New("payment gateway unavailable")
	}

	txnID := fmt.Sprintf("txn_%d", time.Now().Unix())
	log.Printf("[PAYMENT] Success! Transaction ID: %s", txnID)
	return txnID, nil
}

// -------------------------------------------- Wrapped Functions with Retry --------------------------------------------

// SendEmail with retry wrapper
func SendEmail(to, subject string) error {
	// Wrap the base function with retry logic
	retryFn := WithRetry(func() (struct{}, error) {
		err := sendEmailBase(to, subject)
		return struct{}{}, err
	}, 3, 100*time.Millisecond)

	_, err := retryFn()
	return err
}

var sendEmailBase = aspect.Wrap2E("SendEmail", sendEmailImpl)

// ProcessPayment with retry wrapper
func ProcessPayment(amount float64, cardToken string) (string, error) {
	retryFn := WithRetry(func() (string, error) {
		return processPaymentBase(amount, cardToken)
	}, 5, 200*time.Millisecond)

	return retryFn()
}

var processPaymentBase = aspect.Wrap2RE("ProcessPayment", processPaymentImpl)

// -------------------------------------------- Examples --------------------------------------------

func example1_SuccessAfterRetries() {
	fmt.Println("\n========== Example 1: Success After Retries ==========\n")

	emailAttempts = 0
	start := time.Now()

	err := SendEmail("user@example.com", "Welcome!")
	duration := time.Since(start)

	if err != nil {
		fmt.Printf("❌ Email failed: %v\n", err)
	} else {
		fmt.Printf("\n✅ Email sent successfully after %d attempts in %v\n", emailAttempts, duration)
	}
}

func example2_PaymentWithRetries() {
	fmt.Println("\n========== Example 2: Payment with Retries ==========\n")

	paymentAttempts = 0
	start := time.Now()

	txnID, err := ProcessPayment(99.99, "card_token_123")
	duration := time.Since(start)

	if err != nil {
		fmt.Printf("❌ Payment failed: %v\n", err)
	} else {
		fmt.Printf("\n✅ Payment processed: %s (took %v with %d attempts)\n",
			txnID, duration, paymentAttempts)
	}
}

func example3_ExhaustedRetries() {
	fmt.Println("\n========== Example 3: Exhausted Retries ==========\n")

	// Function that always fails
	var failAttempts = 0
	FailingOperation := WithRetry(func() (string, error) {
		failAttempts++
		log.Printf("[OPERATION] Attempt %d - Executing", failAttempts)
		return "", errors.New("permanent failure")
	}, 3, 50*time.Millisecond)

	start := time.Now()
	_, err := FailingOperation()
	duration := time.Since(start)

	if err != nil {
		fmt.Printf("\n❌ Operation failed after %d attempts: %v (took %v)\n", failAttempts, err, duration)
	}
}

// -------------------------------------------- Main --------------------------------------------

func main() {
	setupAOP()

	example1_SuccessAfterRetries()
	example2_PaymentWithRetries()
	example3_ExhaustedRetries()

	fmt.Println("\n========== Retry Examples Complete ==========")
}
