# Snag Polish

The following tasks will polish off the final stages of development for snag:

- ✅ Remove the "auto-generated" from filename terminal messages, just "Filename:" will do.
- ❓ Support terminal messages of "Fetching" when actually getting the content to stdout/file, and "Navigating" for just browser navigation. (Unknown - unable to reproduce issue)
- ✅ Showing the Usage after some of the errors is annoying. This should be changed.
- ✅ Same with list-tabs when "index out of range" and maybe other errors. Annoying.

## Issues

✅ **Fixed:** Browser connection messages

In this test:

```txt
./snag --open-browser https://example.com
Opening 1 valid URL in browser...
Launching browser in visible mode...
✓ Chrome launched in visible mode
[1/1] Opening: https://example.com
✓ [1/1] Opened: https://example.com
✓ Browser will remain open with 1 tabs
Use 'snag --list-tabs' to see opened tabs
Use 'snag --tab <index>' to fetch content from a tab
```

The browser was already open and it got used, but the terminal messages stated "Launching browser in visible mode... ✓ Chrome launched in visible mode"?

**Resolution:** Now shows "✓ Connected to existing browser (visible mode)" when connecting to an existing browser.

---

In this one:

```txt
./snag --open-browser --user-data-dir ./test-profile
Opening browser...
✓ Browser already running on port 9222
You can connect to it using: snag <url>
```

Shouldn't there be a warning that the profile can't be used?

## Review

- Review the [DEBUG] messages in the source code, maybe add more?
