// Package aspect - registry_test validates registry functionality
package aspect

import (
	"fmt"
	"sync"
	"testing"
)

// -------------------------------------------- Tests --------------------------------------------

func TestRegistry_Register(t *testing.T) {
	registry := NewRegistry()

	// Test successful registration
	err := registry.Register("TestFunc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Test duplicate registration
	err = registry.Register("TestFunc")
	if err == nil {
		t.Fatal("expected error for duplicate registration")
	}

	// Test empty name
	err = registry.Register("")
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestRegistry_RegisterOrGet(t *testing.T) {
	registry := NewRegistry()

	// First call should register
	chain1 := registry.RegisterOrGet("TestFunc")
	if chain1 == nil {
		t.Fatal("expected non-nil chain")
	}

	// Second call should return same chain
	chain2 := registry.RegisterOrGet("TestFunc")
	if chain2 != chain1 {
		t.Fatal("expected same chain instance")
	}

	// Should panic on empty name
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for empty name")
		}
	}()
	registry.RegisterOrGet("")
}

func TestRegistry_MustRegister(t *testing.T) {
	registry := NewRegistry()

	// Should not panic on valid registration
	registry.MustRegister("TestFunc1")

	// Should panic on duplicate
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on duplicate registration")
		}
	}()
	registry.MustRegister("TestFunc1")
}

func TestRegistry_AddAdvice(t *testing.T) {
	registry := NewRegistry()

	// Test adding advice to non-existent function
	err := registry.AddAdvice("NonExistent", Advice{
		Type:     Before,
		Priority: 100,
		Handler:  func(ctx *Context) error { return nil },
	})
	if err == nil {
		t.Fatal("expected error for non-existent function")
	}

	// Register function and add advice
	_ = registry.Register("TestFunc")
	err = registry.AddAdvice("TestFunc", Advice{
		Type:     Before,
		Priority: 100,
		Handler:  func(ctx *Context) error { return nil },
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Test empty function name
	err = registry.AddAdvice("", Advice{
		Type:     Before,
		Priority: 100,
		Handler:  func(ctx *Context) error { return nil },
	})
	if err == nil {
		t.Fatal("expected error for empty function name")
	}
}

func TestRegistry_GetAdviceChain(t *testing.T) {
	registry := NewRegistry()

	// Test getting non-existent chain
	_, err := registry.GetAdviceChain("NonExistent")
	if err == nil {
		t.Fatal("expected error for non-existent function")
	}

	// Register and get chain
	_ = registry.Register("TestFunc")
	chain, err := registry.GetAdviceChain("TestFunc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if chain == nil {
		t.Fatal("expected non-nil chain")
	}

	// Test empty name
	_, err = registry.GetAdviceChain("")
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestRegistry_IsRegistered(t *testing.T) {
	registry := NewRegistry()

	if registry.IsRegistered("TestFunc") {
		t.Fatal("expected false for non-existent function")
	}

	_ = registry.Register("TestFunc")

	if !registry.IsRegistered("TestFunc") {
		t.Fatal("expected true for registered function")
	}
}

func TestRegistry_Unregister(t *testing.T) {
	registry := NewRegistry()

	_ = registry.Register("TestFunc")

	if !registry.IsRegistered("TestFunc") {
		t.Fatal("function should be registered")
	}

	registry.Unregister("TestFunc")

	if registry.IsRegistered("TestFunc") {
		t.Fatal("function should be unregistered")
	}

	// Unregistering non-existent function should not panic
	registry.Unregister("NonExistent")
}

func TestRegistry_ListRegistered(t *testing.T) {
	registry := NewRegistry()

	// Empty registry
	names := registry.ListRegistered()
	if len(names) != 0 {
		t.Fatalf("expected 0 names, got %d", len(names))
	}

	// Register multiple functions
	_ = registry.Register("Func1")
	_ = registry.Register("Func2")
	_ = registry.Register("Func3")

	names = registry.ListRegistered()
	if len(names) != 3 {
		t.Fatalf("expected 3 names, got %d", len(names))
	}

	// Check all names are present
	nameMap := make(map[string]bool)
	for _, name := range names {
		nameMap[name] = true
	}

	if !nameMap["Func1"] || !nameMap["Func2"] || !nameMap["Func3"] {
		t.Fatal("not all registered functions are in the list")
	}
}

func TestRegistry_Clear(t *testing.T) {
	registry := NewRegistry()

	_ = registry.Register("Func1")
	_ = registry.Register("Func2")

	if registry.Count() != 2 {
		t.Fatalf("expected 2 functions, got %d", registry.Count())
	}

	registry.Clear()

	if registry.Count() != 0 {
		t.Fatalf("expected 0 functions after clear, got %d", registry.Count())
	}
}

func TestRegistry_Count(t *testing.T) {
	registry := NewRegistry()

	if registry.Count() != 0 {
		t.Fatalf("expected 0, got %d", registry.Count())
	}

	_ = registry.Register("Func1")
	_ = registry.Register("Func2")

	if registry.Count() != 2 {
		t.Fatalf("expected 2, got %d", registry.Count())
	}
}

func TestRegistry_GetAdviceCount(t *testing.T) {
	registry := NewRegistry()

	// Non-existent function
	count := registry.GetAdviceCount("NonExistent")
	if count != 0 {
		t.Fatalf("expected 0, got %d", count)
	}

	// Register and add advice
	_ = registry.Register("TestFunc")
	_ = registry.AddAdvice("TestFunc", Advice{
		Type:     Before,
		Priority: 100,
		Handler:  func(ctx *Context) error { return nil },
	})
	_ = registry.AddAdvice("TestFunc", Advice{
		Type:     After,
		Priority: 100,
		Handler:  func(ctx *Context) error { return nil },
	})

	count = registry.GetAdviceCount("TestFunc")
	if count != 2 {
		t.Fatalf("expected 2, got %d", count)
	}
}

func TestRegistry_ConcurrentAccess(t *testing.T) {
	registry := NewRegistry()

	var wg sync.WaitGroup
	numGoroutines := 100

	// Concurrent registrations
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			_ = registry.Register(fmt.Sprintf("Func%d", n))
		}(i)
	}

	wg.Wait()

	// Concurrent reads
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = registry.ListRegistered()
			_ = registry.Count()
		}()
	}

	wg.Wait()
}

func TestGlobalRegistry_Functions(t *testing.T) {
	// Clear global registry before test
	Clear()

	// Test global Register
	err := Register("GlobalFunc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Test global IsRegistered
	if !IsRegistered("GlobalFunc") {
		t.Fatal("function should be registered in global registry")
	}

	// Test global RegisterOrGet
	chain1 := RegisterOrGet("GlobalFunc2")
	if chain1 == nil {
		t.Fatal("expected non-nil chain")
	}
	chain2 := RegisterOrGet("GlobalFunc2")
	if chain1 != chain2 {
		t.Fatal("expected same chain from RegisterOrGet")
	}

	// Test global AddAdvice
	err = AddAdvice("GlobalFunc", Advice{
		Type:     Before,
		Priority: 100,
		Handler:  func(ctx *Context) error { return nil },
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Test global GetAdviceChain
	chain, err := GetAdviceChain("GlobalFunc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if chain == nil {
		t.Fatal("expected non-nil chain")
	}

	// Test global Count
	if Count() < 2 {
		t.Fatalf("expected at least 2, got %d", Count())
	}

	// Test global GetAdviceCount
	if GetAdviceCount("GlobalFunc") != 1 {
		t.Fatalf("expected 1 advice, got %d", GetAdviceCount("GlobalFunc"))
	}

	// Test global ListRegistered
	names := ListRegistered()
	if len(names) < 2 {
		t.Fatalf("expected at least 2 names, got %d: %v", len(names), names)
	}

	// Test global Unregister
	Unregister("GlobalFunc")
	if IsRegistered("GlobalFunc") {
		t.Fatal("function should be unregistered")
	}

	// Clean up
	Clear()
}

func TestSetGlobalRegistry(t *testing.T) {
	// Save original
	original := GetGlobalRegistry()

	// Create and set new registry
	newRegistry := NewRegistry()
	SetGlobalRegistry(newRegistry)

	// Verify it's the new one
	if GetGlobalRegistry() != newRegistry {
		t.Fatal("global registry was not set correctly")
	}

	// Restore original
	SetGlobalRegistry(original)
}
