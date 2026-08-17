<#
.SYNOPSIS
    Uninstalls linuxcmd: reverses everything install.ps1 did.

.PARAMETER InstallDir
    Must match the directory used at install time. Defaults to
    $env:LOCALAPPDATA\Programs\LinuxCmd.

.PARAMETER KeepFiles
    Removes PATH/AutoRun/env var registrations but leaves the installed
    files in place.
#>
[CmdletBinding()]
param(
    [string]$InstallDir = (Join-Path $env:LOCALAPPDATA "Programs\LinuxCmd"),
    [switch]$KeepFiles
)

$ErrorActionPreference = "Stop"
. (Join-Path $PSScriptRoot "common.ps1")

Write-Host "Uninstalling linuxcmd from $InstallDir" -ForegroundColor Cyan

Remove-UserPathEntry -Entry $InstallDir | Out-Null
Remove-AutoRunHook
Remove-UserEnvVar -Name "LINUXCMD_HOME"
Broadcast-EnvironmentChange

if (-not $KeepFiles) {
    if (Test-Path $InstallDir) {
        Remove-Item -Recurse -Force $InstallDir
        Write-Host "Removed $InstallDir"
    }
}
else {
    Write-Host "Left files in place at $InstallDir (-KeepFiles was passed)"
}

Write-Host ""
Write-Host "Uninstall complete." -ForegroundColor Green
Write-Host "Open a NEW Command Prompt window to see the change; windows already open still have the old PATH."
