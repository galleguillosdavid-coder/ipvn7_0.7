<#
.SYNOPSIS
    gobierno.ps1 - Wrapper PowerShell para el controlador de gobierno IPVN7
.EXAMPLE
    .\sistema\bin\gobierno.ps1 status
    .\sistema\bin\gobierno.ps1 validate
    .\sistema\bin\gobierno.ps1 verify
    .\sistema\bin\gobierno.ps1 close MI_ID
#>
param(
    [Parameter(Position=0)]
    [string]$Command = "status",

    [Parameter(Position=1)]
    [string]$Id = "AUTO"
)

$RepoRoot = Split-Path -Parent (Split-Path -Parent $PSScriptRoot)
$ScriptPy = Join-Path $PSScriptRoot "gobierno.py"

python $ScriptPy $Command $Id
exit $LASTEXITCODE
