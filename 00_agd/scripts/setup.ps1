[CmdletBinding()]
param(
    [switch]$SkipCheck,
    [switch]$SkipTemplates,
    [switch]$InstallCiTemplate,
    [switch]$InstallPrTemplate,
    [switch]$NoTemplateBackup,
    [switch]$SlimCheckout
)

$ErrorActionPreference = "Stop"
if (Get-Variable -Name PSNativeCommandUseErrorActionPreference -ErrorAction SilentlyContinue) {
    $PSNativeCommandUseErrorActionPreference = $false
}

function Invoke-Step {
    param(
        [Parameter(Mandatory = $true)][string]$Title,
        [Parameter(Mandatory = $true)][scriptblock]$Action
    )
    Write-Host ""
    Write-Host "==> $Title"
    & $Action
}

function Invoke-External {
    param(
        [Parameter(Mandatory = $true)][string]$FilePath,
        [string[]]$Arguments = @()
    )
    & $FilePath @Arguments
    if ($LASTEXITCODE -ne 0) {
        throw "Command failed: $FilePath $($Arguments -join ' ')"
    }
}

function Invoke-Git {
    param(
        [string[]]$Arguments = @()
    )
    & git @script:GitPrefixArgs @Arguments
    if ($LASTEXITCODE -ne 0) {
        throw "Command failed: git $($Arguments -join ' ')"
    }
}

function Get-GitOutput {
    param(
        [string[]]$Arguments = @()
    )
    $output = & git @script:GitPrefixArgs @Arguments
    if ($LASTEXITCODE -ne 0) {
        throw "Command failed: git $($Arguments -join ' ')"
    }
    return $output
}

function Install-TemplateFile {
    param(
        [Parameter(Mandatory = $true)][string]$SourcePath,
        [Parameter(Mandatory = $true)][string]$DestinationPath,
        [switch]$NoBackup
    )

    if (-not (Test-Path $SourcePath)) {
        throw "Template source not found: $SourcePath"
    }

    $dstDir = Split-Path -Parent $DestinationPath
    if (-not [string]::IsNullOrWhiteSpace($dstDir)) {
        New-Item -ItemType Directory -Force $dstDir | Out-Null
    }

    if (Test-Path $DestinationPath) {
        $srcText = [System.IO.File]::ReadAllText($SourcePath)
        $dstText = [System.IO.File]::ReadAllText($DestinationPath)
        if ($srcText -ceq $dstText) {
            Write-Host ("already up-to-date: " + $DestinationPath)
            return
        }

        if (-not $NoBackup) {
            $timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
            $backupPath = "$DestinationPath.bak.$timestamp"
            $backupDone = $false
            for ($try = 1; $try -le 5; $try++) {
                try {
                    Copy-Item -Force $DestinationPath $backupPath
                    $backupDone = $true
                    break
                } catch [System.IO.IOException] {
                    if ($try -eq 5) { throw }
                    Start-Sleep -Milliseconds 250
                }
            }
            if ($backupDone) {
                Write-Host ("backup created: " + $backupPath)
            }
        }
    }

    $copied = $false
    for ($try = 1; $try -le 5; $try++) {
        try {
            Copy-Item -Force $SourcePath $DestinationPath
            $copied = $true
            break
        } catch [System.IO.IOException] {
            if ($try -eq 5) { throw }
            Start-Sleep -Milliseconds 250
        }
    }
    if ($copied) {
        Write-Host ("installed: " + $DestinationPath)
    }
}

$scriptRoot = Split-Path -Parent $MyInvocation.MyCommand.Path
$packageRoot = Split-Path -Parent $scriptRoot
$packageName = Split-Path -Leaf $packageRoot
$script:GitPrefixArgs = @()
$savedErrorAction = $ErrorActionPreference
$repoRootOutput = $null
$repoRootCode = 1
try {
    $ErrorActionPreference = "Continue"
    $repoRootOutput = cmd /c "git rev-parse --show-toplevel" 2>&1
    $repoRootCode = $LASTEXITCODE
} finally {
    $ErrorActionPreference = $savedErrorAction
}
if ($repoRootCode -ne 0 -or [string]::IsNullOrWhiteSpace(($repoRootOutput | Select-Object -Last 1))) {
    $repoRootOutput = $null
    $repoRootCode = 1
    try {
        $ErrorActionPreference = "Continue"
        $repoRootOutput = cmd /c "git -c safe.directory=* rev-parse --show-toplevel" 2>&1
        $repoRootCode = $LASTEXITCODE
    } finally {
        $ErrorActionPreference = $savedErrorAction
    }
    if ($repoRootCode -ne 0 -or [string]::IsNullOrWhiteSpace(($repoRootOutput | Select-Object -Last 1))) {
        throw "Failed to resolve repository root."
    }
    $script:GitPrefixArgs = @("-c", "safe.directory=*")
}
$repoRoot = ($repoRootOutput | Select-Object -Last 1).Trim()
if ([string]::IsNullOrWhiteSpace($repoRoot)) {
    throw "Failed to resolve repository root."
}
Set-Location $repoRoot

$packageRoot = Join-Path $repoRoot $packageName
$relHooksPath = "$packageName/.githooks"

Invoke-Step "Repository root" {
    Write-Host $repoRoot
}

Invoke-Step "Checking required tools" {
    if (-not (Get-Command git -ErrorAction SilentlyContinue)) {
        throw "git is required but not found in PATH."
    }
}

Invoke-Step "Validating AGD package folder" {
    if (-not (Test-Path $packageRoot)) {
        throw "'$packageName' folder not found at repository root."
    }
}

if ($SlimCheckout) {
    Invoke-Step "Slim checkout (agd only)" {
        $statusLines = Get-GitOutput -Arguments @("status", "--porcelain")
        if (-not [string]::IsNullOrWhiteSpace(($statusLines -join "`n"))) {
            throw "Slim checkout requires a clean working tree. Commit or stash local changes first."
        }
        Invoke-Git -Arguments @("sparse-checkout", "init", "--no-cone")
        Invoke-Git -Arguments @("sparse-checkout", "set", "/$packageName/")
        Invoke-Git -Arguments @("checkout")
    }
}

Invoke-Step "Workspace layout" {
    $sourceDirs = @()
    foreach ($dirName in @("cmd", "internal", "scripts")) {
        $dirPath = Join-Path $repoRoot $dirName
        if (-not (Test-Path $dirPath)) {
            continue
        }
        $hasFiles = Get-ChildItem -Path $dirPath -Recurse -File -ErrorAction SilentlyContinue | Select-Object -First 1
        if ($null -ne $hasFiles) {
            $sourceDirs += $dirName
        }
    }
    if ($sourceDirs.Count -gt 0) {
        Write-Host ("full-source checkout detected: " + ($sourceDirs -join ", "))
        Write-Host "tip: run $packageName\\setup.cmd -SlimCheckout to keep only $packageName/ in this clone."
    } else {
        Write-Host "minimal checkout detected ($packageName-only)."
    }
}

Invoke-Step "Configuring git hooks path" {
    Invoke-Git -Arguments @("config", "core.hooksPath", $relHooksPath)
    $hooksPath = ((Get-GitOutput -Arguments @("config", "--get", "core.hooksPath")) | Select-Object -Last 1).Trim()
    Write-Host "core.hooksPath = $hooksPath"
}

$agdBin = $null
$agdEnPath = Join-Path $packageRoot "agd_en.exe"
$agdDefaultPath = Join-Path $packageRoot "agd.exe"
if (Test-Path $agdEnPath) {
    $agdBin = $agdEnPath
} elseif (Test-Path $agdDefaultPath) {
    $agdBin = $agdDefaultPath
}

if ($null -eq $agdBin) {
    throw "AGD executable not found. Expected: $packageName\\agd_en.exe or $packageName\\agd.exe"
}

if (-not $SkipCheck) {
    Invoke-Step "Running AGD validation checks" {
        $agdDocsRel = "$packageName\agd_docs"
        $examplesRel = "$packageName\examples"
        $agdDocs = Join-Path $repoRoot $agdDocsRel
        $examples = Join-Path $repoRoot $examplesRel

        if (Test-Path $agdDocs) {
            Invoke-External -FilePath $agdBin -Arguments @("check-all", $agdDocsRel, "--strict")
        } else {
            Write-Host "skip: $agdDocsRel not found"
        }

        if (Test-Path $examples) {
            Invoke-External -FilePath $agdBin -Arguments @("check-all", $examplesRel, "--strict")
        } else {
            Write-Host "skip: $examplesRel not found"
        }
    }
} else {
    Write-Host ""
    Write-Host "==> Validation skipped (-SkipCheck)"
}

$ciExplicit = $PSBoundParameters.ContainsKey("InstallCiTemplate")
$prExplicit = $PSBoundParameters.ContainsKey("InstallPrTemplate")
$installCi = $false
$installPr = $false

if (-not $SkipTemplates) {
    if ($ciExplicit -or $prExplicit) {
        $installCi = $InstallCiTemplate
        $installPr = $InstallPrTemplate
    } else {
        # Default behavior: full setup in one command.
        $installCi = $true
        $installPr = $true
    }
}

if ($installCi) {
    Invoke-Step "Installing CI workflow template" {
        $src = Join-Path $packageRoot "templates\agd-guard.yml"
        $dst = Join-Path $repoRoot ".github\workflows\agd-guard.yml"
        Install-TemplateFile -SourcePath $src -DestinationPath $dst -NoBackup:$NoTemplateBackup
    }
}

if ($installPr) {
    Invoke-Step "Installing PR template" {
        $src = Join-Path $packageRoot "templates\pull_request_template.md"
        $dst = Join-Path $repoRoot ".github\pull_request_template.md"
        Install-TemplateFile -SourcePath $src -DestinationPath $dst -NoBackup:$NoTemplateBackup
    }
}

if ($SkipTemplates) {
    Write-Host ""
    Write-Host "==> Template install skipped (-SkipTemplates)"
}

Write-Host ""
Write-Host "Setup complete."
Write-Host "Next commands:"
Write-Host "  $packageName\\agd_en.exe quick"
Write-Host "  $packageName\\agd_en.exe wizard"
Write-Host ""
Write-Host "Optional setup flags:"
Write-Host "  $packageName\\setup.cmd -SkipCheck"
Write-Host "  $packageName\\setup.cmd -SkipTemplates"
Write-Host "  $packageName\\setup.cmd -InstallCiTemplate"
Write-Host "  $packageName\\setup.cmd -InstallPrTemplate"
Write-Host "  $packageName\\setup.cmd -NoTemplateBackup"
Write-Host "  $packageName\\setup.cmd -SlimCheckout"
