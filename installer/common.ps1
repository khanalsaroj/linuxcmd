<#
.SYNOPSIS
    Shared helpers for install.ps1 / uninstall.ps1: reading and writing
    the user's PATH and other environment state safely.

.DESCRIPTION
    Everything here operates at HKCU (per-user) scope only, so install and
    uninstall never require Administrator rights and never touch other
    users' environments or the machine-wide PATH.
#>

function Get-UserPath {
    # Read directly from the registry rather than $env:PATH, which is the
    # current process' merged (user+machine) view and would let a stray
    # write duplicate machine-scope entries into the user scope.
    $val = [Environment]::GetEnvironmentVariable("Path", "User")
    if ($null -eq $val) { return "" }
    return $val
}

function Set-UserPath {
    param([Parameter(Mandatory)][string]$Value)
    [Environment]::SetEnvironmentVariable("Path", $Value, "User")
}

function Add-UserPathEntry {
    param([Parameter(Mandatory)][string]$Entry)
    $current = Get-UserPath
    $parts = $current -split ";" | Where-Object { $_ -ne "" }
    $already = $parts | Where-Object { $_.TrimEnd("\") -ieq $Entry.TrimEnd("\") }
    if ($already) {
        Write-Host "PATH already contains $Entry"
        return $false
    }
    $newParts = $parts + $Entry
    Set-UserPath -Value ($newParts -join ";")
    Write-Host "Added $Entry to user PATH"
    return $true
}

function Remove-UserPathEntry {
    param([Parameter(Mandatory)][string]$Entry)
    $current = Get-UserPath
    $parts = $current -split ";" | Where-Object { $_ -ne "" }
    $kept = $parts | Where-Object { $_.TrimEnd("\") -ine $Entry.TrimEnd("\") }
    if ($kept.Count -eq $parts.Count) {
        return $false
    }
    Set-UserPath -Value ($kept -join ";")
    Write-Host "Removed $Entry from user PATH"
    return $true
}

function Set-UserEnvVar {
    param([Parameter(Mandatory)][string]$Name, [Parameter(Mandatory)][string]$Value)
    [Environment]::SetEnvironmentVariable($Name, $Value, "User")
}

function Remove-UserEnvVar {
    param([Parameter(Mandatory)][string]$Name)
    [Environment]::SetEnvironmentVariable($Name, $null, "User")
}

# Broadcasts WM_SETTINGCHANGE so components that cache the environment
# (notably Explorer) pick up the PATH/env var change without the user
# having to sign out and back in. New CMD windows started via Explorer
# after this runs will see the change; windows already open will not
# (that limitation is inherent to how Windows propagates environment
# changes and is documented in the README).
function Broadcast-EnvironmentChange {
    $sig = @"
using System;
using System.Runtime.InteropServices;
public static class NativeEnv {
    [DllImport("user32.dll", SetLastError = true, CharSet = CharSet.Auto)]
    public static extern IntPtr SendMessageTimeout(
        IntPtr hWnd, uint Msg, UIntPtr wParam, string lParam,
        uint fuFlags, uint uTimeout, out UIntPtr lpdwResult);
}
"@
    if (-not ("NativeEnv" -as [type])) {
        Add-Type -TypeDefinition $sig -ErrorAction SilentlyContinue | Out-Null
    }
    $HWND_BROADCAST = [IntPtr]0xffff
    $WM_SETTINGCHANGE = 0x001A
    $SMTO_ABORTIFHUNG = 0x0002
    $result = [UIntPtr]::Zero
    [NativeEnv]::SendMessageTimeout($HWND_BROADCAST, $WM_SETTINGCHANGE, [UIntPtr]::Zero, "Environment", $SMTO_ABORTIFHUNG, 5000, [ref]$result) | Out-Null
}

# The AutoRun value can already hold the user's own customizations (e.g.
# a custom prompt script), so the doskey hook is appended/removed as one
# recognizable segment rather than overwriting the whole value.
$script:AutoRunMarker = 'doskey /macrofile="%LINUXCMD_HOME%\linuxcmd.doskey"'
$script:AutoRunRegPath = "HKCU:\Software\Microsoft\Command Processor"

function Add-AutoRunHook {
    if (-not (Test-Path $script:AutoRunRegPath)) {
        New-Item -Path $script:AutoRunRegPath -Force | Out-Null
    }
    $existing = (Get-ItemProperty -Path $script:AutoRunRegPath -Name AutoRun -ErrorAction SilentlyContinue).AutoRun
    if ($existing -and $existing.Contains($script:AutoRunMarker)) {
        Write-Host "AutoRun doskey hook already installed"
        return
    }
    $newValue = if ([string]::IsNullOrWhiteSpace($existing)) {
        $script:AutoRunMarker
    } else {
        "$existing & $script:AutoRunMarker"
    }
    New-ItemProperty -Path $script:AutoRunRegPath -Name AutoRun -Value $newValue -PropertyType String -Force | Out-Null
    Write-Host "Installed AutoRun doskey hook (new CMD windows will load it)"
}

function Remove-AutoRunHook {
    $existing = (Get-ItemProperty -Path $script:AutoRunRegPath -Name AutoRun -ErrorAction SilentlyContinue).AutoRun
    if (-not $existing -or -not $existing.Contains($script:AutoRunMarker)) {
        return
    }
    # Strip our marker plus a single adjoining " & " separator on either side.
    $stripped = $existing.Replace(" & $script:AutoRunMarker", "").Replace("$script:AutoRunMarker & ", "").Replace($script:AutoRunMarker, "")
    if ([string]::IsNullOrWhiteSpace($stripped)) {
        Remove-ItemProperty -Path $script:AutoRunRegPath -Name AutoRun -ErrorAction SilentlyContinue
    } else {
        Set-ItemProperty -Path $script:AutoRunRegPath -Name AutoRun -Value $stripped
    }
    Write-Host "Removed AutoRun doskey hook"
}
