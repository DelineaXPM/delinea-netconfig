# delinea-netconfig demo
# Run from the repo root: vhs demo/demo.tape

Output demo/demo.gif

Set Theme "Dracula"
Set FontSize 13
Set Width 1200
Set Height 700
Set Shell "zsh"
Set Padding 20
Set FontFamily "JetBrains Mono"
Set PlaybackSpeed 1

# ─── Build ────────────────────────────────────────────────────────────────────
Hide
Type "make build 2>/dev/null"
Enter
Sleep 3s
Type "clear"
Enter
Show

# ─── 1. Help overview ─────────────────────────────────────────────────────────
Sleep 500ms
Type "delinea-netconfig --help"
Enter
Sleep 3s

Type "clear"
Enter
Sleep 300ms

# ─── 2. Info — statistics ─────────────────────────────────────────────────────
Type "delinea-netconfig info testdata/network-requirements.json"
Enter
Sleep 3.5s

Type "clear"
Enter
Sleep 300ms

# ─── 3. Info — latest version from setup.delinea.app ─────────────────────────
Type "delinea-netconfig info --latest"
Enter
Sleep 4s

Type "clear"
Enter
Sleep 300ms

# ─── 4. Convert — CSV (pipe to head for brevity) ─────────────────────────────
Type "delinea-netconfig convert -f testdata/network-requirements.json --format csv | head -6"
Enter
Sleep 2.5s

Type "clear"
Enter
Sleep 300ms

# ─── 5. Convert — Terraform with tenant substitution ─────────────────────────
Type "delinea-netconfig convert -f testdata/tenant-test.json --format terraform --tenant acme | head -12"
Enter
Sleep 2.5s

Type "clear"
Enter
Sleep 300ms

# ─── 6. Diff — compare two versions ──────────────────────────────────────────
Type "delinea-netconfig diff testdata/tenant-test.json testdata/network-requirements.json --summary"
Enter
Sleep 3s

Type "clear"
Enter
Sleep 300ms

# ─── 7. Validate ─────────────────────────────────────────────────────────────
Type "delinea-netconfig validate -f testdata/network-requirements.json"
Enter
Sleep 2.5s

Type "clear"
Enter
Sleep 300ms

# ─── 8. TUI — interactive browser ────────────────────────────────────────────
Type "delinea-netconfig tui -f testdata/network-requirements.json"
Enter
Sleep 2.5s

# Navigate down a few entries
Down
Sleep 200ms
Down
Sleep 200ms
Down
Sleep 300ms

# Filter by text
Type "/"
Sleep 400ms
Type "platform"
Sleep 1.5s
Escape
Sleep 500ms

# Open detail view
Enter
Sleep 2s

# Go back to browser
Escape
Sleep 500ms

# Filter by region
Type "r"
Sleep 400ms
Type "us"
Enter
Sleep 1.5s

# Clear region filter
Type "x"
Sleep 800ms

# Quit
Type "q"
Sleep 500ms
