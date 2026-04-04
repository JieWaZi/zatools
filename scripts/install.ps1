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

function Resolve-Version {
    if ($Version) {
        return $Version
    }

    $release = Invoke-RestMethod -Uri "https://api.github.com/repos/$Owner/$Repo/releases/latest"
    if (-not $release.tag_name) {
        throw "Unable to resolve latest release version."
    }
    return $release.tag_name
}

function Resolve-Arch {
    $arch = [System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture.ToString()
    switch ($arch) {
        "X64" { return "amd64" }
        "Arm64" { return "arm64" }
        default { throw "Unsupported architecture: $arch" }
    }
}

$ResolvedVersion = Resolve-Version
$ResolvedArch = Resolve-Arch
$Asset = "${BinaryName}_${ResolvedVersion}_windows_${ResolvedArch}.tar.gz"
$BaseUrl = "https://github.com/$Owner/$Repo/releases/download/$ResolvedVersion"
$TempDir = Join-Path ([System.IO.Path]::GetTempPath()) ([System.Guid]::NewGuid().ToString())
$ArchivePath = Join-Path $TempDir $Asset
$ChecksumsPath = Join-Path $TempDir "checksums.txt"

New-Item -ItemType Directory -Force -Path $TempDir | Out-Null

try {
    Write-Host "Downloading $Asset"
    Invoke-WebRequest -Uri "$BaseUrl/$Asset" -OutFile $ArchivePath
    Invoke-WebRequest -Uri "$BaseUrl/checksums.txt" -OutFile $ChecksumsPath

    $expectedHash = $null
    foreach ($line in Get-Content $ChecksumsPath) {
        $parts = $line -split '\s+'
        if ($parts.Length -ge 2) {
            $candidate = $parts[1].TrimStart("./")
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
    $PathEntries = @()
    if ($UserPath) {
        $PathEntries = $UserPath.Split(';', [System.StringSplitOptions]::RemoveEmptyEntries)
    }

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

    $SessionEntries = $env:Path.Split(';', [System.StringSplitOptions]::RemoveEmptyEntries)
    $HasSessionPath = $false
    foreach ($entry in $SessionEntries) {
        if ($entry.TrimEnd('\') -ieq $NormalizedInstallDir) {
            $HasSessionPath = $true
            break
        }
    }

    if (-not $HasSessionPath) {
        $env:Path = "$InstallDir;$env:Path"
    }

    Write-Host "Installed $BinaryName to $InstallDir"
    Write-Host "Run: $BinaryName --help"
}
finally {
    if (Test-Path $TempDir) {
        Remove-Item -Recurse -Force $TempDir
    }
}
