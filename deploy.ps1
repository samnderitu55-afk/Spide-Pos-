# deploy.ps1 — build Linux binary, upload all changed files, print server-side swap commands
param(
    [string]$Server = "root@64.226.124.37",
    [switch]$SkipBinary,
    [switch]$SkipJS
)

$ErrorActionPreference = "Stop"
$root = "C:\Users\sam\spide-pos"
$stamp = Get-Date -Format "yyyyMMdd_HHmm"

Push-Location $root

if (-not $SkipBinary) {
    Write-Host "→ Building Linux binary..." -ForegroundColor Cyan
    $env:GOOS = "linux"; $env:GOARCH = "amd64"; $env:CGO_ENABLED = "0"
    go build -o spide-pos cmd/server/main.go
    if ($LASTEXITCODE -ne 0) { Pop-Location; throw "Go build failed" }
    $size = (Get-Item .\spide-pos).Length
    Write-Host "  Binary: $size bytes" -ForegroundColor Green

    Write-Host "→ Uploading binary..." -ForegroundColor Cyan
    scp "$root\spide-pos" "${Server}:/tmp/spide-pos.new"
}

if (-not $SkipJS) {
    Write-Host "→ Uploading templates + JS..." -ForegroundColor Cyan
    scp "$root\internal\templates\pos.html"    "${Server}:/tmp/pos.html.new"
    scp "$root\static\js\spide-modals.js"      "${Server}:/tmp/spide-modals.js.new"
    scp "$root\static\js\spide-loaders.js"     "${Server}:/tmp/spide-loaders.js.new"
}

Write-Host ""
Write-Host "═══════════════════════════════════════════════════════════" -ForegroundColor Yellow
Write-Host "Files uploaded. Now paste this on the server:" -ForegroundColor Yellow
Write-Host "═══════════════════════════════════════════════════════════" -ForegroundColor Yellow
Write-Host ""
Write-Host @"
cd /var/www/spide

# Back up current versions
cp spide-pos                                  /root/backups/spide-pos.${stamp}           2>/dev/null || true
cp internal/templates/pos.html                /root/backups/pos.html.${stamp}            2>/dev/null || true
cp static/js/spide-modals.js                  /root/backups/spide-modals.js.${stamp}     2>/dev/null || true
cp static/js/spide-loaders.js                 /root/backups/spide-loaders.js.${stamp}    2>/dev/null || true

# Swap in new files
mv /tmp/spide-pos.new          spide-pos
mv /tmp/pos.html.new           internal/templates/pos.html
mv /tmp/spide-modals.js.new    static/js/spide-modals.js
mv /tmp/spide-loaders.js.new   static/js/spide-loaders.js

chmod +x spide-pos

# Restart (only needed for the binary, but harmless for JS/template-only changes)
systemctl restart spide-pos
sleep 2
systemctl status spide-pos --no-pager
"@
Write-Host ""
Pop-Location