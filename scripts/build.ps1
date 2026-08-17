<#
.SYNOPSIS
    Builds linuxcmd and generates one small per-command launcher for every
    registered command.

.DESCRIPTION
    linuxcmd uses a "multicall binary" design (the same technique BusyBox
    uses): there is exactly one compiled Go program (linuxcmd.exe). Every
    per-command executable on PATH (ls.exe, cp.exe, ...) is a hardlink to
    that single binary; at startup it looks at its own invoked filename to
    decide which command to run. Hardlinks share the same on-disk data, so
    installing 20 "separate" executables costs no extra disk space and
    every command starts exactly as fast as the shared binary does.

    The command list is not hardcoded here: it's read from linuxcmd.exe
    itself via `linuxcmd --list-commands`, which reflects the live command
    registry (internal/command). Add a new file to commands/ and this
    script picks it up automatically, no edits required.

.PARAMETER OutDir
    Where to place linuxcmd.exe and the per-command launchers. Defaults to
    .\dist next to this script's repo root.
#>
[CmdletBinding()]
param(
    [string]$OutDir = (Join-Path $PSScriptRoot "..\dist")
)

$ErrorActionPreference = "Stop"
$repoRoot = Join-Path $PSScriptRoot ".."

$OutDir = [System.IO.Path]::GetFullPath($OutDir)
New-Item -ItemType Directory -Force -Path $OutDir | Out-Null

$enginePath = Join-Path $OutDir "linuxcmd.exe"

Write-Host "Building linuxcmd.exe ..." -ForegroundColor Cyan
Push-Location $repoRoot
try {
    & go build -o $enginePath ./cmd/linuxcmd
    if ($LASTEXITCODE -ne 0) {
        throw "go build failed with exit code $LASTEXITCODE"
    }
}
finally {
    Pop-Location
}

Write-Host "Discovering registered commands ..." -ForegroundColor Cyan
$commandNames = & $enginePath --list-commands
if ($LASTEXITCODE -ne 0 -or -not $commandNames) {
    throw "failed to enumerate commands from linuxcmd.exe --list-commands"
}
Write-Host ("Found {0} commands: {1}" -f $commandNames.Count, ($commandNames -join ", "))

# Recreate every per-command launcher from scratch. This matters: rebuilding
# linuxcmd.exe replaces the file rather than writing it in place, which
# would silently detach any existing hardlink from the new content, so
# stale launchers must always be removed and relinked together with it.
foreach ($name in $commandNames) {
    $target = Join-Path $OutDir "$name.exe"
    if (Test-Path $target) {
        Remove-Item $target -Force
    }
    try {
        New-Item -ItemType HardLink -Path $target -Target $enginePath -ErrorAction Stop | Out-Null
    }
    catch {
        # Hardlinks require the same volume; fall back to a plain copy
        # (e.g. OutDir given on a different drive than the temp build dir).
        Copy-Item -Path $enginePath -Destination $target -Force
    }
}

$totalSize = (Get-Item $enginePath).Length
Write-Host ""
Write-Host "Build complete: $OutDir" -ForegroundColor Green
Write-Host ("  linuxcmd.exe + {0} command launchers, {1:N1} MB shared on disk" -f $commandNames.Count, ($totalSize / 1MB))
