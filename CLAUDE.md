# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build Commands

```bash
make build          # Build binary to ./bin/g4d
make test           # Run tests with race detection and coverage
make lint           # Run golangci-lint (install: curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin)
make fmt            # Format code with go fmt and gofmt -s
make vet            # Run go vet static analysis
make test-coverage  # Generate coverage.html report
```

### Running a Single Test

```bash
go test -v -run TestFunctionName ./internal/package
# Example: go test -v -run TestDetectLinuxDistro ./internal/platform
```

## Architecture

go4dot is a CLI tool for managing dotfiles using GNU Stow, with platform detection and dependency management.

### Package Structure

```
cmd/g4d/main.go           # CLI entry point, all Cobra commands
internal/
  platform/               # OS/distro detection + package manager abstraction
    detect.go             # Platform struct with Detect() method
    packages.go           # PackageManager interface
    packages_{dnf,apt,brew}.go  # Strategy implementations
  config/                 # YAML config loading from .go4dot.yaml
    schema.go             # Config, Dependencies, ConfigItem structs
    loader.go             # LoadFromFile, Discover functions
    validator.go          # Validation logic
  deps/                   # Dependency checking/installation
    check.go              # CheckDependencies using platform detection
    install.go            # InstallDependencies using package managers
  stow/                   # GNU stow wrapper for symlink management
    manager.go            # Stow, Unstow, RestowConfigs functions
```

### Key Design Patterns

**PackageManager interface** (strategy pattern) - each package manager (dnf, apt, brew, pacman) implements:
```go
type PackageManager interface {
    Name() string
    IsAvailable() bool
    Install(packages ...string) error
    IsInstalled(pkg string) bool
    Update() error
    NeedsSudo() bool
}
```

**Config loading flow**: Discover config file -> Parse YAML -> Validate -> Return typed Config struct

### CLI Command Groups

- `g4d detect` - Show platform info
- `g4d config {validate,show}` - Config operations
- `g4d deps {check,install}` - Dependency management
- `g4d stow {add,remove,refresh}` - Symlink management

## Testing Patterns

Tests use table-driven patterns. Example:
```go
func TestSomething(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
    }{...}
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {...})
    }
}
```

## Development Notes

- GNU stow must be installed on the system (not bundled)
- Config files are discovered from: `.`, `~/dotfiles`, `~/.dotfiles`
- Package names may differ across distros - use `MapPackageName()` in `platform/packages.go`
- Error wrapping: use `fmt.Errorf("context: %w", err)`
- See PLAN.md for the 14-phase implementation roadmap and current progress
- Don't attribute CLAUDE Code in commits
