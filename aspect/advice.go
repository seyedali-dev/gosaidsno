// Package aspect - advice defines the advice types and execution chain for AOP
package aspect

import "github.com/seyedali-dev/gosaidsno/utils"

// -------------------------------------------- Constants & Variables --------------------------------------------

const (
	Before         AdviceType = iota // Before advice executes before the target function.
	After                            // After advice executes after the target function (always runs, even on panic).
	Around                           // Around advice wraps the target function execution (can skip it).
	AfterReturning                   // AfterReturning advice executes only if the function returns successfully (no panic/error).
	AfterThrowing                    // AfterThrowing advice executes only if the function panics.
)

// -------------------------------------------- Public Functions --------------------------------------------

// AdviceType represents the type of advice to apply.
type AdviceType int

// AdviceFunc is the signature for advice functions.
// It receives the execution context and can modify it.
type AdviceFunc func(ctx *Context) error

// Advice represents a single piece of advice attached to a function.
type Advice struct {
	Type     AdviceType
	Handler  AdviceFunc
	Priority int // Higher priority executes first (for same type).
}

// AdviceChain manages a collection of advice for a single function.
type AdviceChain struct {
	before         []Advice
	after          []Advice
	around         []Advice
	afterReturning []Advice
	afterThrowing  []Advice
}

// NewAdviceChain creates a new empty advice chain.
func NewAdviceChain() *AdviceChain {
	return &AdviceChain{
		before:         make([]Advice, 0),
		after:          make([]Advice, 0),
		around:         make([]Advice, 0),
		afterReturning: make([]Advice, 0),
		afterThrowing:  make([]Advice, 0),
	}
}

// -------------------------------------------- Public Functions --------------------------------------------

// Add adds advice to the chain based on its type.
func (ac *AdviceChain) Add(advice Advice) {
	switch advice.Type {
	case Before:
		ac.before = append(ac.before, advice)
	case After:
		ac.after = append(ac.after, advice)
	case Around:
		ac.around = append(ac.around, advice)
	case AfterReturning:
		ac.afterReturning = append(ac.afterReturning, advice)
	case AfterThrowing:
		ac.afterThrowing = append(ac.afterThrowing, advice)
	}
}

// ExecuteBefore runs all Before advice in order of priority.
func (ac *AdviceChain) ExecuteBefore(ctx *Context) error {
	return ac.executeAdviceList(ac.before, ctx)
}

// ExecuteAfter runs all After advice in order of priority.
func (ac *AdviceChain) ExecuteAfter(ctx *Context) error {
	return ac.executeAdviceList(ac.after, ctx)
}

// ExecuteAround runs all Around advice in order of priority.
func (ac *AdviceChain) ExecuteAround(ctx *Context) error {
	return ac.executeAdviceList(ac.around, ctx)
}

// ExecuteAfterReturning runs all AfterReturning advice in order of priority.
func (ac *AdviceChain) ExecuteAfterReturning(ctx *Context) error {
	return ac.executeAdviceList(ac.afterReturning, ctx)
}

// ExecuteAfterThrowing runs all AfterThrowing advice in order of priority.
func (ac *AdviceChain) ExecuteAfterThrowing(ctx *Context) error {
	return ac.executeAdviceList(ac.afterThrowing, ctx)
}

// HasAround returns true if the chain has Around advice.
func (ac *AdviceChain) HasAround() bool {
	return len(ac.around) > 0
}

// Count returns the total number of advice in the chain.
func (ac *AdviceChain) Count() int {
	return len(ac.before) + len(ac.after) + len(ac.around) + len(ac.afterReturning) + len(ac.afterThrowing)
}

// -------------------------------------------- Private Helper Functions --------------------------------------------

// executeAdviceList runs a list of advice in priority order.
func (ac *AdviceChain) executeAdviceList(adviceList []Advice, ctx *Context) error {
	// Sort by priority (simple bubble sort - small lists)
	sortedAdviceList := make([]Advice, len(adviceList))
	copy(sortedAdviceList, adviceList)

	utils.BubbleSort(sortedAdviceList, utils.SortDescending, func(a, b Advice) bool {
		return a.Priority < b.Priority
	})

	// Execute in order
	for _, advice := range sortedAdviceList {
		if err := advice.Handler(ctx); err != nil {
			return err
		}
	}
	return nil
}
