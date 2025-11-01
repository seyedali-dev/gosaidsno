# Examples Directory

## Running Examples

```bash
# Basic usage - logging, timing, validation
go run examples/01_basic_usage/main.go

# Caching with Around advice - Broken... needs fix
go run examples/02_caching_pattern/main.go

# Authentication and authorization
go run examples/03_authentication/main.go

# Circuit breaker for fault tolerance
go run examples/04_circuit_breaker/main.go

# Retry with exponential backoff
go run examples/05_retry_pattern/main.go
```

## Examples Overview

### 01_basic_usage
**Real-world use cases:**
- Request/response logging
- Method execution timing
- Input validation (Before advice)
- Success-only actions (AfterReturning)

**Key patterns:**
- Setup AOP once at startup
- Wrap functions during initialization
- Use metadata to pass data between advice

### 02_caching_pattern - Broken... needs fix
**Real-world use cases:**
- Database query caching
- API response caching
- TTL-based cache expiration
- Cache hit/miss metrics

**Key patterns:**
- Around advice checks cache
- Skip execution on cache hit
- AfterReturning populates cache

### 03_authentication
**Real-world use cases:**
- Token validation
- Session management
- Role-based access control (RBAC)
- Audit logging

**Key patterns:**
- Before advice for auth (priority 100)
- Before advice for authz (priority 90)
- Store user info in metadata
- After advice for audit trails

### 04_circuit_breaker
**Real-world use cases:**
- External API fault tolerance
- Database connection protection
- Service degradation handling

**Key patterns:**
- Around advice checks circuit state
- After advice tracks failures
- State transitions (CLOSED → OPEN → HALF_OPEN)

### 05_retry_pattern
**Real-world use cases:**
- Transient failure handling
- Network request retries
- Exponential backoff

**Key patterns:**
- Before advice initializes retry state
- After advice implements retry logic
- Exponential backoff calculation

## Project Setup Pattern

All examples follow this pattern:

```go
// 1. Setup AOP once at startup
func setupAOP() {
    aspect.MustRegister("FunctionName")
    aspect.MustAddAdvice("FunctionName", /* ... */)
}

// 2. Implement business logic
func businessLogicImpl(args...) result {
    // Pure business logic
}

// 3. Wrap functions (once, during initialization)
var BusinessLogic = aspect.Wrap*("FunctionName", businessLogicImpl)

// 4. Use normally throughout application
func main() {
    setupAOP()  // Once
    
    // Use wrapped functions
    result := BusinessLogic(args...)
}
```

## Advice Execution Order

For a function with all advice types:

1. **Before** (high priority → low)
2. **Around** (can skip step 3)
3. Target function
4. **AfterReturning** (only if success)
5. **AfterThrowing** (only if panic)
6. **After** (always runs)

## Common Metadata Keys

Examples use these metadata conventions:

- `startTime` - For timing (type: `time.Time`)
- `userID` - For auth/audit (type: `string`)
- `role` - For authorization (type: `string`)
- `attempt` - For retry logic (type: `int`)
- `maxRetries` - Retry configuration (type: `int`)

## Performance Notes

Each example includes:
- Timing measurements
- Success/failure counts
- Performance impact of advice
- Cache hit rates (where applicable)