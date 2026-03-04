$ErrorActionPreference = "Stop"

$scriptRoot = Split-Path -Parent $MyInvocation.MyCommand.Path
$packageRoot = Split-Path -Parent (Split-Path -Parent $scriptRoot)
$packageName = Split-Path -Leaf $packageRoot

$repoRoot = (git rev-parse --show-toplevel).Trim()
if ([string]::IsNullOrWhiteSpace($repoRoot)) {
    Write-Error "[AGD hook] failed to resolve repository root."
}
Set-Location $repoRoot

function Normalize-RepoRelPath {
    param([string]$PathValue)

    $value = [string]$PathValue
    $value = $value.Trim()
    if ([string]::IsNullOrWhiteSpace($value)) {
        return ""
    }
    $value = $value.Replace('\', '/')
    while ($value.StartsWith("./")) {
        $value = $value.Substring(2)
    }
    return $value.TrimStart('/')
}

function Test-InteractiveSession {
    if (-not [Environment]::UserInteractive) {
        return $false
    }
    if ($Host.Name -ne "ConsoleHost") {
        return $false
    }
    try {
        if ([Console]::IsInputRedirected -or [Console]::IsOutputRedirected -or [Console]::IsErrorRedirected) {
            return $false
        }
    } catch {
        # Keep interactive=true when console redirection status cannot be read.
    }
    return $true
}

function Load-CorePathPatterns {
    param([string]$FilePath)

    $patterns = @()
    if (-not (Test-Path $FilePath)) {
        return $patterns
    }

    foreach ($line in Get-Content $FilePath) {
        $trimmed = [string]$line
        $trimmed = $trimmed.Trim()
        if ([string]::IsNullOrWhiteSpace($trimmed)) {
            continue
        }
        if ($trimmed.StartsWith("#")) {
            continue
        }
        $normalized = Normalize-RepoRelPath $trimmed
        if ([string]::IsNullOrWhiteSpace($normalized)) {
            continue
        }
        if ($normalized.EndsWith("/")) {
            $normalized = $normalized + "*"
        }
        $patterns += $normalized
    }
    return $patterns
}

$staged = @(git diff --cached --name-only --diff-filter=ACMR)
$corePathsRel = "$packageName\policy\core_logic_paths.txt"
$corePathsPath = Join-Path $repoRoot $corePathsRel
$corePatterns = Load-CorePathPatterns $corePathsPath
$coreMatched = @()
if ($staged.Count -gt 0 -and $corePatterns.Count -gt 0) {
    $seen = @{}
    foreach ($stagedPath in $staged) {
        $fileNorm = Normalize-RepoRelPath $stagedPath
        if ([string]::IsNullOrWhiteSpace($fileNorm)) {
            continue
        }
        foreach ($patternNorm in $corePatterns) {
            $matcher = [System.Management.Automation.WildcardPattern]::new(
                $patternNorm,
                [System.Management.Automation.WildcardOptions]::IgnoreCase
            )
            if ($matcher.IsMatch($fileNorm)) {
                if (-not $seen.ContainsKey($fileNorm)) {
                    $seen[$fileNorm] = $true
                    $coreMatched += $fileNorm
                }
                break
            }
        }
    }
}

if ($coreMatched.Count -gt 0) {
    Write-Host ""
    Write-Host "[AGD hook][SOFT-GUARD] Core logic files are staged."
    Write-Host "[AGD hook][SOFT-GUARD] Please review before continuing:"
    foreach ($item in $coreMatched) {
        Write-Host ("  - " + $item)
    }
    Write-Host ("[AGD hook][SOFT-GUARD] Policy file: " + $corePathsRel.Replace('\', '/'))

    if (Test-InteractiveSession) {
        $answer = Read-Host "[AGD hook][SOFT-GUARD] Continue commit? (y/N)"
        $normalizedAnswer = [string]$answer
        $normalizedAnswer = $normalizedAnswer.Trim().ToLowerInvariant()
        if ($normalizedAnswer -ne "y" -and $normalizedAnswer -ne "yes") {
            Write-Host "[AGD hook][SOFT-GUARD] Commit canceled by user review."
            exit 1
        }
    } else {
        Write-Host "[AGD hook][SOFT-GUARD] Non-interactive environment detected. Warning only; continuing."
    }
    Write-Host ""
}

$pkgPattern = '^' + [regex]::Escape($packageName) + '/.*\.agd$'
$examplesPattern = '^' + [regex]::Escape($packageName) + '/examples/.*\.agd$'
$agdChanged = @($staged | Where-Object { $_ -match $pkgPattern })
$examplesChanged = @($staged | Where-Object { $_ -match $examplesPattern })

if ($agdChanged.Count -eq 0) {
    exit 0
}

$agdBin = $null
$agdEnPath = Join-Path $repoRoot "$packageName\agd_en.exe"
$agdDefaultPath = Join-Path $repoRoot "$packageName\agd.exe"
if (Test-Path $agdEnPath) {
    $agdBin = $agdEnPath
} elseif (Test-Path $agdDefaultPath) {
    $agdBin = $agdDefaultPath
}

if ($null -eq $agdBin) {
    Write-Host "[AGD hook] AGD executable not found."
    Write-Host "Run setup first: $packageName\\setup.cmd"
    exit 1
}

$agdDocsRel = "$packageName\agd_docs"
$examplesRel = "$packageName\examples"
$agdDocsPath = Join-Path $repoRoot $agdDocsRel
$examplesPath = Join-Path $repoRoot $examplesRel

if (Test-Path $agdDocsPath) {
    Write-Host "[AGD hook] validating $agdDocsRel tree..."
    & $agdBin check-all $agdDocsRel --strict
    if ($LASTEXITCODE -ne 0) {
        exit $LASTEXITCODE
    }
}

if ($examplesChanged.Count -gt 0 -and (Test-Path $examplesPath)) {
    Write-Host "[AGD hook] validating $examplesRel tree..."
    & $agdBin check-all $examplesRel --strict
    if ($LASTEXITCODE -ne 0) {
        exit $LASTEXITCODE
    }
}

foreach ($file in $agdChanged) {
    Write-Host "[AGD hook] checking staged AGD file: $file"
    & $agdBin check $file
    if ($LASTEXITCODE -ne 0) {
        exit $LASTEXITCODE
    }
}

exit 0
