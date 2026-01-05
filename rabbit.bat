@echo off
chcp 65001 >nul 2>&1
setlocal enabledelayedexpansion

:: Rabbit Panel Windows Build Script
:: Usage: rabbit.bat {build|help} [target] [options]

:: Get script directory
set "SCRIPT_DIR=%~dp0"
cd /d "%SCRIPT_DIR%"

:: Config
set "FRONTEND_DIR=frontend"
set "BACKEND_DIR=backend"
set "BUILD_LDFLAGS=-s -w"
set "VERSION="
set "SKIP_FRONTEND="

:: Main logic
if "%~1"=="" goto show_help
if "%~1"=="build" goto parse_build_args
if "%~1"=="build-frontend" goto build_frontend_only
if "%~1"=="build-backend" goto parse_build_backend_args
if "%~1"=="help" goto show_help
if "%~1"=="--help" goto show_help
if "%~1"=="-h" goto show_help

echo Unknown command: %~1
goto show_help

:: ============================================
:: Parse build arguments
:: ============================================
:parse_build_args
set "TARGET=%~2"
if "%TARGET%"=="" set "TARGET=auto"

:: Parse remaining arguments
shift
shift
:parse_build_loop
if "%~1"=="" goto do_build
if "%~1"=="--skip-frontend" (
    set "SKIP_FRONTEND=1"
    shift
    goto parse_build_loop
)
if "%~1"=="-v" (
    set "VERSION=%~2"
    shift
    shift
    goto parse_build_loop
)
if "%~1"=="--version" (
    set "VERSION=%~2"
    shift
    shift
    goto parse_build_loop
)
shift
goto parse_build_loop

:do_build
if not "%SKIP_FRONTEND%"=="1" (
    call :build_frontend
    if errorlevel 1 exit /b 1
)
call :do_build_backend
goto end

:: ============================================
:: Parse build-backend arguments
:: ============================================
:parse_build_backend_args
set "TARGET=%~2"
if "%TARGET%"=="" set "TARGET=auto"

:: Parse remaining arguments
shift
shift
:parse_backend_loop
if "%~1"=="" goto do_build_backend
if "%~1"=="-v" (
    set "VERSION=%~2"
    shift
    shift
    goto parse_backend_loop
)
if "%~1"=="--version" (
    set "VERSION=%~2"
    shift
    shift
    goto parse_backend_loop
)
shift
goto parse_backend_loop

:: ============================================
:: Build frontend only
:: ============================================
:build_frontend_only
call :build_frontend
goto end

:: ============================================
:: Build frontend
:: ============================================
:build_frontend
echo [INFO] Building frontend...

where node >nul 2>&1
if errorlevel 1 (
    echo [ERROR] Node.js not found, please install Node.js 18+
    exit /b 1
)

pushd "%FRONTEND_DIR%"

if not exist "node_modules" (
    echo [INFO] Installing frontend dependencies...
    call npm install
    if errorlevel 1 (
        echo [ERROR] Failed to install frontend dependencies
        popd
        exit /b 1
    )
)

call npm run build
if errorlevel 1 (
    echo [ERROR] Frontend build failed
    popd
    exit /b 1
)

popd
echo [OK] Frontend build completed
goto :eof

:: ============================================
:: Build backend
:: ============================================
:do_build_backend
where go >nul 2>&1
if errorlevel 1 (
    echo [ERROR] Go not found, please install Go 1.22+
    exit /b 1
)

set "GOPROXY=https://goproxy.cn,direct"

if "%TARGET%"=="auto" (
    call :build_single_target windows amd64 "" windows-amd64 .exe
    goto :eof
)

if "%TARGET%"=="all" (
    call :build_single_target linux amd64 "" linux-amd64 ""
    call :build_single_target linux arm64 "" linux-arm64 ""
    call :build_single_target linux arm 7 linux-armv7 ""
    call :build_single_target windows amd64 "" windows-amd64 .exe
    call :build_single_target windows arm64 "" windows-arm64 .exe
    call :build_single_target darwin amd64 "" darwin-amd64 ""
    call :build_single_target darwin arm64 "" darwin-arm64 ""
    goto :eof
)

if "%TARGET%"=="linux-amd64" (
    call :build_single_target linux amd64 "" linux-amd64 ""
    goto :eof
)

if "%TARGET%"=="linux-arm64" (
    call :build_single_target linux arm64 "" linux-arm64 ""
    goto :eof
)

if "%TARGET%"=="linux-armv7" (
    call :build_single_target linux arm 7 linux-armv7 ""
    goto :eof
)

if "%TARGET%"=="windows-amd64" (
    call :build_single_target windows amd64 "" windows-amd64 .exe
    goto :eof
)

if "%TARGET%"=="windows-arm64" (
    call :build_single_target windows arm64 "" windows-arm64 .exe
    goto :eof
)

if "%TARGET%"=="darwin-amd64" (
    call :build_single_target darwin amd64 "" darwin-amd64 ""
    goto :eof
)

if "%TARGET%"=="darwin-arm64" (
    call :build_single_target darwin arm64 "" darwin-arm64 ""
    goto :eof
)

echo [ERROR] Unknown target: %TARGET%
echo Available: auto, all, linux-amd64, linux-arm64, linux-armv7, windows-amd64, windows-arm64, darwin-amd64, darwin-arm64
exit /b 1

:: ============================================
:: Build single target
:: Args: %1=GOOS %2=GOARCH %3=GOARM %4=suffix %5=ext
:: ============================================
:build_single_target
set "B_GOOS=%~1"
set "B_GOARCH=%~2"
set "B_GOARM=%~3"
set "B_SUFFIX=%~4"
set "B_EXT=%~5"

:: Build output filename with optional version
if "%VERSION%"=="" (
    set "OUTPUT_NAME=rabbit-panel-%B_SUFFIX%"
) else (
    set "OUTPUT_NAME=rabbit-panel-%B_SUFFIX%-%VERSION%"
)

set "OUTPUT_DIR=dist\%OUTPUT_NAME%-release"
set "OUTPUT_FILE=%OUTPUT_DIR%\%OUTPUT_NAME%%B_EXT%"

if not exist "%OUTPUT_DIR%" mkdir "%OUTPUT_DIR%"

echo [INFO] Building %OUTPUT_NAME% ...

pushd "%BACKEND_DIR%"

set "GOOS=%B_GOOS%"
set "GOARCH=%B_GOARCH%"
set "CGO_ENABLED=0"
if not "%B_GOARM%"=="" set "GOARM=%B_GOARM%"

go build -trimpath -ldflags="%BUILD_LDFLAGS%" -o "..\%OUTPUT_FILE%" .

if errorlevel 1 (
    echo [ERROR] Build failed: %B_SUFFIX%
    popd
    exit /b 1
)

popd

echo [OK] Output: %OUTPUT_FILE%
goto :eof

:: ============================================
:: Show help
:: ============================================
:show_help
echo.
echo Rabbit Panel Windows Build Script
echo.
echo Usage: %~nx0 ^<command^> [options]
echo.
echo Commands:
echo   build [target]          Build the project
echo   build-frontend          Build frontend only
echo   build-backend [target]  Build backend only
echo   help                    Show this help
echo.
echo Build Targets:
echo   auto            Auto detect (default: windows-amd64)
echo   all             Build all platforms
echo   linux-amd64     Linux AMD64
echo   linux-arm64     Linux ARM64
echo   linux-armv7     Linux ARMv7
echo   windows-amd64   Windows AMD64
echo   windows-arm64   Windows ARM64
echo   darwin-amd64    macOS AMD64 (Intel)
echo   darwin-arm64    macOS ARM64 (Apple Silicon)
echo.
echo Options:
echo   --skip-frontend         Skip frontend build
echo   -v, --version ^<ver^>     Set version (e.g., v1.3.2)
echo.
echo Examples:
echo   %~nx0 build                                   Build for current platform
echo   %~nx0 build all                               Build for all platforms
echo   %~nx0 build linux-amd64                       Build Linux AMD64 only
echo   %~nx0 build linux-amd64 -v v1.3.2             Build with version
echo   %~nx0 build all --version v1.3.2              Build all with version
echo   %~nx0 build auto --skip-frontend              Skip frontend build
echo   %~nx0 build linux-arm64 --skip-frontend -v v1.3.2
echo.
echo Output naming:
echo   Without version: rabbit-panel-linux-amd64.exe
echo   With version:    rabbit-panel-linux-amd64-v1.3.2.exe
echo.
goto end

:end
endlocal
