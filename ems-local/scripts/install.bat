@echo off
REM ============================================================
REM  Local EMS — Windows Installer
REM  Right-click → Run as Administrator
REM ============================================================

echo.
echo ╔══════════════════════════════════════════════╗
echo ║       Local EMS — Windows Installer          ║
echo ╚══════════════════════════════════════════════╝
echo.

REM --- Check admin ---
net session >nul 2>&1
if %errorlevel% neq 0 (
    echo ❌ Please right-click and select "Run as Administrator"
    pause
    exit /b 1
)

REM --- Set paths ---
set INSTALL_DIR=C:\ems
set SCRIPT_DIR=%~dp0

REM --- Check files ---
if not exist "%SCRIPT_DIR%ems-core.exe" (
    echo ❌ ems-core.exe not found in %SCRIPT_DIR%
    pause
    exit /b 1
)
if not exist "%SCRIPT_DIR%config.yaml" (
    echo ❌ config.yaml not found in %SCRIPT_DIR%
    pause
    exit /b 1
)
if not exist "%SCRIPT_DIR%.env" (
    echo ❌ .env not found in %SCRIPT_DIR%
    pause
    exit /b 1
)

REM --- Stop existing service if running ---
sc query ems >nul 2>&1
if %errorlevel% equ 0 (
    echo Stopping existing EMS service...
    sc stop ems >nul 2>&1
    timeout /t 3 >nul
    sc delete ems >nul 2>&1
    timeout /t 2 >nul
)

REM --- Install files ---
echo Installing to %INSTALL_DIR%...
if not exist "%INSTALL_DIR%" mkdir "%INSTALL_DIR%"
if not exist "%INSTALL_DIR%\data" mkdir "%INSTALL_DIR%\data"

copy /Y "%SCRIPT_DIR%ems-core.exe" "%INSTALL_DIR%\ems-core.exe" >nul
copy /Y "%SCRIPT_DIR%config.yaml"  "%INSTALL_DIR%\config.yaml" >nul
copy /Y "%SCRIPT_DIR%.env"         "%INSTALL_DIR%\.env" >nul

REM --- Create start script ---
(
echo @echo off
echo cd /d "%INSTALL_DIR%"
echo echo Starting Local EMS...
echo echo Dashboard: http://localhost:8080
echo echo Press Ctrl+C to stop
echo echo.
echo ems-core.exe --config config.yaml
echo pause
) > "%INSTALL_DIR%\Start EMS.bat"

REM --- Create Windows Service ---
echo Creating Windows service...
sc create ems binPath= "\"%INSTALL_DIR%\ems-core.exe\" --config \"%INSTALL_DIR%\config.yaml\"" start= auto DisplayName= "Local EMS" >nul 2>&1
sc description ems "Local Energy Management System" >nul 2>&1
sc start ems >nul 2>&1

REM --- Create desktop shortcut for dashboard ---
echo Creating desktop shortcut...
powershell -Command "$ws = New-Object -ComObject WScript.Shell; $s = $ws.CreateShortcut([Environment]::GetFolderPath('CommonDesktopDirectory') + '\EMS Dashboard.url'); $s.TargetPath = 'http://localhost:8080'; $s.Save()" >nul 2>&1

REM --- Done ---
echo.
echo ════════════════════════════════════════════════
echo   ✅ EMS installed successfully!
echo.
echo   Install dir:  %INSTALL_DIR%
echo   Dashboard:    http://localhost:8080
echo   Config:       %INSTALL_DIR%\.env
echo.
echo   The EMS service starts automatically with Windows.
echo   A dashboard shortcut has been added to your Desktop.
echo.
echo   To manage the service:
echo     sc stop ems       — stop
echo     sc start ems      — start
echo     sc query ems      — check status
echo ════════════════════════════════════════════════
echo.
pause
