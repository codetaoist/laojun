@echo off
chcp 65001 >nul
title Laojun System One-Click Deployment

echo.
echo ========================================
echo    Laojun System One-Click Deployment
echo ========================================
echo.

:: Check PowerShell
powershell -Command "Get-Host" >nul 2>&1
if errorlevel 1 (
    echo [ERROR] PowerShell not found, please ensure Windows PowerShell is installed
    pause
    exit /b 1
)

:: Check Docker
docker --version >nul 2>&1
if errorlevel 1 (
    echo [ERROR] Docker not found, please install Docker Desktop first
    echo Download: https://www.docker.com/products/docker-desktop
    pause
    exit /b 1
)

echo [INFO] Starting deployment script...
echo.

:: Run PowerShell deployment script
powershell -ExecutionPolicy Bypass -File "%~dp0one-click-deploy.ps1"

echo.
echo Deployment completed! Press any key to exit...
pause >nul