# snag - UX Improvements

User experience enhancements to be implemented in future iterations.

**Status**: Planning phase
**Priority**: Post-v1.0

---

## 1. Progress Indicators for Long Operations

**Current Behavior**:
Long operations have no visual feedback, leaving users uncertain if the tool is working:
- Browser launch: 2-5 seconds of silence
- Page load: 5-30 seconds of silence
- Conversion: Usually <1 second (fine)

**User Experience Issue**:
```bash
$ snag https://slow-site.com
# ... 10 seconds of silence ...
# User: "Is it working? Did it hang?"
```

**Proposed Solution**:
Add spinner or progress indicators during long operations:

```bash
$ snag https://example.com
⠋ Launching browser...
✓ Chrome launched in headless mode
⠙ Fetching https://example.com...
✓ Fetched successfully
⠹ Converting HTML to Markdown...
✓ Converted to Markdown
```

**Implementation Options**:
1. **Spinner library** - Use package like `github.com/briandowns/spinner`
2. **Progress dots** - Simple periodic dots: `Fetching...........`
3. **Custom progress** - Build minimal progress indicator

**Considerations**:
- Must write to stderr (don't interfere with stdout output)
- Should respect quiet mode
- Should work with NO_COLOR environment variable
- Spinner should stop cleanly on completion/error

**Complexity**: MEDIUM

---

## 2. WaitFor Element Timeout Feedback

**Current Behavior**:
When using `--wait-for` selector, if the element never appears, users wait the full page timeout (default 30s) with no feedback.

**Location**: fetch.go:71-80

**User Experience Issue**:
```bash
$ snag https://example.com --wait-for "#missing-element"
# ... 30 seconds of silence ...
✗ failed to find selector #missing-element
```

**Proposed Solution**:

**Option 1**: Progress feedback during wait
```bash
$ snag https://example.com --wait-for "#content"
⠋ Waiting for selector: #content (timeout: 30s)...
✓ Selector found: #content
```

**Option 2**: Separate timeout for element wait
```bash
$ snag https://example.com --wait-for "#content" --wait-timeout 5
# Uses 5s for element wait instead of full page timeout
```

**Option 3**: Both - progress feedback + configurable timeout
```bash
$ snag https://example.com --wait-for "#content" --wait-timeout 10
⠋ Waiting for selector: #content (timeout: 10s)...
⏱  5s elapsed...
✓ Selector found: #content
```

**Implementation**:
```go
// Add progress feedback
logger.Progress("Waiting for selector: %s (timeout: %ds)", opts.WaitFor, timeout)

// Or add separate timeout
elem, err := page.Timeout(opts.WaitTimeout * time.Second).Element(opts.WaitFor)
```

**Complexity**: LOW

---

## Future Considerations

- **ETA estimation**: Show estimated time remaining for known-slow operations
- **Verbose progress**: Show sub-steps in verbose mode (DNS lookup, connection, rendering, etc.)
- **Progress bars**: For multi-page operations or batch mode
- **Interactive mode**: Allow Ctrl+C to gracefully cancel with cleanup
- **Color customization**: Allow users to customize color scheme

---

**Document Version**: 1.0
**Created**: 2025-10-17
**Last Updated**: 2025-10-17
