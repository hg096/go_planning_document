@echo off
setlocal EnableExtensions

if /I "%~1"=="-h" goto usage
if /I "%~1"=="--help" goto usage
if "%~1"=="" goto interactive

set "SERVICE_DOC=%~1"
shift
if "%~1"=="" goto usage
set "RUN_CMD=%1"
:collect_non_interactive_cmd
shift
if "%~1"=="" goto non_interactive_cmd_ready
set "RUN_CMD=%RUN_CMD% %1"
goto collect_non_interactive_cmd
:non_interactive_cmd_ready

set "MAX_AGE=%AGD_GATE_MAX_AGE_DAYS%"
if "%MAX_AGE%"=="" set "MAX_AGE=0"
call :validate_non_negative "%MAX_AGE%"
if errorlevel 1 exit /b 2

set "ALLOW_FLAG="
set "ALLOW_RAW=%AGD_GATE_ALLOW_MISSING_IMPACT%"
if /I "%ALLOW_RAW%"=="1" set "ALLOW_FLAG=--allow-missing-impact"
if /I "%ALLOW_RAW%"=="true" set "ALLOW_FLAG=--allow-missing-impact"
if /I "%ALLOW_RAW%"=="y" set "ALLOW_FLAG=--allow-missing-impact"
if /I "%ALLOW_RAW%"=="yes" set "ALLOW_FLAG=--allow-missing-impact"

goto run
 
:interactive
echo [SAFE-RUN] Interactive mode
set "SERVICE_DOC="
set /p SERVICE_DOC=Service doc path or --auto [--auto]: 
if "%SERVICE_DOC%"=="" set "SERVICE_DOC=--auto"

set "MAX_AGE=0"
set /p MAX_AGE=Max age days [0]: 
if "%MAX_AGE%"=="" set "MAX_AGE=0"
call :validate_non_negative "%MAX_AGE%"
if errorlevel 1 exit /b 2

set "ALLOW_ANSWER="
set /p ALLOW_ANSWER=Allow missing impact? [N]: 
set "ALLOW_FLAG="
if /I "%ALLOW_ANSWER%"=="y" set "ALLOW_FLAG=--allow-missing-impact"
if /I "%ALLOW_ANSWER%"=="yes" set "ALLOW_FLAG=--allow-missing-impact"

set "RUN_CMD="
set /p RUN_CMD=Command to execute after gate passes: 
if "%RUN_CMD%"=="" (
  echo [SAFE-RUN] Command is required.
  goto usage
)

:run
if /I "%SERVICE_DOC%"=="--auto" (
  if "%ALLOW_FLAG%"=="" (
    call agd.exe service-gate --max-age-days %MAX_AGE%
  ) else (
    call agd.exe service-gate --max-age-days %MAX_AGE% %ALLOW_FLAG%
  )
) else (
  if "%ALLOW_FLAG%"=="" (
    call agd.exe service-gate "%SERVICE_DOC%" --max-age-days %MAX_AGE%
  ) else (
    call agd.exe service-gate "%SERVICE_DOC%" --max-age-days %MAX_AGE% %ALLOW_FLAG%
  )
)

if errorlevel 1 (
  echo [SAFE-RUN] BLOCKED: service-gate failed.
  exit /b 1
)

echo [SAFE-RUN] Gate passed. Running command:
echo %RUN_CMD%
cmd /c %RUN_CMD%
set "CMD_CODE=%errorlevel%"
if not "%CMD_CODE%"=="0" (
  echo [SAFE-RUN] Command failed with code %CMD_CODE%.
)
exit /b %CMD_CODE%

:validate_non_negative
set "INPUT_NUM=%~1"
echo(%INPUT_NUM%| findstr /r "^[0-9][0-9]*$" >nul
if errorlevel 1 (
  echo [SAFE-RUN] Invalid max age days: %INPUT_NUM%
  exit /b 1
)
exit /b %errorlevel%

:usage
echo Usage:
echo   run-safe.cmd ^<service-doc^|--auto^> ^<command...^>
echo   run-safe.cmd
echo.
echo Examples:
echo   run-safe.cmd 10_source/service/service_logic_checkout_core "go test ./..."
echo   run-safe.cmd --auto "go test ./..."
echo.
echo Optional environment variables:
echo   set AGD_GATE_MAX_AGE_DAYS=0
echo   set AGD_GATE_ALLOW_MISSING_IMPACT=0
exit /b 2
