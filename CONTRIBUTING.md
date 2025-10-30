# Contributing to gosaidsno

Thanks for your interest in contributing! 🎉

---

## Code Style

We follow **Go best practices** and **clean code principles**:

### General Guidelines

- **SOLID principles** - Single responsibility, open/closed, etc.
- **DRY** - Don't repeat yourself
- **Clear naming** - Functions and variables should be self-documenting
- **Minimal comments** - Code should be readable without excessive comments
- **Error handling** - Always handle errors explicitly

### Go-Specific

- Run `go fmt` before committing
- Keep functions focused and small (<50 lines)
- Use meaningful variable names (avoid `x`, `y`, `z` except in math contexts)
- Document exported functions with proper godoc format

### Documentation Format

#### Package-level docstring:

```go
// Package aspect - advice defines the advice types and execution chain for AOP
package aspect
```

#### Function docstring (one-liner):

```go
// NewContext creates a new execution context for the given function.
func NewContext(functionName string, args ...any) *Context {
```

#### Comment blocks to separate code:

```go
// -------------------------------------------- Types --------------------------------------------

type Aspect struct {}

// -------------------------------------------- Public Functions --------------------------------------------

func PublicFunc() {}

// -------------------------------------------- Private Helper Functions --------------------------------------------

func privateHelper() {}
```

---

## Testing

### Writing Tests

- Place tests in `*_test.go` files
- Use table-driven tests where appropriate
- Test edge cases and error paths
- Aim for >80% coverage

### Running Tests

```bash
# Run all tests
make test

# Run with coverage
make coverage

# Run benchmarks
make bench
```

### Test Naming

```go
func TestFunctionName_Behavior(t *testing.T) {
    // Arrange
    // Act
    // Assert
}
```

---

## Adding New Wrapper Functions

If you need to add support for a new function signature:

1. **Follow the naming pattern:**
    - `Wrap<num_args><return_type_codes>`
    - Example: `Wrap3R2E` = 3 args, 2 results + error

2. **Use generics for type safety:**
   ```go
   func Wrap3RE[A, B, C, R any](name string, fn func(A, B, C) (R, error)) func(A, B, C) (R, error)
   ```

3. **Call executeWithAdvice:**
   ```go
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
   ```

4. **Add tests** in `aspect/wrap_test.go`

5. **Update QUICKSTART.md** with the new wrapper in the table

---

## Pull Request Process

1. **Fork the repository**

2. **Create a feature branch:**
   ```bash
   git checkout -b feat/your-feature-name
   ```

3. **Make your changes:**
    - Write code
    - Add tests
    - Update docs if needed

4. **Run checks:**
   ```bash
   make fmt
   make test
   make lint
   ```

5. **Commit with conventional commits:**
   ```
   feat(wrap): add Wrap4RE for four-argument functions
   
   - Adds support for func(A, B, C, D) (R, error)
   - Includes tests and documentation
   ```

6. **Push and create PR:**
   ```bash
   git push origin feat/your-feature-name
   ```

7. **PR Guidelines:**
    - Clear title and description
    - Reference any related issues
    - Ensure CI passes
    - Respond to review feedback

---

## Commit Message Format

We use **conventional commits**:

```
<type>(<scope>): <subject>

<body>
```

### Types:

- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `chore`: Build/tooling changes
- `refactor`: Code refactoring
- `test`: Adding tests
- `perf`: Performance improvements

### Examples:

```
feat(advice): add AfterFinally advice type

Adds a new advice type that runs after After advice.
Useful for cleanup that must happen last.

fix(wrap): handle nil context in Wrap2E

Previously panicked on nil context. Now returns error.

docs(readme): update installation instructions

Added go install command and module path.

chore(deps): update golangci-lint to v1.55

test(advice): add test for priority ordering
```

---

## Project Structure

```
gosaidsno/
├── aspect/              # Core library code
│   ├── advice.go        # Advice types & chain
│   ├── context.go       # Execution context
│   ├── registry.go      # Function registry
│   ├── wrap.go          # Generic wrappers
│   └── *_test.go        # Tests for each file
├── examples/            # Usage examples
│   └── basic_usage.go   # Comprehensive demo
├── QUICKSTART.md        # Quick start guide
├── ARCHITECTURE.md      # Design documentation
├── CONTRIBUTING.md      # This file
├── README.md            # Main readme
├── Makefile             # Build automation
├── go.mod               # Module definition
└── .gitignore           # Git ignore rules
```

---

## Development Setup

1. **Install Go 1.21+:**
   ```bash
   go version  # Should be 1.21 or higher
   ```

2. **Clone the repo:**
   ```bash
   git clone https://github.com/yourusername/gosaidsno.git
   cd gosaidsno
   ```

3. **Install dev tools:**
   ```bash
   # golangci-lint
   go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
   ```

4. **Run tests:**
   ```bash
   make test
   ```

5. **Start coding!**

---

## Questions?

- Open an issue for bugs or feature requests
- Check existing issues before creating new ones
- Be respectful and constructive

---

## Code of Conduct

- Be professional and respectful
- Focus on constructive feedback
- Help others learn
- Assume good intent

---

**Thanks for contributing to gosaidsno!** 🚀