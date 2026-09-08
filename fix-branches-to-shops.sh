#!/bin/bash

# ============================================
# SPIDE POS - branches → shops Migration Script
# ============================================

set -e

echo "🕷️  Starting branches → shops migration..."
echo "============================================"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Backup
echo -e "${YELLOW}📦 Creating backup...${NC}"
BACKUP_DIR="./backup_before_migration_$(date +%Y%m%d_%H%M%S)"
mkdir -p "$BACKUP_DIR"
cp -r . "$BACKUP_DIR" 2>/dev/null || true
echo -e "${GREEN}✅ Backup created at: $BACKUP_DIR${NC}"

# Update Go files
echo -e "${YELLOW}🔧 Updating Go files...${NC}"
find . -name "*.go" -not -path "./backup*" -not -path "./vendor*" -type f | while read -r file; do
    sed -i 's/branches/shops/g' "$file" 2>/dev/null || true
    sed -i 's/Branches/Shops/g' "$file" 2>/dev/null || true
    sed -i 's/Branch/Shop/g' "$file" 2>/dev/null || true
    echo -e "  ${GREEN}✓${NC} Updated: $file"
done

# Update HTML files
echo -e "${YELLOW}🌐 Updating HTML files...${NC}"
find . -name "*.html" -not -path "./backup*" -not -path "./vendor*" -type f | while read -r file; do
    sed -i 's/\/api\/branches/\/api\/shops/g' "$file" 2>/dev/null || true
    sed -i 's/branches/shops/g' "$file" 2>/dev/null || true
    sed -i 's/Branches/Shops/g' "$file" 2>/dev/null || true
    echo -e "  ${GREEN}✓${NC} Updated: $file"
done

# Update JS files
echo -e "${YELLOW}📜 Updating JavaScript files...${NC}"
find . -name "*.js" -not -path "./backup*" -not -path "./vendor*" -type f | while read -r file; do
    sed -i 's/\/api\/branches/\/api\/shops/g' "$file" 2>/dev/null || true
    sed -i 's/branches/shops/g' "$file" 2>/dev/null || true
    sed -i 's/Branches/Shops/g' "$file" 2>/dev/null || true
    echo -e "  ${GREEN}✓${NC} Updated: $file"
done

# Remove old files
echo -e "${YELLOW}🗑️  Removing old files...${NC}"
rm -f internal/db/branches.go 2>/dev/null && echo -e "  ${GREEN}✓${NC} Removed: internal/db/branches.go"
rm -f internal/handlers/branches.go 2>/dev/null && echo -e "  ${GREEN}✓${NC} Removed: internal/handlers/branches.go"

# Check for remaining references
echo -e "${YELLOW}🔍 Checking for remaining references...${NC}"
REMAINING=$(grep -r "branches" --include="*.go" --include="*.html" --include="*.js" --exclude-dir="backup*" --exclude-dir="vendor*" . 2>/dev/null | grep -v "shops" | wc -l || echo "0")

if [ "$REMAINING" -eq "0" ]; then
    echo -e "${GREEN}✅ No remaining 'branches' references found!${NC}"
else
    echo -e "${RED}⚠️  Found $REMAINING remaining references:${NC}"
    grep -r "branches" --include="*.go" --include="*.html" --include="*.js" --exclude-dir="backup*" --exclude-dir="vendor*" . 2>/dev/null | grep -v "shops"
fi

echo ""
echo "============================================"
echo -e "${GREEN}✅ Migration Complete!${NC}"
echo "============================================"
echo ""
echo "📊 Summary:"
echo "  - Backup: $BACKUP_DIR"
echo "  - All files updated"
echo "  - Old files removed"
echo ""
echo "📝 Next Steps:"
echo "  1. Run the database migration script"
echo "  2. Test the application"
echo ""
EOF./fix-branches-to-shops.sh
./fix-branches-to-shops.sh
cat > fix-branches-to-shops.ps1 << 'EOF'
# ============================================
# SPIDE POS - branches → shops Migration Script
# PowerShell Version for Windows
# ============================================

Write-Host "🕷️  Starting branches -> shops migration..." -ForegroundColor Cyan
Write-Host "============================================" -ForegroundColor Cyan

# Backup
Write-Host "📦 Creating backup..." -ForegroundColor Yellow
$BACKUP_DIR = "backup_before_migration_$(Get-Date -Format 'yyyyMMdd_HHmmss')"
New-Item -ItemType Directory -Path $BACKUP_DIR -Force | Out-Null
Copy-Item -Path ".\*" -Destination $BACKUP_DIR -Recurse -Force -ErrorAction SilentlyContinue
Write-Host "✅ Backup created at: $BACKUP_DIR" -ForegroundColor Green

# Update Go files
Write-Host "`n🔧 Updating Go files..." -ForegroundColor Yellow
Get-ChildItem -Recurse -Filter "*.go" -Exclude "backup*","vendor*" | ForEach-Object {
    $content = Get-Content $_.FullName -Raw
    $content = $content -replace 'branches', 'shops' `
                        -replace 'Branches', 'Shops' `
                        -replace 'Branch', 'Shop' `
                        -replace 'branch', 'shop'
    Set-Content -Path $_.FullName -Value $content -NoNewline
    Write-Host "  ✓ Updated: $($_.FullName)" -ForegroundColor Green
}

# Update HTML files
Write-Host "`n🌐 Updating HTML files..." -ForegroundColor Yellow
Get-ChildItem -Recurse -Filter "*.html" -Exclude "backup*","vendor*" | ForEach-Object {
    $content = Get-Content $_.FullName -Raw
    $content = $content -replace '/api/branches', '/api/shops' `
                        -replace 'branches', 'shops' `
                        -replace 'Branches', 'Shops' `
                        -replace 'Branch', 'Shop'
    Set-Content -Path $_.FullName -Value $content -NoNewline
    Write-Host "  ✓ Updated: $($_.FullName)" -ForegroundColor Green
}

# Update JS files
Write-Host "`n📜 Updating JavaScript files..." -ForegroundColor Yellow
Get-ChildItem -Recurse -Filter "*.js" -Exclude "backup*","vendor*" | ForEach-Object {
    $content = Get-Content $_.FullName -Raw
    $content = $content -replace '/api/branches', '/api/shops' `
                        -replace 'branches', 'shops' `
                        -replace 'Branches', 'Shops'
    Set-Content -Path $_.FullName -Value $content -NoNewline
    Write-Host "  ✓ Updated: $($_.FullName)" -ForegroundColor Green
}

# Remove old files
Write-Host "`n🗑️  Removing old files..." -ForegroundColor Yellow
if (Test-Path "internal/db/branches.go") {
    Remove-Item "internal/db/branches.go" -Force
    Write-Host "  ✓ Removed: internal/db/branches.go" -ForegroundColor Green
}
if (Test-Path "internal/handlers/branches.go") {
    Remove-Item "internal/handlers/branches.go" -Force
    Write-Host "  ✓ Removed: internal/handlers/branches.go" -ForegroundColor Green
}

# Check for remaining references
Write-Host "`n🔍 Checking for remaining references..." -ForegroundColor Yellow
$REMAINING = (Select-String -Path "*.go","*.html","*.js" -Pattern "branches" -Exclude "backup*","vendor*" -Recurse | Where-Object { $_ -notmatch "shops" }).Count

if ($REMAINING -eq 0) {
    Write-Host "✅ No remaining 'branches' references found!" -ForegroundColor Green
} else {
    Write-Host "⚠️  Found $REMAINING remaining references:" -ForegroundColor Red
    Select-String -Path "*.go","*.html","*.js" -Pattern "branches" -Exclude "backup*","vendor*" -Recurse | Where-Object { $_ -notmatch "shops" }
}

Write-Host "`n============================================" -ForegroundColor Cyan
Write-Host "✅ Migration Complete!" -ForegroundColor Green
Write-Host "============================================" -ForegroundColor Cyan
Write-Host "`n📊 Summary:"
Write-Host "  - Backup: $BACKUP_DIR"
Write-Host "  - All files updated"
Write-Host "  - Old files removed"
Write-Host ""
Write-Host "📝 Next Steps:"
Write-Host "  1. Run the database migration script"
Write-Host "  2. Test the application"
Write-Host ""
