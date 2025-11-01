// Package main - basic_usage demonstrates core AOP features with real-world scenarios
package main

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/seyedali-dev/gosaidsno/aspect"
)

// -------------------------------------------- Domain Models --------------------------------------------

type User struct {
	ID       string
	Username string
	Email    string
}

type Order struct {
	ID     string
	UserID string
	Amount float64
}

// -------------------------------------------- Setup --------------------------------------------

func setupAOP() {
	log.Println("=== Setting up AOP ===")

	// Register all functions
	aspect.MustRegister("GetUser")
	aspect.MustRegister("CreateOrder")
	aspect.MustRegister("ValidateUser")
	aspect.MustRegister("SendNotification")

	setupLogging()
	setupTiming()
	setupValidation()
	setupPanicRecovery()

	log.Println("=== AOP Setup Complete ===")
	log.Println()
}

func setupLogging() {
	for _, fn := range []string{"GetUser", "CreateOrder", "SendNotification"} {
		aspect.MustAddAdvice(fn, aspect.Advice{
			Type:     aspect.Before,
			Priority: 100,
			Handler: func(ctx *aspect.Context) error {
				log.Println("[BEFORE]")
				log.Printf("[LOG] Starting %s", ctx.FunctionName)
				return nil
			},
		})

		aspect.MustAddAdvice(fn, aspect.Advice{
			Type:     aspect.After,
			Priority: 100,
			Handler: func(ctx *aspect.Context) error {
				log.Println("[AFTER]")
				status := "SUCCESS"
				if ctx.Error != nil {
					status = "FAILED"
				}
				log.Printf("[LOG] Completed %s - %s", ctx.FunctionName, status)
				return nil
			},
		})
	}
}

func setupTiming() {
	for _, fn := range []string{"GetUser", "CreateOrder"} {
		aspect.MustAddAdvice(fn, aspect.Advice{
			Type:     aspect.Before,
			Priority: 90,
			Handler: func(ctx *aspect.Context) error {
				log.Println("[BEFORE]")
				ctx.Metadata["start"] = time.Now()
				return nil
			},
		})

		aspect.MustAddAdvice(fn, aspect.Advice{
			Type:     aspect.After,
			Priority: 90,
			Handler: func(ctx *aspect.Context) error {
				log.Println("[AFTER]")
				start, ok := ctx.Metadata["start"].(time.Time)
				if !ok {
					return nil // Skip if timing not initialized
				}
				duration := time.Since(start)
				log.Printf("[PERF] %s took %v", ctx.FunctionName, duration)
				return nil
			},
		})
	}
}

func setupValidation() {
	aspect.MustAddAdvice("CreateOrder", aspect.Advice{
		Type:     aspect.Before,
		Priority: 110, // Higher priority, runs first
		Handler: func(ctx *aspect.Context) error {
			log.Println("[BEFORE]")
			userID := ctx.Args[0].(string)
			amount := ctx.Args[1].(float64)

			if userID == "" {
				return errors.New("userID cannot be empty")
			}
			if amount <= 0 {
				return errors.New("amount must be positive")
			}
			log.Printf("[VALIDATE] Order validation passed")
			return nil
		},
	})
}

func setupPanicRecovery() {
	for _, fn := range aspect.ListRegistered() {
		aspect.MustAddAdvice(fn, aspect.Advice{
			Type:     aspect.AfterThrowing,
			Priority: 100,
			Handler: func(ctx *aspect.Context) error {
				log.Println("[AFTER_THROWING]")
				log.Printf("[PANIC RECOVERY] Function %s panicked: %v", ctx.FunctionName, ctx.PanicValue)
				// In production: send alert, log to monitoring system
				return nil
			},
		})
	}
}

// -------------------------------------------- Business Logic (Unwrapped) --------------------------------------------

func getUserImpl(id string) (*User, error) {
	// Simulate database query
	time.Sleep(50 * time.Millisecond)

	if id == "" {
		return nil, errors.New("user ID is required")
	}

	return &User{
		ID:       id,
		Username: "john_doe",
		Email:    "john@example.com",
	}, nil
}

func createOrderImpl(userID string, amount float64) (*Order, error) {
	// Simulate order creation
	time.Sleep(100 * time.Millisecond)

	order := &Order{
		ID:     fmt.Sprintf("order_%d", time.Now().Unix()),
		UserID: userID,
		Amount: amount,
	}

	return order, nil
}

func validateUserImpl(user *User) error {
	if user.Email == "invalid@example.com" {
		return errors.New("invalid email domain")
	}
	return nil
}

func sendNotificationImpl(userID, message string) {
	// Simulate notification sending
	time.Sleep(30 * time.Millisecond)
	log.Printf("[EMAIL] Sent to user %s: %s", userID, message)
}

// -------------------------------------------- Wrapped Functions --------------------------------------------

var (
	GetUser          = aspect.Wrap1RE("GetUser", getUserImpl)
	CreateOrder      = aspect.Wrap2RE("CreateOrder", createOrderImpl)
	ValidateUser     = aspect.Wrap1E("ValidateUser", validateUserImpl)
	SendNotification = aspect.Wrap2("SendNotification", sendNotificationImpl)
)

// -------------------------------------------- Examples --------------------------------------------

func example1_BasicLoggingAndTiming() {
	fmt.Println("\n========== Example 1: Basic Logging & Timing ==========")

	// Normal successful operation
	user, err := GetUser("user_123")
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	fmt.Printf("\nResult: Got user %s (%s)\n", user.Username, user.Email)
}

func example2_Validation() {
	fmt.Println("\n========== Example 2: Pre-execution Validation ==========")

	// This will fail validation
	log.Println("\n--- Attempting to create order with invalid data ---")
	func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("\n❌ Order creation rejected by validation: %v\n", r)
			}
		}()
		_, _ = CreateOrder("", -100)
	}()

	// This will succeed
	log.Println("\n--- Creating valid order ---")
	order, err := CreateOrder("user_123", 99.99)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	fmt.Printf("\n✅ Order created: %s for $%.2f\n", order.ID, order.Amount)
}

func example3_ErrorHandling() {
	fmt.Println("\n========== Example 3: Error Handling ==========")

	// Success case
	log.Println("\n--- Validating valid user ---")
	validUser := &User{ID: "1", Username: "john", Email: "john@example.com"}
	err := ValidateUser(validUser)
	if err == nil {
		fmt.Println("✅ User validation passed")
	}

	// Error case
	log.Println("\n--- Validating invalid user ---")
	invalidUser := &User{ID: "2", Username: "bad", Email: "invalid@example.com"}
	err = ValidateUser(invalidUser)
	if err != nil {
		fmt.Printf("❌ User validation failed: %v\n", err)
	}
}

func example4_AfterReturning() {
	fmt.Println("\n========== Example 4: AfterReturning (Success-only logic) ==========")

	// Add AfterReturning advice
	aspect.MustAddAdvice("CreateOrder", aspect.Advice{
		Type:     aspect.AfterReturning,
		Priority: 50,
		Handler: func(ctx *aspect.Context) error {
			log.Println("[AFTER_RETURNING]")
			log.Printf("[SUCCESS HOOK] Order created successfully, sending confirmation...")
			order := ctx.Results[0].(*Order)
			SendNotification(order.UserID, fmt.Sprintf("Order %s confirmed!", order.ID))
			return nil
		},
	})

	order, _ := CreateOrder("user_456", 149.99)
	fmt.Printf("\n✅ Order %s completed with confirmation sent\n", order.ID)
}

// -------------------------------------------- Main --------------------------------------------

func main() {
	// Setup AOP once at startup
	setupAOP()

	// Run examples
	example1_BasicLoggingAndTiming()
	example2_Validation()
	example3_ErrorHandling()
	example4_AfterReturning()

	fmt.Println("\n========== All Examples Complete ==========")
}
