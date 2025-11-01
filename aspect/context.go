// Package aspect - context provides execution context for aspect-oriented advice
package aspect

import "fmt"

// -------------------------------------------- Types --------------------------------------------

// Context holds the execution state for a single function invocation.
// It captures arguments, return values, errors, and panic information.
type Context struct {
	FunctionName string         // FunctionName is the registered name of the wrapped function.
	Args         []any          // Args contains the function arguments (caller must cast to correct types).
	Results      []any          // Results contains the function return values (populated after execution).
	Error        error          // Error holds any error returned by the function.
	PanicValue   any            // PanicValue holds the recovered panic value if a panic occurred.
	Metadata     map[string]any // Metadata allows storing custom key-value pairs for advice communication.
	Skipped      bool           // Skipped indicates if the target function execution should be skipped (set by Around advice).
}

// NewContext creates a new execution context for the given function.
func NewContext(functionName string, args ...any) *Context {
	return &Context{
		FunctionName: functionName,
		Args:         args,
		Metadata:     make(map[string]any),
		Results:      make([]any, 0),
	}
}

// -------------------------------------------- Public Functions --------------------------------------------

// SetResult sets a return value at the specified index.
func (aopCtx *Context) SetResult(index int, value any) {
	if index < 0 {
		return // Invalid index
	}

	// Extend results slice if needed
	for len(aopCtx.Results) <= index {
		aopCtx.Results = append(aopCtx.Results, nil)
	}
	aopCtx.Results[index] = value
}

// GetResult retrieves a return value at the specified index.
func (aopCtx *Context) GetResult(index int) any {
	if index < 0 || index >= len(aopCtx.Results) {
		return nil
	}
	return aopCtx.Results[index]
}

// HasPanic returns true if a panic was recovered during execution.
func (aopCtx *Context) HasPanic() bool {
	return aopCtx.PanicValue != nil
}

// String returns a formatted string representation of the context implementing fmt.Stringer interface.
func (aopCtx *Context) String() string {
	return fmt.Sprintf("Context{Function: %s, Args: %v, Results: %v, Error: %v, Panic: %v}",
		aopCtx.FunctionName, aopCtx.Args, aopCtx.Results, aopCtx.Error, aopCtx.PanicValue)
}
