$ErrorActionPreference = "Stop"

$repoRoot = (git rev-parse --show-toplevel).Trim()
if ([string]::IsNullOrWhiteSpace($repoRoot)) {
    Write-Error "[AGD hook] failed to resolve repository root."
}
Set-Location $repoRoot

$staged = @(git diff --cached --name-only --diff-filter=ACMR)
$agdChanged = @($staged | Where-Object { $_ -match '^agd/.*\.agd$' })
$examplesChanged = @($staged | Where-Object { $_ -match '^agd/examples/.*\.agd$' })

if ($agdChanged.Count -eq 0) {
    exit 0
}

$agdBin = $null
if (Test-Path ".\agd\agd_en.exe") {
    $agdBin = ".\agd\agd_en.exe"
} elseif (Test-Path ".\agd\agd.exe") {
    $agdBin = ".\agd\agd.exe"
}

if ($null -eq $agdBin) {
    Write-Host "[AGD hook] agd executable not found."
    Write-Host "Run setup first: agd\\setup.cmd"
    exit 1
}

if (Test-Path ".\agd\agd_docs") {
    Write-Host "[AGD hook] validating agd\\agd_docs tree..."
    & $agdBin check-all agd\agd_docs --strict
    if ($LASTEXITCODE -ne 0) {
        exit $LASTEXITCODE
    }
}

if ($examplesChanged.Count -gt 0 -and (Test-Path ".\agd\examples")) {
    Write-Host "[AGD hook] validating agd\\examples tree..."
    & $agdBin check-all agd\examples --strict
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
