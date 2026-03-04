$ErrorActionPreference = "Stop"

$scriptRoot = Split-Path -Parent $MyInvocation.MyCommand.Path
$packageRoot = Split-Path -Parent (Split-Path -Parent $scriptRoot)
$packageName = Split-Path -Leaf $packageRoot

$repoRoot = (git rev-parse --show-toplevel).Trim()
if ([string]::IsNullOrWhiteSpace($repoRoot)) {
    Write-Error "[AGD hook] failed to resolve repository root."
}
Set-Location $repoRoot

$staged = @(git diff --cached --name-only --diff-filter=ACMR)
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
