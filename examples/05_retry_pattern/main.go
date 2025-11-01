// Package main - retry_pattern demonstrates wrapper pattern for automatic retries
package main

import (
	"errors"
	"fmt"
	"log"
	"math"
	"time"

	"github.com/seyedali-dev/gosaidsno/aspect"
	"github.com/seyedali-dev/gosaidsno/examples/utils"
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
				log.Printf("   🔄 [RETRY] Attempt %d/%d failed, retrying in %v...", attempt, maxRetries, delay)
				log.Printf("   ⏳ [RETRY] Exponential backoff: 2^%d * %v = %v", attempt-1, baseDelay, delay)
				time.Sleep(delay)
			}

			log.Printf("   🎯 [RETRY] Making attempt %d/%d", attempt+1, maxRetries+1)
			result, err = fn()
			if err == nil {
				if attempt > 0 {
					log.Printf("   ✅ [RETRY] Success on attempt %d/%d", attempt+1, maxRetries+1)
				} else {
					log.Printf("   ✅ [RETRY] Success on first attempt")
				}
				return result, nil
			}

			log.Printf("   ❌ [RETRY] Attempt %d/%d failed: %v", attempt+1, maxRetries+1, err)
		}

		log.Printf("   💥 [RETRY] Exhausted all %d retries", maxRetries)
		log.Printf("   🚨 [RETRY] Final failure after %d attempts", maxRetries+1)
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
				utils.LogBefore(ctx, 100, "TIMING START")
				ctx.Metadata["start"] = time.Now()
				log.Printf("   ⏱️  [TIMING] Started execution timer")
				return nil
			},
		})

		aspect.MustAddAdvice(fn, aspect.Advice{
			Type:     aspect.After,
			Priority: 100,
			Handler: func(ctx *aspect.Context) error {
				utils.LogAfter(ctx, 100, "TIMING END")
				start := ctx.Metadata["start"].(time.Time)
				duration := time.Since(start)
				status := "SUCCESS"
				if ctx.Error != nil {
					status = "FAILED"
				}
				log.Printf("   📊 [MONITOR] %s %s in %v", ctx.FunctionName, status, duration)
				log.Printf("   📈 [METRICS] Execution completed with status: %s", status)
				return nil
			},
		})
	}

	// Add detailed logging for email service
	aspect.MustAddAdvice("SendEmail", aspect.Advice{
		Type:     aspect.Before,
		Priority: 90,
		Handler: func(ctx *aspect.Context) error {
			utils.LogBefore(ctx, 90, "EMAIL LOG")
			to := ctx.Args[0].(string)
			subject := ctx.Args[1].(string)
			log.Printf("   📧 [EMAIL] Preparing to send email to: %s", to)
			log.Printf("   📝 [EMAIL] Subject: %s", subject)
			return nil
		},
	})

	// Add detailed logging for payment service
	aspect.MustAddAdvice("ProcessPayment", aspect.Advice{
		Type:     aspect.Before,
		Priority: 90,
		Handler: func(ctx *aspect.Context) error {
			utils.LogBefore(ctx, 90, "PAYMENT LOG")
			amount := ctx.Args[0].(float64)
			cardToken := ctx.Args[1].(string)
			log.Printf("   💳 [PAYMENT] Processing payment: $%.2f", amount)
			log.Printf("   🔐 [PAYMENT] Card token: %s...", cardToken[:8])
			return nil
		},
	})

	log.Println("=== AOP Setup Complete ===\n")
}

// -------------------------------------------- Business Logic --------------------------------------------

var emailAttempts = 0

func sendEmailImpl(to, subject string) error {
	emailAttempts++
	log.Printf("   📧 [BUSINESS] sendEmailImpl executing - attempt #%d", emailAttempts)
	log.Printf("   📤 [BUSINESS] Connecting to SMTP server...")

	// Simulate transient failures
	if emailAttempts <= 2 {
		log.Printf("   ❌ [BUSINESS] SMTP connection timeout (simulated)")
		return errors.New("SMTP connection timeout")
	}

	log.Printf("   ✅ [BUSINESS] Email sent successfully!")
	log.Printf("   📨 [BUSINESS] Recipient: %s", to)
	log.Printf("   📄 [BUSINESS] Subject: %s", subject)
	return nil
}

var paymentAttempts = 0

func processPaymentImpl(amount float64, cardToken string) (string, error) {
	paymentAttempts++
	log.Printf("   💳 [BUSINESS] processPaymentImpl executing - attempt #%d", paymentAttempts)
	log.Printf("   🔄 [BUSINESS] Processing payment of $%.2f", amount)

	// Simulate various failure scenarios
	if paymentAttempts == 1 {
		log.Printf("   ❌ [BUSINESS] Network timeout (simulated)")
		return "", errors.New("network timeout")
	}
	if paymentAttempts == 2 {
		log.Printf("   ❌ [BUSINESS] Payment gateway unavailable (simulated)")
		return "", errors.New("payment gateway unavailable")
	}

	txnID := fmt.Sprintf("txn_%d", time.Now().Unix())
	log.Printf("   ✅ [BUSINESS] Payment processed successfully!")
	log.Printf("   🧾 [BUSINESS] Transaction ID: %s", txnID)
	log.Printf("   💰 [BUSINESS] Amount: $%.2f", amount)
	return txnID, nil
}

// -------------------------------------------- Wrapped Functions with Retry --------------------------------------------

// SendEmail with retry wrapper
func SendEmail(to, subject string) error {
	log.Printf("🚀 [ENTRY] SendEmail called with retry wrapper")
	log.Printf("   🔧 [RETRY CONFIG] Max retries: 3, Base delay: 100ms")

	// Wrap the base function with retry logic
	retryFn := WithRetry(func() (struct{}, error) {
		err := sendEmailBase(to, subject)
		return struct{}{}, err
	}, 3, 100*time.Millisecond)

	_, err := retryFn()

	if err != nil {
		log.Printf("💥 [EXIT] SendEmail failed after all retries: %v", err)
	} else {
		log.Printf("✅ [EXIT] SendEmail completed successfully")
	}

	return err
}

var sendEmailBase = aspect.Wrap2E("SendEmail", sendEmailImpl)

// ProcessPayment with retry wrapper
func ProcessPayment(amount float64, cardToken string) (string, error) {
	log.Printf("🚀 [ENTRY] ProcessPayment called with retry wrapper")
	log.Printf("   🔧 [RETRY CONFIG] Max retries: 5, Base delay: 200ms")

	retryFn := WithRetry(func() (string, error) {
		return processPaymentBase(amount, cardToken)
	}, 5, 200*time.Millisecond)

	result, err := retryFn()

	if err != nil {
		log.Printf("💥 [EXIT] ProcessPayment failed after all retries: %v", err)
	} else {
		log.Printf("✅ [EXIT] ProcessPayment completed successfully: %s", result)
	}

	return result, err
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
		log.Printf("   💥 [BUSINESS] FailingOperation executing - attempt #%d", failAttempts)
		log.Printf("   ❌ [BUSINESS] Permanent failure (simulated)")
		return "", errors.New("permanent failure")
	}, 3, 50*time.Millisecond)

	log.Printf("🚀 [ENTRY] FailingOperation called (will exhaust retries)")
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
