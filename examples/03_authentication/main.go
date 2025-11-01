// Package main - authentication demonstrates Before advice for auth/authz checks
package main

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/seyedali-dev/gosaidsno/aspect"
)

// -------------------------------------------- Auth System --------------------------------------------

type Session struct {
	UserID    string
	Role      string
	ExpiresAt time.Time
}

var sessions = map[string]*Session{
	"valid_token":   {UserID: "user_123", Role: "admin", ExpiresAt: time.Now().Add(1 * time.Hour)},
	"user_token":    {UserID: "user_456", Role: "user", ExpiresAt: time.Now().Add(1 * time.Hour)},
	"expired_token": {UserID: "user_789", Role: "user", ExpiresAt: time.Now().Add(-1 * time.Hour)},
}

func validateToken(token string) (*Session, error) {
	session, ok := sessions[token]
	if !ok {
		return nil, errors.New("invalid token")
	}
	if time.Now().After(session.ExpiresAt) {
		return nil, errors.New("token expired")
	}
	return session, nil
}

func hasRole(userID, requiredRole string) bool {
	for _, session := range sessions {
		if session.UserID == userID {
			return session.Role == requiredRole || session.Role == "admin"
		}
	}
	return false
}

// -------------------------------------------- Setup --------------------------------------------

func setupAOP() {
	log.Println("=== Setting up Authentication AOP ===")

	aspect.MustRegister("GetUserData")
	aspect.MustRegister("DeleteUser")
	aspect.MustRegister("UpdateSettings")

	// Authentication check (Before advice, priority 100)
	for _, fn := range []string{"GetUserData", "DeleteUser", "UpdateSettings"} {
		aspect.MustAddAdvice(fn, aspect.Advice{
			Type:     aspect.Before,
			Priority: 100, // Highest - run first
			Handler: func(ctx *aspect.Context) error {
				token := ctx.Args[0].(string)

				session, err := validateToken(token)
				if err != nil {
					log.Printf("[AUTH FAILED] %s - %v", ctx.FunctionName, err)
					return fmt.Errorf("authentication failed: %w", err)
				}

				// Store authenticated user in metadata
				ctx.Metadata["userID"] = session.UserID
				ctx.Metadata["role"] = session.Role
				log.Printf("[AUTH SUCCESS] %s - user: %s, role: %s", ctx.FunctionName, session.UserID, session.Role)
				return nil
			},
		})
	}

	// Authorization check for DeleteUser (requires admin role)
	aspect.MustAddAdvice("DeleteUser", aspect.Advice{
		Type:     aspect.Before,
		Priority: 90, // After authentication
		Handler: func(ctx *aspect.Context) error {
			role := ctx.Metadata["role"].(string)

			if role != "admin" {
				userID := ctx.Metadata["userID"].(string)
				log.Printf("[AUTHZ FAILED] DeleteUser - user %s does not have admin role", userID)
				return errors.New("permission denied: admin role required")
			}

			log.Printf("[AUTHZ SUCCESS] DeleteUser - admin access granted")
			return nil
		},
	})

	// Audit logging (After advice)
	for _, fn := range []string{"DeleteUser", "UpdateSettings"} {
		aspect.MustAddAdvice(fn, aspect.Advice{
			Type:     aspect.After,
			Priority: 100,
			Handler: func(ctx *aspect.Context) error {
				userID, _ := ctx.Metadata["userID"].(string)
				status := "SUCCESS"
				if ctx.Error != nil {
					status = "FAILED"
				}

				log.Printf("[AUDIT] Function: %s, User: %s, Status: %s, Args: %v",
					ctx.FunctionName, userID, status, ctx.Args[1:])
				return nil
			},
		})
	}

	log.Println("=== AOP Setup Complete ===\n")
}

// -------------------------------------------- Business Logic --------------------------------------------

func getUserDataImpl(token, userID string) (map[string]string, error) {
	log.Printf("[EXEC] Fetching data for user %s", userID)
	return map[string]string{
		"id":    userID,
		"name":  "John Doe",
		"email": "john@example.com",
	}, nil
}

func deleteUserImpl(token, userID string) error {
	log.Printf("[EXEC] Deleting user %s from database", userID)
	time.Sleep(50 * time.Millisecond)
	return nil
}

func updateSettingsImpl(token string, settings map[string]interface{}) error {
	log.Printf("[EXEC] Updating settings: %v", settings)
	time.Sleep(30 * time.Millisecond)
	return nil
}

// -------------------------------------------- Wrapped Functions --------------------------------------------

var (
	GetUserData    = aspect.Wrap2RE("GetUserData", getUserDataImpl)
	DeleteUser     = aspect.Wrap2E("DeleteUser", deleteUserImpl)
	UpdateSettings = aspect.Wrap2E("UpdateSettings", updateSettingsImpl)
)

// -------------------------------------------- Examples --------------------------------------------

func example1_SuccessfulAuth() {
	fmt.Println("\n========== Example 1: Successful Authentication ==========\n")

	data, err := GetUserData("valid_token", "user_123")
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		return
	}

	fmt.Printf("✅ Got user data: %v\n", data)
}

func example2_InvalidToken() {
	fmt.Println("\n========== Example 2: Invalid Token ==========\n")

	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("❌ Request rejected: %v\n", r)
		}
	}()

	_, _ = GetUserData("invalid_token", "user_123")
}

func example3_ExpiredToken() {
	fmt.Println("\n========== Example 3: Expired Token ==========\n")

	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("❌ Request rejected: %v\n", r)
		}
	}()

	_, _ = GetUserData("expired_token", "user_789")
}

func example4_AuthorizationSuccess() {
	fmt.Println("\n========== Example 4: Authorization Success (Admin) ==========\n")

	err := DeleteUser("valid_token", "user_999")
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		return
	}

	fmt.Println("✅ User deleted successfully")
}

func example5_AuthorizationFailure() {
	fmt.Println("\n========== Example 5: Authorization Failure (Non-Admin) ==========\n")

	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("❌ Request rejected: %v\n", r)
		}
	}()

	_ = DeleteUser("user_token", "user_999")
}

func example6_AuditLogging() {
	fmt.Println("\n========== Example 6: Audit Logging ==========\n")

	settings := map[string]interface{}{
		"theme":         "dark",
		"notifications": true,
	}

	err := UpdateSettings("valid_token", settings)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		return
	}

	fmt.Println("✅ Settings updated (check audit log above)")
}

// -------------------------------------------- Main --------------------------------------------

func main() {
	setupAOP()

	example1_SuccessfulAuth()
	example2_InvalidToken()
	example3_ExpiredToken()
	example4_AuthorizationSuccess()
	example5_AuthorizationFailure()
	example6_AuditLogging()

	fmt.Println("\n========== Authentication Examples Complete ==========")
}
