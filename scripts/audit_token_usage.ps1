# ==============================================================================
# scripts/audit_token_usage.ps1 - Auditor Local de Tokens Zero-Cost (0 Tokens de API)
# Analiza la sesion actual del agente, calcula tokens reales/estimados y emite
# recomendaciones proactivas de ahorro y eficiencia para el usuario y la IA.
# ==============================================================================

param(
    [string]$ConversationId = "",
    [switch]$Detailed
)

$ErrorActionPreference = "Stop"

# 1. Localizar directorio de AppData de Antigravity
$appDataDir = Join-Path $env:USERPROFILE ".gemini\antigravity-ide"
if (-not (Test-Path $appDataDir)) {
    Write-Host "[✗] No se encontro el directorio de Antigravity en: $appDataDir" -ForegroundColor Red
    exit 1
}

# 2. Si no se especifico ConversationId, detectar la sesion mas reciente en brain
if ($ConversationId -eq "") {
    $brainDir = Join-Path $appDataDir "brain"
    $latest = Get-ChildItem -Path $brainDir -Directory | Sort-Object LastWriteTime -Descending | Select-Object -First 1
    if ($null -eq $latest) {
        Write-Host "[✗] No se encontraron sesiones en: $brainDir" -ForegroundColor Red
        exit 1
    }
    $ConversationId = $latest.Name
}

$transcriptPath = Join-Path $appDataDir "brain\$ConversationId\.system_generated\logs\transcript.jsonl"
if (-not (Test-Path $transcriptPath)) {
    Write-Host "[✗] No se encontro el transcript en: $transcriptPath" -ForegroundColor Red
    exit 1
}

Write-Host "================================================================================" -ForegroundColor Cyan
Write-Host "         AUDITORIA LOCAL DE CONSUMO DE TOKENS IPVN7 (ZERO-API-COST)            " -ForegroundColor Cyan
Write-Host "================================================================================" -ForegroundColor Cyan
Write-Host "  Sesion ID       : $ConversationId" -ForegroundColor Gray
Write-Host "  Archivo Fuente  : $transcriptPath" -ForegroundColor Gray

# 3. Procesar lineas del transcript
$lines = Get-Content -Path $transcriptPath
$totalSteps = $lines.Count

$userTurns = 0
$assistantTurns = 0
$toolSteps = 0
$totalInputChars = 0
$totalOutputChars = 0

$stepData = @()
$runningContextChars = 20000 * 3.5 # Estimacion base del System Prompt + Tool Definitions (~20k tokens)

foreach ($line in $lines) {
    if ([string]::IsNullOrWhiteSpace($line)) { continue }
    try {
        $entry = $line | ConvertFrom-Json
    } catch {
        continue
    }

    $stepType = $entry.type
    $contentLen = 0
    if ($null -ne $entry.content) {
        $contentLen = $entry.content.Length
    }

    # Tokenizer factor aproximado (Espanol + Codigo estructurado: ~3.2 chars/token)
    $estTokens = [math]::Round($contentLen / 3.2)

    if ($stepType -eq "USER_INPUT") {
        $userTurns++
        $runningContextChars += $contentLen
    } elseif ($stepType -eq "PLANNER_RESPONSE") {
        $assistantTurns++
        $totalOutputChars += $contentLen
        $runningContextChars += $contentLen
    } elseif ($stepType -eq "TOOL_EXECUTION_OUTPUT") {
        $toolSteps++
        $runningContextChars += $contentLen
    } else {
        $runningContextChars += $contentLen
    }

    $promptTokensAtStep = [math]::Round($runningContextChars / 3.2)

    $stepData += [PSCustomObject]@{
        Index       = $entry.step_index
        Type        = $stepType
        Chars       = $contentLen
        Tokens      = $estTokens
        CumulPrompt = $promptTokensAtStep
    }
}

$totalOutputTokens = [math]::Round($totalOutputChars / 3.2)
$currentContextWindowTokens = [math]::Round($runningContextChars / 3.2)

# Costo aproximado basado en tarifas estandar Gemini 1.5/2.0 Flash
# Input: $0.075 por millon | Output: $0.30 por millon
$costInputUSD = ($currentContextWindowTokens * $assistantTurns * 0.075) / 1000000.0
$costOutputUSD = ($totalOutputTokens * 0.30) / 1000000.0
$totalEstimatedCostUSD = $costInputUSD + $costOutputUSD

Write-Host ""
Write-Host "--- RESUMEN METRICO DE LA SESION ---" -ForegroundColor Yellow
Write-Host "  Pasos Totales Registrados : $totalSteps"
Write-Host "  Turnos de Usuario         : $userTurns"
Write-Host "  Respuestas Asistente (IA) : $assistantTurns"
Write-Host "  Llamadas/Salidas Tools    : $toolSteps"
Write-Host "  Tamano Ventana Contexto   : ~$currentContextWindowTokens tokens" -ForegroundColor Green
Write-Host "  Tokens Generados (Salida) : ~$totalOutputTokens tokens" -ForegroundColor Green
Write-Host "  Costo Acumulado Estimado  : `$$([math]::Round($totalEstimatedCostUSD, 4)) USD" -ForegroundColor Magenta

# 4. Deteccion de Picos y Fugas de Contexto
$spikes = $stepData | Where-Object { $_.Chars -gt 15000 } | Sort-Object Chars -Descending

Write-Host ""
Write-Host '--- DETECCION DE PICOS DE CONSUMO (>15,000 caracteres en 1 paso) ---' -ForegroundColor Yellow
if ($spikes.Count -gt 0) {
    Write-Host "  Se detectaron $($spikes.Count) inyecciones masivas que incrementan la ventana de contexto:" -ForegroundColor Red
    foreach ($sp in $spikes | Select-Object -First 5) {
        Write-Host "    * Paso [$($sp.Index)] ($($sp.Type)): $($sp.Chars) chars (~$($sp.Tokens) tokens)" -ForegroundColor DarkYellow
    }
} else {
    Write-Host '  [OK] Cero picos masivos detectados. El flujo de lectura es limpio y controlado.' -ForegroundColor Green
}

# 5. Recomendaciones Proactivas de Eficiencia
Write-Host ""
Write-Host '--- RECOMENDACIONES DE OPTIMIZACION PROACTIVA ---' -ForegroundColor Yellow
if ($currentContextWindowTokens -gt 150000) {
    Write-Host '  [!] Ventana de contexto muy alta (>150k tokens). Cada turno ahora cuesta mas tokens de entrada.' -ForegroundColor DarkYellow
    Write-Host '      Accion: Si cambias radicalmente de tema o tarea, inicia un nuevo chat o permite la compactacion.'
} else {
    Write-Host '  [OK] Ventana de contexto en rango optimo (<150k tokens).' -ForegroundColor Green
}

Write-Host '  [OK] Ejecutar scripts y tests via PowerShell local cuesta 0 TOKENS de API.' -ForegroundColor Green
Write-Host '  [OK] El Demonio Autonomo (scripts/run_autonomous_daemon.ps1) opera con 0 TOKENS.' -ForegroundColor Green
Write-Host '================================================================================' -ForegroundColor Cyan
