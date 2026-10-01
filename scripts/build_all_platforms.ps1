# ==============================================================================
# build_all_platforms.ps1 - Compilador Multiplataforma Universal IPVN7
# Genera binarios estaticos independientes (CGO_ENABLED=0) para Windows, Linux y macOS
# ==============================================================================

param(
    [switch]$SkipHashes
)

$ErrorActionPreference = "Stop"
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$RepoRoot = Split-Path -Parent $ScriptDir
$SrcDir = Join-Path $RepoRoot "src"
$BinDir = Join-Path $RepoRoot "bin"

if (-not (Test-Path $BinDir)) {
    New-Item -ItemType Directory -Path $BinDir -Force | Out-Null
}

Write-Host "================================================================" -ForegroundColor Cyan
Write-Host "  IPVN7 - COMPILACION MULTIPLATAFORMA UNIVERSAL - ROL K" -ForegroundColor Cyan
Write-Host "================================================================" -ForegroundColor Cyan

$targets = @(
    @{ OS = "windows"; Arch = "amd64"; Output = "ipvn7-windows-amd64.exe" },
    @{ OS = "windows"; Arch = "arm64"; Output = "ipvn7-windows-arm64.exe" },
    @{ OS = "linux";   Arch = "amd64"; Output = "ipvn7-linux-amd64" },
    @{ OS = "linux";   Arch = "arm64"; Output = "ipvn7-linux-arm64" },
    @{ OS = "darwin";  Arch = "amd64"; Output = "ipvn7-darwin-amd64" },
    @{ OS = "darwin";  Arch = "arm64"; Output = "ipvn7-darwin-arm64" }
)

$ldflags = "-s -w"

Push-Location $SrcDir
try {
    foreach ($t in $targets) {
        $outFile = Join-Path $BinDir $t.Output
        $targetDesc = "[-] Compilando " + $t.OS + "/" + $t.Arch + " -> " + $t.Output + "..."
        Write-Host $targetDesc -NoNewline
        
        $env:GOOS = $t.OS
        $env:GOARCH = $t.Arch
        $env:CGO_ENABLED = "0"
        
        $buildOut = & go build -trimpath -ldflags "$ldflags" -o $outFile ./cmd/ipvn7 2>&1
        if ($LASTEXITCODE -eq 0 -and (Test-Path $outFile)) {
            Write-Host " [OK]" -ForegroundColor Green
        } else {
            Write-Host " [ERROR]" -ForegroundColor Red
            Write-Host $buildOut -ForegroundColor DarkGray
        }
    }

    # Binario canonico local de conveniencia
    $localBin = Join-Path $BinDir "ipvn7.exe"
    if (Test-Path (Join-Path $BinDir "ipvn7-windows-amd64.exe")) {
        Copy-Item (Join-Path $BinDir "ipvn7-windows-amd64.exe") $localBin -Force
    }

} finally {
    $env:GOOS = ""
    $env:GOARCH = ""
    $env:CGO_ENABLED = ""
    Pop-Location
}

# Generar manifiesto SHA256
if (-not $SkipHashes) {
    Write-Host "`n[-] Calculando sumas de verificacion SHA256..." -ForegroundColor Cyan
    $hashList = @()
    $hashFile = Join-Path $BinDir "SHA256SUMS.txt"
    
    foreach ($t in $targets) {
        $p = Join-Path $BinDir $t.Output
        if (Test-Path $p) {
            $h = Get-FileHash -Path $p -Algorithm SHA256
            $hashList += ($h.Hash + "  " + $t.Output)
        }
    }
    
    $hashList | Set-Content -Path $hashFile -Encoding utf8
    Write-Host " [OK] Manifiesto guardado en bin/SHA256SUMS.txt" -ForegroundColor Green
}

Write-Host "================================================================" -ForegroundColor Cyan
Write-Host "  COMPILACION COMPLETADA EXITOSAMENTE (0 DEPENDENCIAS)" -ForegroundColor Cyan
Write-Host "================================================================" -ForegroundColor Cyan
