# Setup Guide - Organizing AOP in Your Project

## Project Structure

```
myproject/
├── main.go
├── aop/
│   └── setup.go          # One-time AOP configuration
├── services/
│   ├── user_service.go   # Business logic with wrapped functions
│   └── payment_service.go
└── go.mod
```

## Pattern 1: Centralized Setup (Recommended)

**aop/setup.go** - Configure all advice once at startup:

```go
package aop

import (
    "log"
    "time"
    "github.com/seyedali-dev/gosaidsno/aspect"
)

// InitAOP registers all functions and advice at application startup.
func InitAOP() {
    setupLogging()
    setupTiming()
    setupErrorHandling()
    log.Println("[AOP] Initialization complete")
}

func setupLogging() {
    // Register functions that need logging
    aspect.MustRegister("UserService.GetUser")
    aspect.MustRegister("UserService.CreateUser")
    aspect.MustRegister("PaymentService.ProcessPayment")
    
    // Add Before advice for all
    for _, funcName := range []string{
        "UserService.GetUser",
        "UserService.CreateUser", 
        "PaymentService.ProcessPayment",
    } {
        aspect.MustAddAdvice(funcName, aspect.Advice{
            Type: aspect.Before,
            Priority: 100,
            Handler: func(ctx *aspect.Context) error {
                log.Printf("[BEFORE] %s called with args: %v", ctx.FunctionName, ctx.Args)
                return nil
            },
        })
    }
}

func setupTiming() {
    // Functions to time
    funcs := []string{"UserService.GetUser", "PaymentService.ProcessPayment"}
    
    for _, funcName := range funcs {
        // Before: capture start time
        aspect.MustAddAdvice(funcName, aspect.Advice{
            Type: aspect.Before,
            Priority: 90,
            Handler: func(ctx *aspect.Context) error {
                ctx.Metadata["startTime"] = time.Now()
                return nil
            },
        })
        
        // After: log duration
        aspect.MustAddAdvice(funcName, aspect.Advice{
            Type: aspect.After,
            Priority: 90,
            Handler: func(ctx *aspect.Context) error {
                start := ctx.Metadata["startTime"].(time.Time)
                log.Printf("[TIMING] %s took %v", ctx.FunctionName, time.Since(start))
                return nil
            },
        })
    }
}

func setupErrorHandling() {
    // Add AfterThrowing for all registered functions
    for _, funcName := range aspect.ListRegistered() {
        aspect.MustAddAdvice(funcName, aspect.Advice{
            Type: aspect.AfterThrowing,
            Priority: 100,
            Handler: func(ctx *aspect.Context) error {
                log.Printf("[PANIC] %s panicked: %v", ctx.FunctionName, ctx.PanicValue)
                // Send alert, record metric, etc.
                return nil
            },
        })
    }
}
```

**main.go** - Initialize once at startup:

```go
package main

import (
    "myproject/aop"
    "myproject/services"
)

func main() {
    // Initialize AOP once at startup
    aop.InitAOP()
    
    // Use services normally
    userService := services.NewUserService()
    user, _ := userService.GetUser("user_123")
    // ... rest of application
}
```

**services/user_service.go** - Wrap functions, use normally:

```go
package services

import "github.com/seyedali-dev/gosaidsno/aspect"

type UserService struct {
    getUser func(string) (*User, error)
}

func NewUserService() *UserService {
    // Wrap once during initialization
    return &UserService{
        getUser: aspect.Wrap1RE("UserService.GetUser", getUserImpl),
    }
}

func (s *UserService) GetUser(id string) (*User, error) {
    return s.getUser(id)
}

func getUserImpl(id string) (*User, error) {
    // Actual business logic
    return fetchFromDB(id)
}
```

## Pattern 2: Per-Service Setup

Each service configures its own advice:

```go
package services

func init() {
    // Register and configure advice for this service
    aspect.MustRegister("PaymentService.Process")
    aspect.MustAddAdvice("PaymentService.Process", /* ... */)
}
```

## Key Points

1. **Register once** at startup (via `aop.InitAOP()` or service `init()`)
2. **Wrap once** during object construction
3. **Use normally** throughout the application
4. **Advice is global** - applies to all instances