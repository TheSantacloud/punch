# Contributing to Punch

Contributions to Punch are highly encouraged and appreciated. To maintain the quality and consistency of the codebase, please follow these guidelines:

## Update Interfaces

If you're making changes that affect the interfaces for the repositories, you should regenerate the mock implementations. This can be done using the following command:

```bash
mockgen -source=pkg/repositories/interfaces.go -destination=pkg/repositories/mock_repository.go -package=repositories
```

This command uses `mockgen` to create a new mock implementation based on the updated interface.

## Linting

Code quality is crucial, so please ensure your contributions pass our linting checks. We use `golangci-lint` for this purpose. Linting is automatically checked in GitHub Actions, and changes that do not pass these checks will not be approved. This helps us maintain a consistent code quality and style.

## Code Formatting

Please use `go fmt` to format your code before submitting a pull request. This ensures that all code in the repository adheres to the standard Go formatting guidelines, making it more readable and maintainable.

