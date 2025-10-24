@echo off
chcp 65001 >nul
title Laojun System - Local Images Deployment

echo ========================================
echo Laojun System - Local Images Deployment
echo ========================================
echo.

echo [INFO] Checking PowerShell...
where powershell >nul 2>&1
if %errorlevel% neq 0 (
    echo [ERROR] PowerShell not found! Please install PowerShell.
    pause
    exit /b 1
)

echo [INFO] Checking Docker...
where docker >nul 2>&1
if %errorlevel% neq 0 (
    echo [ERROR] Docker not found! Please install Docker Desktop.
    pause
    exit /b 1
)

echo [INFO] Starting deployment with local images priority...
echo.

powershell -ExecutionPolicy Bypass -File "%~dp0one-click-deploy.ps1" -UseLocal

if %errorlevel% equ 0 (
    echo.
    echo ========================================
    echo Deployment completed successfully!
    echo ========================================
    echo.
    echo Access URLs:
    echo - Plugin Marketplace: http://localhost:8080
    echo - Admin Backend: http://localhost:8081
    echo - API Documentation: http://localhost:8082/docs
    echo - Config Center: http://localhost:8083
    echo.
) else (
    echo.
    echo [ERROR] Deployment failed! Please check the logs above.
    echo.
)

pause