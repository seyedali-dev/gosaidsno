// Package aspect - advice_test validates advice execution and ordering
package aspect

import (
	"errors"
	"testing"
)

// -------------------------------------------- Tests --------------------------------------------

func TestAdviceChain_ExecutionOrder(t *testing.T) {
	chain := NewAdviceChain()
	var order []string

	// Add advice with different priorities
	chain.Add(Advice{
		Type:     Before,
		Priority: 10,
		Handler: func(ctx *Context) error {
			order = append(order, "before-10")
			return nil
		},
	})
	chain.Add(Advice{
		Type:     Before,
		Priority: 20,
		Handler: func(ctx *Context) error {
			order = append(order, "before-20")
			return nil
		},
	})

	ctx := NewContext("test")
	_ = chain.ExecuteBefore(ctx)

	// Higher priority should execute first
	if len(order) != 2 {
		t.Fatalf("expected 2 advice executions, got %d", len(order))
	}
	if order[0] != "before-20" {
		t.Errorf("expected first execution to be 'before-20', got '%s'", order[0])
	}
	if order[1] != "before-10" {
		t.Errorf("expected second execution to be 'before-10', got '%s'", order[1])
	}
}

func TestAdviceChain_ErrorPropagation(t *testing.T) {
	chain := NewAdviceChain()

	chain.Add(Advice{
		Type:     Before,
		Priority: 100,
		Handler: func(ctx *Context) error {
			return errors.New("advice failed")
		},
	})

	ctx := NewContext("test")
	err := chain.ExecuteBefore(ctx)

	if err == nil {
		t.Fatal("expected error from advice, got nil")
	}
}

func TestContext_SetAndGetResult(t *testing.T) {
	ctx := NewContext("test")

	ctx.SetResult(0, "hello")
	ctx.SetResult(1, 42)

	result0 := ctx.GetResult(0)
	if result0 != "hello" {
		t.Errorf("expected 'hello', got %v", result0)
	}

	result1 := ctx.GetResult(1)
	if result1 != 42 {
		t.Errorf("expected 42, got %v", result1)
	}

	// Out of bounds should return nil
	result2 := ctx.GetResult(999)
	if result2 != nil {
		t.Errorf("expected nil for out of bounds, got %v", result2)
	}
}

func TestContext_HasPanic(t *testing.T) {
	ctx := NewContext("test")

	if ctx.HasPanic() {
		t.Error("expected HasPanic to be false initially")
	}

	ctx.PanicValue = "something went wrong"

	if !ctx.HasPanic() {
		t.Error("expected HasPanic to be true after setting PanicValue")
	}
}

func TestWrap1R_BasicExecution(t *testing.T) {
	MustRegister("TestFunc")

	var beforeCalled, afterCalled bool

	MustAddAdvice("TestFunc", Advice{
		Type:     Before,
		Priority: 100,
		Handler: func(ctx *Context) error {
			beforeCalled = true
			if len(ctx.Args) != 1 {
				t.Errorf("expected 1 arg, got %d", len(ctx.Args))
			}
			return nil
		},
	})

	MustAddAdvice("TestFunc", Advice{
		Type:     After,
		Priority: 100,
		Handler: func(ctx *Context) error {
			afterCalled = true
			return nil
		},
	})

	fn := func(x int) int {
		return x * 2
	}

	wrapped := Wrap1R("TestFunc", fn)
	result := wrapped(5)

	if result != 10 {
		t.Errorf("expected result 10, got %d", result)
	}

	if !beforeCalled {
		t.Error("Before advice was not called")
	}

	if !afterCalled {
		t.Error("After advice was not called")
	}

	Unregister("TestFunc")
}

func TestWrap1RE_ErrorCapture(t *testing.T) {
	MustRegister("TestFuncError")

	var capturedError error

	MustAddAdvice("TestFuncError", Advice{
		Type:     After,
		Priority: 100,
		Handler: func(ctx *Context) error {
			capturedError = ctx.Error
			return nil
		},
	})

	fn := func(x int) (int, error) {
		if x == 0 {
			return 0, errors.New("cannot be zero")
		}
		return x * 2, nil
	}

	wrapped := Wrap1RE("TestFuncError", fn)
	_, err := wrapped(0)

	if err == nil {
		t.Fatal("expected error from function, got nil")
	}

	if capturedError == nil {
		t.Fatal("expected error to be captured in context")
	}

	if capturedError.Error() != "cannot be zero" {
		t.Errorf("unexpected error message: %s", capturedError.Error())
	}

	Unregister("TestFuncError")
}

func TestAfterReturning_OnlyOnSuccess(t *testing.T) {
	MustRegister("TestAfterReturning")

	var afterReturningCalled bool

	MustAddAdvice("TestAfterReturning", Advice{
		Type:     AfterReturning,
		Priority: 100,
		Handler: func(ctx *Context) error {
			afterReturningCalled = true
			return nil
		},
	})

	// Test success case
	successFn := func(x int) (int, error) {
		return x * 2, nil
	}

	wrapped := Wrap1RE("TestAfterReturning", successFn)
	_, _ = wrapped(5)

	if !afterReturningCalled {
		t.Error("AfterReturning should be called on success")
	}

	// Reset flag
	afterReturningCalled = false

	// Test error case
	errorFn := func(x int) (int, error) {
		return 0, errors.New("error")
	}

	wrapped2 := Wrap1RE("TestAfterReturning", errorFn)
	_, _ = wrapped2(5)

	if afterReturningCalled {
		t.Error("AfterReturning should NOT be called on error")
	}

	Unregister("TestAfterReturning")
}

func TestAfterThrowing_OnPanic(t *testing.T) {
	MustRegister("TestAfterThrowing")

	var afterThrowingCalled bool
	var capturedPanic interface{}

	MustAddAdvice("TestAfterThrowing", Advice{
		Type:     AfterThrowing,
		Priority: 100,
		Handler: func(ctx *Context) error {
			afterThrowingCalled = true
			capturedPanic = ctx.PanicValue
			return nil
		},
	})

	panicFn := func(x int) {
		if x == 0 {
			panic("zero panic")
		}
	}

	wrapped := Wrap1("TestAfterThrowing", panicFn)

	// Catch the re-panic
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic to be re-thrown")
		}
	}()

	wrapped(0)

	if !afterThrowingCalled {
		t.Error("AfterThrowing should be called on panic")
	}

	if capturedPanic == nil {
		t.Error("panic value should be captured")
	}

	Unregister("TestAfterThrowing")
}

func TestAround_SkipExecution(t *testing.T) {
	MustRegister("TestAround")
	MustAddAdvice("TestAround", Advice{
		Type:     Around,
		Priority: 100,
		Handler: func(ctx *Context) error {
			ctx.Skipped = true
			ctx.SetResult(0, 999) // Return custom value
			return nil
		},
	})

	var targetCalled bool
	fn := func(x int) int {
		targetCalled = true
		return x * 2
	}

	wrapped := Wrap1R("TestAround", fn)
	result := wrapped(5)

	if targetCalled {
		t.Error("target function should not be called when Skipped=true")
	}

	if result != 999 {
		t.Errorf("expected result from Around advice (999), got %d", result)
	}

	Unregister("TestAround")
}
