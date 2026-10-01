# ==============================================================================
# build_hardened.ps1 - Compilacion Blindada y Ofuscada IPVN7 - Rol N
# Aplica ofuscacion de AST, cifrado estatico de literales y stripping DWARF
# ==============================================================================

param(
    [string]$TargetOS = "windows",
    [string]$TargetArch = "amd64",
    [switch]$DeepObfuscate,
    [switch]$AllTargets,
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
Write-Host "  IPVN7 - COMPILACION BLINDADA Y OFUSCACION (ROL N)" -ForegroundColor Cyan
Write-Host "================================================================" -ForegroundColor Cyan

# 1. Detectar Garble si se solicita ofuscacion profunda (-DeepObfuscate)
$garbleCmd = "garble"
$garbleAvailable = $false

if ($DeepObfuscate) {
    if (Get-Command "garble" -ErrorAction SilentlyContinue) {
        $garbleAvailable = $true
    } else {
        $goPath = if ($env:GOPATH) { $env:GOPATH } else { Join-Path $HOME "go" }
        $candidate = Join-Path $goPath "bin\garble.exe"
        if (Test-Path $candidate) {
            $garbleCmd = $candidate
            $garbleAvailable = $true
        }
    }

    if ($garbleAvailable) {
        Write-Host "[+] Modo Ofuscacion Profunda: GARBLE activo" -ForegroundColor Green
        Write-Host "    Banderas: -literals (strings cifrados), -tiny (poda tipos)" -ForegroundColor DarkGray
        Write-Host "    [NOTA] Requiere exclusion en Windows Defender para evitar falsos positivos." -ForegroundColor Yellow
    } else {
        Write-Host "[!] ADVERTENCIA: Garble no detectado en PATH/GOPATH." -ForegroundColor Yellow
        Write-Host "    Fallback: Compilando con Stripping Nativo Go (-trimpath -ldflags='-s -w')." -ForegroundColor Yellow
    }
} else {
    Write-Host "[+] Modo: Blindaje de Produccion con Stripping DWARF Limpio" -ForegroundColor Green
    Write-Host "    Banderas: -trimpath (sin rutas locales) y -ldflags='-s -w' (sin simbolos/debug)" -ForegroundColor DarkGray
    Write-Host "    (Para ofuscacion extrema de strings usa el flag -DeepObfuscate)" -ForegroundColor DarkGray
}

$targets = @()
if ($AllTargets) {
    $targets = @(
        @{ OS = "windows"; Arch = "amd64"; Output = "ipvn7-hardened-windows-amd64.exe" },
        @{ OS = "windows"; Arch = "arm64"; Output = "ipvn7-hardened-windows-arm64.exe" },
        @{ OS = "linux";   Arch = "amd64"; Output = "ipvn7-hardened-linux-amd64" },
        @{ OS = "linux";   Arch = "arm64"; Output = "ipvn7-hardened-linux-arm64" },
        @{ OS = "darwin";  Arch = "amd64"; Output = "ipvn7-hardened-darwin-amd64" },
        @{ OS = "darwin";  Arch = "arm64"; Output = "ipvn7-hardened-darwin-arm64" }
    )
} else {
    $ext = if ($TargetOS -eq "windows") { ".exe" } else { "" }
    $targets = @(
        @{ OS = $TargetOS; Arch = $TargetArch; Output = "ipvn7-hardened$ext" }
    )
}

$ldflags = "-s -w"

Push-Location $SrcDir
try {
    foreach ($t in $targets) {
        $outFile = Join-Path $BinDir $t.Output
        $targetDesc = "[-] Compilando blindado " + $t.OS + "/" + $t.Arch + " -> " + $t.Output + "..."
        Write-Host $targetDesc -NoNewline

        $env:GOOS = $t.OS
        $env:GOARCH = $t.Arch
        $env:CGO_ENABLED = "0"

        $buildOut = ""
        $oldEAP = $ErrorActionPreference
        $ErrorActionPreference = "Continue"
        try {
            if ($garbleAvailable) {
                $buildOut = & $garbleCmd -literals -tiny -seed=random build -trimpath -ldflags "$ldflags" -o $outFile ./cmd/ipvn7 2>&1
            } else {
                $buildOut = & go build -trimpath -ldflags "$ldflags" -o $outFile ./cmd/ipvn7 2>&1
            }
        } finally {
            $ErrorActionPreference = $oldEAP
        }

        if ($LASTEXITCODE -eq 0 -and (Test-Path $outFile)) {
            Write-Host " [OK]" -ForegroundColor Green
        } else {
            Write-Host " [ERROR]" -ForegroundColor Red
            Write-Host $buildOut -ForegroundColor DarkGray
        }
    }

    # Si es el objetivo local de windows, actualizar ipvn7.exe canonico si se solicita
    $localHardened = Join-Path $BinDir "ipvn7-hardened.exe"
    $canonBin = Join-Path $BinDir "ipvn7.exe"
    if (Test-Path $localHardened) {
        Copy-Item $localHardened $canonBin -Force
        Write-Host "[-] Binario canonico bin/ipvn7.exe actualizado con version blindada." -ForegroundColor Green
    }
} finally {
    $env:GOOS = ""
    $env:GOARCH = ""
    $env:CGO_ENABLED = ""
    Pop-Location
}

# 2. Registrar sumas SHA-256
if (-not $SkipHashes) {
    Write-Host "`n[-] Actualizando sumas SHA256 en bin/SHA256SUMS.txt..." -ForegroundColor Cyan
    $hashFile = Join-Path $BinDir "SHA256SUMS.txt"
    $existingHashes = @{}
    if (Test-Path $hashFile) {
        Get-Content $hashFile | ForEach-Object {
            $parts = $_ -split "\s+", 2
            if ($parts.Count -eq 2) { $existingHashes[$parts[1].Trim()] = $parts[0].Trim() }
        }
    }

    foreach ($t in $targets) {
        $p = Join-Path $BinDir $t.Output
        if (Test-Path $p) {
            $h = (Get-FileHash -Path $p -Algorithm SHA256).Hash
            $existingHashes[$t.Output] = $h
        }
    }
    if (Test-Path $canonBin) {
        $existingHashes["ipvn7.exe"] = (Get-FileHash -Path $canonBin -Algorithm SHA256).Hash
    }

    $lines = $existingHashes.Keys | Sort-Object | ForEach-Object { "$($existingHashes[$_])  $_" }
    $lines | Set-Content -Path $hashFile -Encoding utf8
    Write-Host " [OK] Manifiesto SHA256 actualizado con exito." -ForegroundColor Green
}

Write-Host "================================================================" -ForegroundColor Cyan
Write-Host "  BLINDAJE Y EMPAQUETADO COMPLETADO EXITOSAMENTE (ROL N)" -ForegroundColor Cyan
Write-Host "================================================================" -ForegroundColor Cyan
