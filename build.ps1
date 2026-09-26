[CmdletBinding()]
param([ValidateSet('all','font','rgb')][string]$Module = 'all')

$ErrorActionPreference = 'Stop'
$ProjectRoot = Split-Path -Parent $MyInvocation.MyCommand.Path
$Dist = Join-Path $ProjectRoot 'dist'
New-Item -ItemType Directory -Force -Path $Dist | Out-Null

function Get-ModuleVersion([string]$ModuleDir) {
    $Prop = Get-Content (Join-Path $ModuleDir 'module.prop')
    $Line = $Prop | Where-Object { $_ -match '^version=' } | Select-Object -First 1
    if (-not $Line) { throw "Missing version in $ModuleDir/module.prop" }
    return ($Line -replace '^version=', '').Trim()
}

function Compress-Module([string]$ModuleDir, [string]$OutputFile) {
    if (Test-Path -LiteralPath $OutputFile) { Remove-Item -LiteralPath $OutputFile -Force }
    $Items = Get-ChildItem -LiteralPath $ModuleDir | Where-Object { $_.Name -ne 'src' }
    # Android ZIP entries must use forward slashes, including on Windows.
    Add-Type -AssemblyName System.IO.Compression
    $Stream = [IO.File]::Open($OutputFile, [IO.FileMode]::CreateNew)
    $Archive = [IO.Compression.ZipArchive]::new($Stream, [IO.Compression.ZipArchiveMode]::Create)
    try {
        foreach ($Item in $Items) {
            $Files = if ($Item.PSIsContainer) { Get-ChildItem $Item.FullName -Recurse -File } else { $Item }
            foreach ($File in $Files) {
                $Name = $File.FullName.Substring($ModuleDir.Length + 1).Replace('\', '/')
                $Entry = $Archive.CreateEntry($Name)
                $InputStream = $File.OpenRead()
                $OutputStream = $Entry.Open()
                try { $InputStream.CopyTo($OutputStream) } finally { $InputStream.Dispose(); $OutputStream.Dispose() }
            }
        }
    } finally { $Archive.Dispose(); $Stream.Dispose() }
}

if ($Module -in @('all','font')) {
    $FontDir = Join-Path $ProjectRoot 'font'
    $Version = Get-ModuleVersion $FontDir
    $Output = Join-Path $Dist "KPA_MYuppy_Font_v$Version.zip"
    Compress-Module $FontDir $Output
    Write-Host "Built $Output"
}

if ($Module -in @('all','rgb')) {
    $RgbDir = Join-Path $ProjectRoot 'rgb'
    $Version = Get-ModuleVersion $RgbDir
    $BinDir = Join-Path $RgbDir 'bin'
    New-Item -ItemType Directory -Force -Path $BinDir | Out-Null
    $env:GOOS = 'android'
    $env:GOARCH = 'arm64'
    $env:CGO_ENABLED = '0'
    & go build -trimpath -ldflags='-s -w' -o (Join-Path $BinDir 'kpa_rgb_daemon') (Join-Path $RgbDir 'src/main.go')
    if ($LASTEXITCODE -ne 0) { throw 'Go build failed.' }
    $Output = Join-Path $Dist "KPA_RGB_Control_v$Version.zip"
    Compress-Module $RgbDir $Output
    Write-Host "Built $Output"
}
