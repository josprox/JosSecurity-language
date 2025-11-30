@echo off
echo Building JosSecurity Distribution...
go run cmd/dist/main.go
if %ERRORLEVEL% NEQ 0 (
    echo Build failed!
    exit /b %ERRORLEVEL%
)
echo Build complete.
pause
