<#
.SYNOPSIS
    Installs linuxcmd: a Linux command compatibility layer for Windows CMD.

.DESCRIPTION
    - Builds (or reuses a prebuilt) linuxcmd.exe and one launcher per
      registered command, installed to a per-user directory (no admin
      rights required).
    - Adds that directory to the current user's PATH.
    - Does NOT modify cmd.exe itself.
    - Everything this script does is reversed by uninstall.ps1.

.PARAMETER InstallDir
    Where to install. Defaults to $env:LOCALAPPDATA\Programs\LinuxCmd.

.PARAMETER EnableDoskeyOverrides
    Opt-in. Also registers a DOSKEY macro layer (via the standard HKCU
    "AutoRun" extension point, not a modification of cmd.exe) so that
    cd, mkdir, rmdir and echo work when typed bare at an interactive
    prompt too, not just as Name.exe. See README.md for why this is
    needed and what it changes. Off by default because it's a small,
    persistent, every-new-CMD-window behavior change and should be an
    explicit choice.

.EXAMPLE
    .\install.ps1
    .\install.ps1 -EnableDoskeyOverrides
    .\install.ps1 -InstallDir "D:\Tools\LinuxCmd" -EnableDoskeyOverrides
#>
[CmdletBinding()]
param(
    [string]$InstallDir = (Join-Path $env:LOCALAPPDATA "Programs\LinuxCmd"),
    [switch]$EnableDoskeyOverrides
)

$ErrorActionPreference = "Stop"
$repoRoot = Join-Path $PSScriptRoot ".."
. (Join-Path $PSScriptRoot "common.ps1")

Write-Host "Installing linuxcmd to $InstallDir" -ForegroundColor Cyan
New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null

# --- Get linuxcmd.exe -------------------------------------------------
# Prefer a prebuilt engine sitting next to the installer (how a packaged
# release ships); fall back to building from source (the dev workflow),
# which requires Go on PATH.
$prebuilt = Join-Path $repoRoot "dist\linuxcmd.exe"
$enginePath = Join-Path $InstallDir "linuxcmd.exe"

if (Test-Path $prebuilt) {
    Write-Host "Using prebuilt engine: $prebuilt"
    Copy-Item -Path $prebuilt -Destination $enginePath -Force
}
else {
    if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
        throw "No prebuilt binary found at $prebuilt and Go is not on PATH. Install Go, or run scripts\build.ps1 first."
    }
    Write-Host "No prebuilt binary found; building from source ..."
    Push-Location $repoRoot
    try {
        & go build -ldflags "-X linuxcmd/internal/version.Version=dev" -o $enginePath ./cmd/linuxcmd
        if ($LASTEXITCODE -ne 0) { throw "go build failed with exit code $LASTEXITCODE" }
    }
    finally {
        Pop-Location
    }
}

# --- Generate per-command launchers ------------------------------------
Write-Host "Generating command launchers ..."
$commandNames = & $enginePath --list-commands
if ($LASTEXITCODE -ne 0 -or -not $commandNames) {
    throw "failed to enumerate commands from linuxcmd.exe --list-commands"
}
foreach ($name in $commandNames) {
    $target = Join-Path $InstallDir "$name.exe"
    if (Test-Path $target) { Remove-Item $target -Force }
    try {
        New-Item -ItemType HardLink -Path $target -Target $enginePath -ErrorAction Stop | Out-Null
    }
    catch {
        Copy-Item -Path $enginePath -Destination $target -Force
    }
}
Write-Host ("Installed {0} commands: {1}" -f $commandNames.Count, ($commandNames -join ", "))

# --- PATH ---------------------------------------------------------------
$pathChanged = Add-UserPathEntry -Entry $InstallDir

# --- Optional DOSKEY/AutoRun layer --------------------------------------
Set-UserEnvVar -Name "LINUXCMD_HOME" -Value $InstallDir
$doskeySrc = Join-Path $PSScriptRoot "linuxcmd.doskey"
$doskeyDst = Join-Path $InstallDir "linuxcmd.doskey"
Copy-Item -Path $doskeySrc -Destination $doskeyDst -Force

if ($EnableDoskeyOverrides) {
    Add-AutoRunHook
}
else {
    Write-Host "Skipping DOSKEY/AutoRun layer (bare 'cd', 'mkdir', 'rmdir', 'echo' will keep using cmd.exe's builtins)."
    Write-Host "Re-run with -EnableDoskeyOverrides to enable it."
}

Broadcast-EnvironmentChange

Write-Host ""
Write-Host "Install complete." -ForegroundColor Green
Write-Host "Open a NEW Command Prompt window and try:  ls -la"
if ($pathChanged) {
    Write-Host "(PATH was just updated; windows already open won't see it until reopened.)"
}
if (-not $EnableDoskeyOverrides) {
    Write-Host "cd, mkdir, rmdir and echo will still resolve to cmd.exe's own builtins unless you re-run with -EnableDoskeyOverrides."
}
