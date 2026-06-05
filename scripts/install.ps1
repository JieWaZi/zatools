$ErrorActionPreference = "Stop"

$Owner = if ($env:ZATOOLS_OWNER) { $env:ZATOOLS_OWNER } else { "JieWaZi" }
$Repo = if ($env:ZATOOLS_REPO) { $env:ZATOOLS_REPO } else { "zatools" }
$BinaryName = "zatools"
$Version = if ($env:VERSION) { $env:VERSION } else { "" }
$InstallDir = if ($env:ZATOOLS_INSTALL_DIR) {
    $env:ZATOOLS_INSTALL_DIR
} else {
    Join-Path $env:LOCALAPPDATA "Programs\zatools\bin"
}

function Resolve-BaseUrl {
    if ($Version) {
        return "https://github.com/$Owner/$Repo/releases/download/$Version"
    }

    return "https://github.com/$Owner/$Repo/releases/latest/download"
}

function Resolve-Arch {
    $arch = if ($env:PROCESSOR_ARCHITEW6432) {
        $env:PROCESSOR_ARCHITEW6432
    } else {
        $env:PROCESSOR_ARCHITECTURE
    }

    if ([string]::IsNullOrEmpty($arch)) {
        throw "Unable to determine Windows architecture"
    }

    $arch = $arch.ToUpperInvariant()
    switch ($arch) {
        "AMD64" { return "amd64" }
        "ARM64" { return "arm64" }
        default { throw "Unsupported architecture: $arch" }
    }
}

function Resolve-AssetFromChecksums {
    param(
        [string]$ChecksumsPath,
        [string]$AssetPattern,
        [string]$PreferredAsset
    )

    $matchingAssets = @()
    foreach ($line in Get-Content $ChecksumsPath) {
        $parts = $line -split '\s+'
        if ($parts.Length -lt 2) {
            continue
        }

        $candidate = $parts[1] -replace '^\./', ''
        if ($PreferredAsset -and $candidate -eq $PreferredAsset) {
            return $candidate
        }
        if ($candidate -match $AssetPattern) {
            $matchingAssets += $candidate
        }
    }

    if ($PreferredAsset) {
        throw "Unable to find checksum for $PreferredAsset"
    }
    if ($matchingAssets.Length -eq 0) {
        throw "Unable to find a matching release asset in checksums.txt"
    }
    if ($matchingAssets.Length -gt 1) {
        throw "Multiple matching release assets found: $($matchingAssets -join ', ')"
    }

    return $matchingAssets[0]
}

function Split-PathEntries {
    param(
        [string]$PathValue
    )

    if ([string]::IsNullOrEmpty($PathValue)) {
        return @()
    }

    return $PathValue.Split(';', [System.StringSplitOptions]::RemoveEmptyEntries)
}

$ResolvedArch = Resolve-Arch
$BaseUrl = Resolve-BaseUrl
$PreferredAsset = if ($Version) { "${BinaryName}_${Version}_windows_${ResolvedArch}.tar.gz" } else { "" }
$AssetPattern = "^$([regex]::Escape($BinaryName))_.+_windows_$([regex]::Escape($ResolvedArch))\.tar\.gz$"
$TempDir = Join-Path ([System.IO.Path]::GetTempPath()) ([System.Guid]::NewGuid().ToString())
$ChecksumsPath = Join-Path $TempDir "checksums.txt"

New-Item -ItemType Directory -Force -Path $TempDir | Out-Null

try {
    Invoke-WebRequest -Uri "$BaseUrl/checksums.txt" -OutFile $ChecksumsPath
    $Asset = Resolve-AssetFromChecksums -ChecksumsPath $ChecksumsPath -AssetPattern $AssetPattern -PreferredAsset $PreferredAsset
    $ArchivePath = Join-Path $TempDir $Asset

    Write-Host "Downloading $Asset"
    Invoke-WebRequest -Uri "$BaseUrl/$Asset" -OutFile $ArchivePath

    $expectedHash = $null
    foreach ($line in Get-Content $ChecksumsPath) {
        $parts = $line -split '\s+'
        if ($parts.Length -ge 2) {
            $candidate = $parts[1] -replace '^\./', ''
            if ($candidate -eq $Asset) {
                $expectedHash = $parts[0].ToLowerInvariant()
                break
            }
        }
    }

    if (-not $expectedHash) {
        throw "Unable to find checksum for $Asset"
    }

    $actualHash = (Get-FileHash -Path $ArchivePath -Algorithm SHA256).Hash.ToLowerInvariant()
    if ($actualHash -ne $expectedHash) {
        throw "Checksum verification failed for $Asset"
    }

    tar -xzf $ArchivePath -C $TempDir

    $BinaryPath = Join-Path $TempDir "$BinaryName.exe"
    if (-not (Test-Path $BinaryPath)) {
        throw "Archive did not contain $BinaryName.exe"
    }

    New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
    Copy-Item -Path $BinaryPath -Destination (Join-Path $InstallDir "$BinaryName.exe") -Force

    $UserPath = [Environment]::GetEnvironmentVariable("Path", "User")
    $NormalizedInstallDir = $InstallDir.TrimEnd('\')
    $PathEntries = Split-PathEntries -PathValue $UserPath

    $HasUserPath = $false
    foreach ($entry in $PathEntries) {
        if ($entry.TrimEnd('\') -ieq $NormalizedInstallDir) {
            $HasUserPath = $true
            break
        }
    }

    if (-not $HasUserPath) {
        $NewUserPath = if ($UserPath) { "$UserPath;$InstallDir" } else { $InstallDir }
        [Environment]::SetEnvironmentVariable("Path", $NewUserPath, "User")
    }

    $SessionEntries = Split-PathEntries -PathValue $env:Path
    $HasSessionPath = $false
    foreach ($entry in $SessionEntries) {
        if ($entry.TrimEnd('\') -ieq $NormalizedInstallDir) {
            $HasSessionPath = $true
            break
        }
    }

    if (-not $HasSessionPath) {
        $env:Path = if ($env:Path) { "$InstallDir;$env:Path" } else { $InstallDir }
    }

    Write-Host "Installed $BinaryName to $InstallDir"
    Write-Host "Run: $BinaryName --help"
}
finally {
    if (Test-Path $TempDir) {
        Remove-Item -Recurse -Force $TempDir
    }
}
