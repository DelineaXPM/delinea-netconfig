# delinea-netconfig demo
# Run from the repo root: vhs demo/demo.tape
# Requires: make build (binary must exist at ./delinea-netconfig)

Output demo/demo.gif

Set Theme "Dracula"
Set FontSize 13
Set Width 1200
Set Height 700
Set Shell "zsh"
Set Padding 20
Set FontFamily "JetBrains Mono"
Set PlaybackSpeed 1

# ─── Setup (hidden) ───────────────────────────────────────────────────────────
Hide
Type "make build 2>/dev/null && export PATH=$PATH:$(pwd)"
Enter
Sleep 3s
Type "clear"
Enter
Show

# ─── 1. Check the latest published requirements ───────────────────────────────
Sleep 300ms
Type "# Check what Delinea has published — no file needed"
Enter
Sleep 600ms
Type "delinea-netconfig info --latest"
Enter
Sleep 5s

Type "clear"
Enter
Sleep 300ms

# ─── 2. Full statistics — fetched live from setup.delinea.app ─────────────────
Type "# Full statistics fetched live from setup.delinea.app"
Enter
Sleep 600ms
Type "delinea-netconfig info"
Enter
Sleep 5s

Type "clear"
Enter
Sleep 300ms

# ─── 3. Convert to CSV — straight from the URL ───────────────────────────────
Type "# Convert to CSV — straight from setup.delinea.app"
Enter
Sleep 600ms
Type "delinea-netconfig convert -u https://setup.delinea.app/network-requirements --format csv | head -8"
Enter
Sleep 6s

Type "clear"
Enter
Sleep 300ms

# ─── 4. Convert to Terraform with tenant substitution ────────────────────────
Type "# Terraform output with your tenant name substituted in"
Enter
Sleep 600ms
Type "delinea-netconfig convert -u https://setup.delinea.app/network-requirements --format terraform --tenant mycompany | head -14"
Enter
Sleep 6s

Type "clear"
Enter
Sleep 300ms

# ─── 5. Validate the live requirements ───────────────────────────────────────
Type "# Validate the live requirements"
Enter
Sleep 600ms
Type "delinea-netconfig validate"
Enter
Sleep 5s

Type "clear"
Enter
Sleep 300ms

# ─── 6. Raw JSON — like curl | jq but built-in ───────────────────────────────
Type "# Pretty-print the raw JSON — built-in curl | jq"
Enter
Sleep 600ms
Type "delinea-netconfig info --raw | head -20"
Enter
Sleep 5s

Type "clear"
Enter
Sleep 300ms

# ─── 7. Interactive TUI — defaults to setup.delinea.app ──────────────────────
Type "# Interactive TUI — no flags needed, loads from setup.delinea.app"
Enter
Sleep 600ms
Type "delinea-netconfig tui"
Enter
Sleep 4s

# Navigate down a few entries
Down
Sleep 250ms
Down
Sleep 250ms
Down
Sleep 250ms
Down
Sleep 300ms

# Live text filter
Type "/"
Sleep 400ms
Type "platform"
Sleep 1.8s
Escape
Sleep 500ms

# Open entry detail
Enter
Sleep 2s

# Back to browser
Escape
Sleep 400ms

# Filter by region
Type "r"
Sleep 400ms
Type "us"
Enter
Sleep 1.5s

# Clear region filter and quit
Type "x"
Sleep 600ms
Type "q"
Sleep 500ms
