@echo off
REM ============================================================
REM  Local EMS — Windows Uninstaller
REM  Right-click → Run as Administrator
REM ============================================================

echo.
echo ╔══════════════════════════════════════════════╗
echo ║       Local EMS — Windows Uninstaller        ║
echo ╚══════════════════════════════════════════════╝
echo.

net session >nul 2>&1
if %errorlevel% neq 0 (
    echo ❌ Please right-click and select "Run as Administrator"
    pause
    exit /b 1
)

set INSTALL_DIR=C:\ems

REM --- Stop and delete service ---
sc query ems >nul 2>&1
if %errorlevel% equ 0 (
    echo Stopping EMS service...
    sc stop ems >nul 2>&1
    timeout /t 3 >nul
    echo Removing EMS service...
    sc delete ems >nul 2>&1
)

REM --- Remove desktop shortcut ---
del "%PUBLIC%\Desktop\EMS Dashboard.url" >nul 2>&1

REM --- Ask about data ---
echo.
set /p CONFIRM="Delete database and historical data? [y/N]: "
if /i "%CONFIRM%"=="y" (
    echo Removing %INSTALL_DIR% including data...
    rmdir /s /q "%INSTALL_DIR%" >nul 2>&1
) else (
    echo Removing binaries and config, keeping data...
    del "%INSTALL_DIR%\ems-core.exe" >nul 2>&1
    del "%INSTALL_DIR%\config.yaml" >nul 2>&1
    del "%INSTALL_DIR%\.env" >nul 2>&1
    del "%INSTALL_DIR%\Start EMS.bat" >nul 2>&1
    echo Data preserved at: %INSTALL_DIR%\data\
)

echo.
echo ✅ EMS uninstalled.
echo.
pause
