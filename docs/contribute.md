# Contribution Guidelines

## Table of Contents

- [Coding Style](#coding-style)
- [Comment Guidelines](#comment-guidelines)
- [Git Workflow](#git-workflow)
- [Pull Request Process](#pull-request-process)
- [Testing](#testing)
- [Documentation](#documentation)

## Coding Style

### General Guidelines

- Follow Go best practices and idioms
- Use consistent naming conventions:
  - PascalCase for exported functions, types, and variables
  - camelCase for unexported functions, types, and variables
  - snake_case for filenames and package names
  - UPPER_SNAKE_CASE for constants

- Keep functions small and focused (single responsibility principle)
- Use descriptive function and variable names
- Avoid magic numbers and strings
- Use constants for configuration values that don't change at runtime

### Code Structure

- Group related code into packages
- Follow the standard Go project structure:
  - `cmd/` for command-line applications
  - `pkg/` for library code (if applicable)
  - `internal/` for private application code
  - `api/` for API definitions
  - `docs/` for documentation
  - `tests/` for test files

## Comment Guidelines

### General Rules

- **All comments must be in English**
- Use **new line comments** instead of end-of-line comments
- Keep comments concise and meaningful
- Update comments when you update the code
- Use comments to explain why something is done, not what is done (the code should show that)

### Example of Good Comments

```go
// Calculate the daily summary by aggregating cash flows for the given date
func CalculateDailySummary(date string) (*Summary, error) {
    // Validate date format before processing
    if err := validation.ValidateDate(date); err != nil {
        return nil, err
    }
    
    // Implementation...
}
```

### Example of Bad Comments

```go
func CalculateDailySummary(date string) (*Summary, error) {
    if err := validation.ValidateDate(date); err != nil { // Validate date
        return nil, err
    }
    
    // This loop iterates through the cash flows
    for _, flow := range cashFlows {
        // Add to total
        total += flow.Amount
    }
}
```

### Documentation Comments

- Use GoDoc comments for all exported functions, types, and variables
- Start with a capital letter and end with a period
- Use the third person singular form ("Calculates" instead of "Calculate")

```go
// CalculateDailySummary calculates the daily summary for the given date.
// It returns a summary object with income, expense, and balance.
func CalculateDailySummary(date string) (*Summary, error) {
    // Implementation...
}
```

## Git Workflow

### Branch Naming

Use the following branch naming convention:

```
<type>/<short-description>
```

Where `<type>` is one of:
- `feat` for new features
- `fix` for bug fixes
- `docs` for documentation changes
- `refactor` for code refactoring
- `test` for test-related changes
- `ci` for CI/CD changes

Examples:
- `feat/add-user-authentication`
- `fix/cash-flow-calculation`
- `docs/update-api-documentation`

### Commit Messages

Follow the Conventional Commits format:

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

Where `<type>` is one of:
- `feat` for new features
- `fix` for bug fixes
- `docs` for documentation changes
- `refactor` for code refactoring
- `test` for test-related changes
- `ci` for CI/CD changes
- `chore` for other changes (e.g., dependencies, build scripts)

Example:

```
feat: add daily summary endpoint

Add a new endpoint to calculate and return daily cash flow summaries
```

## Pull Request Process

1. Create a new branch from `develop`
2. Make your changes
3. Run all tests to ensure they pass
4. Update documentation if necessary
5. Create a pull request to `develop`
6. Request review from at least one team member
7. Address any review comments
8. Once approved, merge the pull request

## Testing

- Write unit tests for all new functions and methods
- Aim for at least 80% test coverage
- Use table-driven tests for multiple test cases
- Test edge cases and error conditions
- Run tests before submitting a pull request

```bash
# Run all tests
go test ./...

# Run tests for a specific package
go test ./service/cash_flow_service

# Run tests with coverage
go test -cover ./...
```

## Documentation

- Update the OpenAPI spec (`docs/openapi.yaml`) when adding or modifying endpoints
- Generate HTML documentation from the OpenAPI spec using the provided scripts:
  - Bash: `./scripts/generate-docs.sh`
  - PowerShell: `./scripts/generate-docs.ps1`
- Keep README.md and other documentation files up to date
- Document any breaking changes

## Code Review Checklist

Before submitting a pull request, ensure your code meets the following criteria:

- [ ] Follows the coding style guidelines
- [ ] Has appropriate comments
- [ ] Passes all tests
- [ ] Has adequate test coverage
- [ ] Updates documentation if necessary
- [ ] Follows the git workflow
- [ ] Has a clear and descriptive commit message

## Getting Help

If you have any questions or need help, please:

- Check the existing documentation
- Ask in the project's communication channel
- Create an issue for clarification

Thank you for contributing to CashLenX!
