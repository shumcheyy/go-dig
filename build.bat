@echo off
REM Build script for go-dig Windows executable

echo Building go-dig.exe for Windows...

REM Create build directory if it doesn't exist
if not exist "build" mkdir build

REM Build the executable
set GOOS=windows
set GOARCH=amd64
go build -ldflags "-s -w" -o build\go-dig.exe .

if %ERRORLEVEL% EQU 0 (
    echo Build successful: build\go-dig.exe
    echo.
    echo To test the executable, run:
    echo   build\go-dig.exe google.com
) else (
    echo Build failed!
    exit /b 1
)