# snag Polish No2 - test-interactive

The issues listed here were noticed during the ./test-interactive test run.

---

**FIXED**

In the following example, the "Use 'snag... etc" messages are a little too much and need to be removed.

```txt
./snag --open-browser https://example.com
Opening 1 valid URL in browser...
✓ Connected to existing browser (visible mode)
[1/1] Opening: https://example.com
✓ [1/1] Opened: https://example.com
✓ Browser will remain open with 1 tabs
Use 'snag --list-tabs' to see opened tabs
Use 'snag --tab <index>' to fetch content from a tab
```

---

**FIXED**

In the following example, I can't find the test-profile directory:

```txt
./snag --open-browser --user-data-dir ./test-profile
Opening browser...
✓ Browser opened on port 9222
Browser is running with remote debugging enabled
You can now connect to it using: snag <url>
```

**Fix:** Updated browser.go:OpenBrowserOnly() to apply user-data-dir and user-agent settings to the launcher. Changed validate.go:validateUserDataDir() to create the directory using os.MkdirAll(path, 0755) instead of letting Chrome create it. Updated docs/arguments/user-data-dir.md to reflect the new behavior.

---

**FIXED**

This test failed:

```txt
Test 76/76: Error on non-existent profile dir
─────────────────────────────────────────────
Section: Error Tests
Working directory: /tmp/snag-test-a5ZT7o
────────────────────────────────────────

./snag --user-data-dir /nonexistent/path https://example.com
✓ Chromium launched in headless mode
Fetching https://example.com...
✓ Fetched successfully
# Example Domain

This domain is for use in documentation examples without needing permission. Avoid use in operations.

[Learn more](https://iana.org/domains/example)

────────────────────────────────────────
✖ Expected error but command succeeded
```

**Fix:** The validation now creates directories with os.MkdirAll() but properly errors on truly inaccessible paths (e.g., /nonexistent/path errors with "read-only file system"). Test expectations now align with actual behavior.

---

I tested on Linux and the browser is still running in the background after closing. This needs some research.

