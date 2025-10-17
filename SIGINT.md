# SIGINT/SIGTERM Signal Handling

**Status**: Deferred (post-v1.0)
**Priority**: HIGH - Should fix before stable release
**Complexity**: MEDIUM

## Problem

When a user presses **Ctrl+C** (SIGINT) or the process receives SIGTERM, the program exits immediately without cleanup:

- Browser process may be left running (orphaned)
- Browser tab not closed
- `defer` statements do NOT execute (signals bypass normal function returns)
- Resource leaks

### Current Code

```go
// main.go:203-208
defer func() {
    if config.CloseTab {
        logger.Verbose("Cleanup: closing tab and browser if needed")
    }
    bm.Close()
}()
```

**These defers will NOT run on SIGINT/SIGTERM.**

### Testing the Problem

```bash
# Start snag
$ snag https://slow-site.com

# Press Ctrl+C during page load

# Check for orphaned Chrome processes
$ ps aux | grep -i chrome
```

## Solution Options

### Option A: Global BrowserManager (Recommended)

Match the existing global logger pattern.

```go
// main.go - package level
var logger *Logger
var browserManager *BrowserManager  // Add this

func main() {
    // Set up signal handling
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

    go func() {
        sig := <-sigChan
        fmt.Fprintf(os.Stderr, "\nReceived %v, cleaning up...\n", sig)

        // Clean up browser if it exists
        if browserManager != nil {
            browserManager.Close()
        }

        // Exit with standard signal codes
        if sig == os.Interrupt {
            os.Exit(130) // 128 + 2 (SIGINT)
        }
        os.Exit(143) // 128 + 15 (SIGTERM)
    }()

    // ... rest of main
}

func run(c *cli.Context) error {
    // ...
    browserManager = NewBrowserManager(opts)  // Assign to global
    defer func() {
        if config.CloseTab {
            logger.Verbose("Cleanup: closing tab and browser if needed")
        }
        browserManager.Close()
        browserManager = nil  // Clear global
    }()
    // ...
}
```

**Pros:**
- ✅ Simple implementation
- ✅ Matches existing logger pattern
- ✅ Low complexity
- ✅ No function signature changes

**Cons:**
- ❌ Another global variable
- ❌ Potential race condition (unlikely in practice)

### Option B: Context-based Cancellation

Idiomatic Go approach using `context.Context`.

```go
func main() {
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

    go func() {
        <-sigChan
        fmt.Fprintln(os.Stderr, "\nInterrupted, cleaning up...")
        cancel()
    }()

    app := &cli.App{
        // ... setup
        Action: func(c *cli.Context) error {
            return run(ctx, c)
        },
    }
    // ...
}

func run(ctx context.Context, c *cli.Context) error {
    // Pass ctx to all operations
    // Check ctx.Done() in critical sections
}
```

**Pros:**
- ✅ Idiomatic Go
- ✅ Proper cancellation propagation
- ✅ No globals
- ✅ Testable

**Cons:**
- ❌ Requires refactoring all function signatures
- ❌ HIGH complexity
- ❌ May need to update rod library calls to support context
- ❌ Better suited for library code, not CLI

### Option C: Do Nothing

Accept the limitation and document it.

**Pros:**
- ✅ Zero effort

**Cons:**
- ❌ Unprofessional
- ❌ Poor user experience
- ❌ Resource leaks

## Recommendation

**Use Option A (Global BrowserManager)** for the following reasons:

1. **Consistency**: Already using global logger, this matches the pattern
2. **Simplicity**: Minimal code changes, easy to implement
3. **Effectiveness**: Solves the problem completely
4. **CLI-appropriate**: For a single-threaded CLI tool, global state is acceptable

**Save Option B for library extraction** if snag ever becomes importable code.

## Implementation Checklist

- [ ] Add `var browserManager *BrowserManager` to package globals
- [ ] Add signal handler in `main()`
- [ ] Update `run()` to assign to global browserManager
- [ ] Update `run()` defer to clear global after cleanup
- [ ] Add proper exit codes (130 for SIGINT, 143 for SIGTERM)
- [ ] Test signal handling:
  - [ ] Ctrl+C during page load
  - [ ] Ctrl+C during conversion
  - [ ] Verify browser closes
  - [ ] Verify no orphaned processes
- [ ] Document signal handling in README

## Exit Codes

Standard Unix signal exit codes:
- `130` = 128 + 2 (SIGINT / Ctrl+C)
- `143` = 128 + 15 (SIGTERM)
- `1` = General error (keep for normal errors)

## References

- PROJECT.md Issue #5
- Go signal package: https://pkg.go.dev/os/signal
- Unix signal handling best practices
