# linuxcmd installer for Windows (PowerShell).
#
#   iwr -useb https://raw.githubusercontent.com/khanalsaroj/linuXwin/main/main/install.ps1 | iex
#
# Downloads the latest (or pinned) release archive, verifies its checksum,
# and hands off to the repo's own installer/install.ps1 -- which does the
# real work (per-command launcher generation, PATH, optional DOSKEY layer)
# from the archive's dist/ + installer/ layout.
#
# Environment overrides (piping into `iex` can't take script parameters,
# so configuration goes through env vars instead, same as the args a
# normal `installer\install.ps1` call would take):
#   $env:LINUXCMD_VERSION          install a specific version (e.g. v1.2.3), default: latest
#   $env:LINUXCMD_INSTALL_DIR      install location, default: %LOCALAPPDATA%\Programs\LinuxCmd
#   $env:LINUXCMD_ENABLE_DOSKEY    set to "1" to also enable the cd/mkdir/rmdir/echo DOSKEY layer
$ErrorActionPreference = "Stop"

$Repo = "khanalsaroj/linuXwin"
$BinName = "linuxcmd"
$InstallDir = if ($env:LINUXCMD_INSTALL_DIR) { $env:LINUXCMD_INSTALL_DIR } else { "$env:LOCALAPPDATA\Programs\LinuxCmd" }

function Info($m) { Write-Host "  $m" }
function Ok($m)   { Write-Host "  $m" -ForegroundColor Green }
function Warn($m) { Write-Host "  $m" -ForegroundColor Yellow }
function Die($m)  { Write-Host "  $m" -ForegroundColor Red; exit 1 }

Write-Host ""
Write-Host "  linuxcmd installer" -ForegroundColor Cyan
Write-Host "  a Linux command compatibility layer for Windows CMD"
Write-Host ""

# ---------- Architecture ----------
$Arch = switch ($env:PROCESSOR_ARCHITECTURE) {
    "AMD64" { "amd64" }
    "ARM64" { "arm64" }
    "x86"   { "386" }
    default { if ([Environment]::Is64BitOperatingSystem) { "amd64" } else { "386" } }
}

# ---------- Version ----------
if ($env:LINUXCMD_VERSION) {
    $Version = $env:LINUXCMD_VERSION.TrimStart("v")
} else {
    try {
        $Version = (Invoke-RestMethod "https://api.github.com/repos/$Repo/releases/latest").tag_name.TrimStart("v")
    } catch {
        Die "Could not resolve the latest release: $_"
    }
}
Info "Installing $BinName v$Version for windows/$Arch"

$Asset = "$BinName-windows-$Arch.zip"
$Url = "https://github.com/$Repo/releases/download/v$Version/$Asset"

$Tmp = Join-Path $env:TEMP ("linuxcmd-" + [System.Guid]::NewGuid().ToString())
New-Item -ItemType Directory -Force -Path $Tmp | Out-Null
$Zip = Join-Path $Tmp $Asset

Info "Downloading $Url"
try {
    Invoke-WebRequest -Uri $Url -OutFile $Zip -UseBasicParsing
} catch {
    Die "Download failed - does a release exist for windows/$Arch? ($_)"
}

# ---------- Checksum verification (best effort) ----------
try {
    $SumsUrl = "https://github.com/$Repo/releases/download/v$Version/checksums.txt"
    $SumsResponse = Invoke-WebRequest -Uri $SumsUrl -UseBasicParsing
    # GitHub serves release assets (including checksums.txt) as
    # application/octet-stream, so -UseBasicParsing hands back .Content as
    # a raw byte[] rather than a decoded string; splitting a byte array by
    # line silently produces garbage instead of an error, so decode it
    # explicitly rather than assuming .Content is already text.
    $Sums = if ($SumsResponse.Content -is [byte[]]) {
        [System.Text.Encoding]::UTF8.GetString($SumsResponse.Content)
    } else {
        $SumsResponse.Content
    }
    $Line = ($Sums -split "`n") | Where-Object { $_ -match [regex]::Escape($Asset) } | Select-Object -First 1
    if ($Line) {
        $Expected = ($Line.Trim() -split "\s+")[0]
        $Actual = (Get-FileHash -Algorithm SHA256 -Path $Zip).Hash
        if ($Expected -and ($Expected.ToLower() -ne $Actual.ToLower())) {
            Die "Checksum mismatch for $Asset"
        }
        Ok "Checksum verified"
    } else {
        Warn "No checksum entry for $Asset - skipping verification"
    }
} catch {
    Warn "Skipping checksum verification ($_)"
}

# ---------- Extract ----------
Expand-Archive -Force -Path $Zip -DestinationPath $Tmp

# ---------- Hand off to the real installer ----------
# The release archive mirrors the repo's dist/ + installer/ layout (see
# .github/workflows/release.yml), so the richer local installer --
# per-command hardlink generation, PATH management, the optional
# DOSKEY/AutoRun layer -- just runs from the extracted copy.
$RealInstaller = Join-Path $Tmp "installer\install.ps1"
if (-not (Test-Path $RealInstaller)) {
    Die "installer\install.ps1 not found in the downloaded archive"
}

$InstallArgs = @{ InstallDir = $InstallDir }
if ($env:LINUXCMD_ENABLE_DOSKEY -eq "1" -or $env:LINUXCMD_ENABLE_DOSKEY -eq "true") {
    $InstallArgs["EnableDoskeyOverrides"] = $true
}
& $RealInstaller @InstallArgs

Remove-Item -Recurse -Force $Tmp -ErrorAction SilentlyContinue
