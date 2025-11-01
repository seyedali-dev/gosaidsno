// Package aspect - wrap provides function wrapping utilities with AOP advice execution
package aspect

import "fmt"

// -------------------------------------------- Public Functions --------------------------------------------

// Wrap0 wraps a function with no arguments and no return values.
func Wrap0(name string, fn func()) func() {
	return func() {
		executeWithAdvice(name, func(ctx *Context) {
			fn()
		})
	}
}

// Wrap0R wraps a function with no arguments and one return value.
func Wrap0R[R any](name string, fn func() R) func() R {
	return func() R {
		var result R
		ctx := executeWithAdvice(name, func(ctx *Context) {
			result = fn()
			ctx.SetResult(0, result)
		})

		// If Around advice set a result and skipped execution, use that result
		if ctx != nil && len(ctx.Results) > 0 && ctx.Results[0] != nil {
			result = ctx.Results[0].(R)
		}

		return result
	}
}

// Wrap0RE wraps a function with no arguments and returns (result, error).
func Wrap0RE[R any](name string, fn func() (R, error)) func() (R, error) {
	return func() (R, error) {
		var result R
		var err error
		ctx := executeWithAdvice(name, func(ctx *Context) {
			result, err = fn()
			ctx.SetResult(0, result)
			ctx.Error = err
		})

		// If Around advice set a result and skipped execution, use that result
		if ctx != nil && len(ctx.Results) > 0 && ctx.Results[0] != nil {
			result = ctx.Results[0].(R)
		}
		if ctx != nil && ctx.Error != nil {
			err = ctx.Error
		}

		return result, err
	}
}

// Wrap1 wraps a function with one argument and no return values.
func Wrap1[A any](name string, fn func(A)) func(A) {
	return func(a A) {
		executeWithAdvice(name, func(ctx *Context) {
			fn(a)
		}, a)
	}
}

// Wrap1R wraps a function with one argument and one return value.
func Wrap1R[A, R any](name string, fn func(A) R) func(A) R {
	return func(a A) R {
		var result R
		ctx := executeWithAdvice(name, func(ctx *Context) {
			result = fn(a)
			ctx.SetResult(0, result)
		}, a)

		// If Around advice set a result and skipped execution, use that result
		if ctx != nil && len(ctx.Results) > 0 && ctx.Results[0] != nil {
			result = ctx.Results[0].(R)
		}

		return result
	}
}

// Wrap1RE wraps a function with one argument and returns (result, error).
func Wrap1RE[A, R any](name string, fn func(A) (R, error)) func(A) (R, error) {
	return func(a A) (R, error) {
		var result R
		var err error
		executeWithAdvice(name, func(ctx *Context) {
			result, err = fn(a)
			ctx.SetResult(0, result)
			ctx.Error = err
		}, a)
		return result, err
	}
}

// Wrap1E wraps a function with one argument and returns error.
func Wrap1E[A any](name string, fn func(A) error) func(A) error {
	return func(a A) error {
		var err error
		executeWithAdvice(name, func(ctx *Context) {
			err = fn(a)
			ctx.Error = err
		}, a)
		return err
	}
}

// Wrap2 wraps a function with two arguments and no return values.
func Wrap2[A, B any](name string, fn func(A, B)) func(A, B) {
	return func(a A, b B) {
		executeWithAdvice(name, func(ctx *Context) {
			fn(a, b)
		}, a, b)
	}
}

// Wrap2R wraps a function with two arguments and one return value.
func Wrap2R[A, B, R any](name string, fn func(A, B) R) func(A, B) R {
	return func(a A, b B) R {
		var result R
		executeWithAdvice(name, func(ctx *Context) {
			result = fn(a, b)
			ctx.SetResult(0, result)
		}, a, b)
		return result
	}
}

// Wrap2RE wraps a function with two arguments and returns (result, error).
func Wrap2RE[A, B, R any](name string, fn func(A, B) (R, error)) func(A, B) (R, error) {
	return func(a A, b B) (R, error) {
		var result R
		var err error
		executeWithAdvice(name, func(ctx *Context) {
			result, err = fn(a, b)
			ctx.SetResult(0, result)
			ctx.Error = err
		}, a, b)
		return result, err
	}
}

// Wrap2E wraps a function with two arguments and returns error.
func Wrap2E[A, B any](name string, fn func(A, B) error) func(A, B) error {
	return func(a A, b B) error {
		var err error
		executeWithAdvice(name, func(ctx *Context) {
			err = fn(a, b)
			ctx.Error = err
		}, a, b)
		return err
	}
}

// Wrap3RE wraps a function with three arguments and returns (result, error).
func Wrap3RE[A, B, C, R any](name string, fn func(A, B, C) (R, error)) func(A, B, C) (R, error) {
	return func(a A, b B, c C) (R, error) {
		var result R
		var err error
		executeWithAdvice(name, func(ctx *Context) {
			result, err = fn(a, b, c)
			ctx.SetResult(0, result)
			ctx.Error = err
		}, a, b, c)
		return result, err
	}
}

// -------------------------------------------- Private Helper Functions --------------------------------------------

// executeWithAdvice executes a function with full advice chain support and returns the context.
func executeWithAdvice(functionName string, targetFn func(*Context), args ...any) *Context {
	// Get advice chain from registry
	chain, err := GetAdviceChain(functionName)
	if err != nil {
		// No advice registered, just execute target function
		ctx := NewContext(functionName, args...)
		targetFn(ctx)
		return ctx
	}

	// Create execution context
	ctx := NewContext(functionName, args...)

	// Defer After advice (always runs)
	defer func() {
		_ = chain.ExecuteAfter(ctx)
	}()

	// Defer panic recovery and AfterThrowing advice
	defer func() {
		if r := recover(); r != nil {
			ctx.PanicValue = r
			_ = chain.ExecuteAfterThrowing(ctx)

			// Re-panic to maintain panic semantics
			panic(r)
		}
	}()

	// Execute Before advice
	if err := chain.ExecuteBefore(ctx); err != nil {
		panic(fmt.Errorf("before advice failed: %w", err))
	}

	// Execute Around advice (if any)
	if chain.HasAround() {
		if err := chain.ExecuteAround(ctx); err != nil {
			panic(fmt.Errorf("around advice failed: %w", err))
		}
		// If Around advice sets Skipped, don't execute target function
		if ctx.Skipped {
			// Execute AfterReturning if no error (Around advice might have set result)
			if ctx.Error == nil && !ctx.HasPanic() {
				_ = chain.ExecuteAfterReturning(ctx)
			}
			return ctx
		}
	}

	// Execute target function
	targetFn(ctx)

	// Execute AfterReturning advice (only if no error and no panic)
	if ctx.Error == nil && !ctx.HasPanic() {
		_ = chain.ExecuteAfterReturning(ctx)
	}

	return ctx
}
