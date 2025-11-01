// Package main - retry_pattern demonstrates Before/After advice for automatic retries
package main

import (
	"errors"
	"fmt"
	"log"
	"math"
	"time"

	"github.com/seyedali-dev/gosaidsno/aspect"
)

// -------------------------------------------- Setup --------------------------------------------

func setupAOP() {
	log.Println("=== Setting up Retry AOP ===")

	aspect.MustRegister("SendEmail")
	aspect.MustRegister("ProcessPayment")

	// Retry logic with exponential backoff
	setupRetry("SendEmail", 3, 100*time.Millisecond)
	setupRetry("ProcessPayment", 5, 200*time.Millisecond)

	log.Println("=== AOP Setup Complete ===\n")
}

func setupRetry(funcName string, maxRetries int, baseDelay time.Duration) {
	// Before: initialize retry metadata
	aspect.MustAddAdvice(funcName, aspect.Advice{
		Type:     aspect.Before,
		Priority: 100,
		Handler: func(ctx *aspect.Context) error {
			ctx.Metadata["attempt"] = 0
			ctx.Metadata["maxRetries"] = maxRetries
			ctx.Metadata["baseDelay"] = baseDelay
			return nil
		},
	})

	// After: implement retry logic
	aspect.MustAddAdvice(funcName, aspect.Advice{
		Type:     aspect.After,
		Priority: 100,
		Handler: func(ctx *aspect.Context) error {
			if ctx.Error == nil {
				return nil // Success, no retry needed
			}

			attempt := ctx.Metadata["attempt"].(int)
			maxRetries := ctx.Metadata["maxRetries"].(int)
			baseDelay := ctx.Metadata["baseDelay"].(time.Duration)

			if attempt < maxRetries {
				// Calculate exponential backoff
				delay := time.Duration(math.Pow(2, float64(attempt))) * baseDelay
				log.Printf("[RETRY] %s attempt %d/%d failed, retrying in %v...",
					ctx.FunctionName, attempt+1, maxRetries, delay)

				time.Sleep(delay)
				ctx.Metadata["attempt"] = attempt + 1
			} else {
				log.Printf("[RETRY] %s exhausted all %d retries", ctx.FunctionName, maxRetries)
			}

			return nil
		},
	})
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

// -------------------------------------------- Wrapped Functions --------------------------------------------

var (
	SendEmail      = aspect.Wrap2E("SendEmail", sendEmailImpl)
	ProcessPayment = aspect.Wrap2RE("ProcessPayment", processPaymentImpl)
)

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
		fmt.Printf("✅ Email sent successfully after %d attempts in %v\n", emailAttempts, duration)
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
		fmt.Printf("✅ Payment processed: %s (took %v with %d attempts)\n",
			txnID, duration, paymentAttempts)
	}
}

func example3_ExhaustedRetries() {
	fmt.Println("\n========== Example 3: Exhausted Retries ==========\n")

	// Reset circuit for this example
	aspect.Clear()
	setupAOP()

	// Setup function that always fails
	aspect.MustRegister("FailingOperation")
	setupRetry("FailingOperation", 3, 50*time.Millisecond)

	var FailingOperation = aspect.Wrap1E("FailingOperation", func(data string) error {
		log.Printf("[OPERATION] Executing with data: %s", data)
		return errors.New("permanent failure")
	})

	start := time.Now()
	err := FailingOperation("test_data")
	duration := time.Since(start)

	if err != nil {
		fmt.Printf("❌ Operation failed after all retries: %v (took %v)\n", err, duration)
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
