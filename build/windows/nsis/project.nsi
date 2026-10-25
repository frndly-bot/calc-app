Unicode true

!include "MUI2.nsh"
!include "LogicLib.nsh"
!include "x64.nsh"

!define APP_NAME "Calc"
!define APP_EXE "calc-app.exe"
!define APP_VERSION "0.1.0"
!define APP_PUBLISHER "Calc App Contributors"
!define APP_ID "com.zcode.calc"
!ifndef ARG_WAILS_AMD64_BINARY
	!define ARG_WAILS_AMD64_BINARY "..\..\..\bin\calc-app.exe"
!endif
!ifdef ARG_WAILS_ARM64_BINARY
	!define APP_BINARY "${ARG_WAILS_ARM64_BINARY}"
!else
	!define APP_BINARY "${ARG_WAILS_AMD64_BINARY}"
!endif

Name "${APP_NAME}"
OutFile "..\..\..\bin\calc-app-installer.exe"
InstallDir "$PROGRAMFILES64\${APP_NAME}"
InstallDirRegKey HKLM "Software\${APP_ID}" "InstallDir"
RequestExecutionLevel admin
ManifestDPIAware true

!define MUI_ICON "..\icon.ico"
!define MUI_UNICON "..\icon.ico"
!define MUI_FINISHPAGE_RUN "$INSTDIR\${APP_EXE}"
!define MUI_FINISHPAGE_NOAUTOCLOSE

!insertmacro MUI_PAGE_WELCOME
!insertmacro MUI_PAGE_DIRECTORY
!insertmacro MUI_PAGE_INSTFILES
!insertmacro MUI_PAGE_FINISH
!insertmacro MUI_UNPAGE_INSTFILES
!insertmacro MUI_LANGUAGE "English"

VIProductVersion "${APP_VERSION}.0"
VIFileVersion "${APP_VERSION}.0"
VIAddVersionKey "CompanyName" "${APP_PUBLISHER}"
VIAddVersionKey "FileDescription" "${APP_NAME} Installer"
VIAddVersionKey "FileVersion" "${APP_VERSION}"
VIAddVersionKey "LegalCopyright" "(c) 2026, ${APP_PUBLISHER}"
VIAddVersionKey "ProductName" "${APP_NAME}"
VIAddVersionKey "ProductVersion" "${APP_VERSION}"

Function .onInit
	${IfNot} ${RunningX64}
		MessageBox MB_ICONSTOP "${APP_NAME} requires 64-bit Windows."
		Abort
	${EndIf}
FunctionEnd

Section "Install"
	SetOutPath "$INSTDIR"
	File "${APP_BINARY}"

	WriteRegStr HKLM "Software\${APP_ID}" "InstallDir" "$INSTDIR"
	WriteUninstaller "$INSTDIR\uninstall.exe"

	CreateShortcut "$SMPROGRAMS\${APP_NAME}.lnk" "$INSTDIR\${APP_EXE}"
	CreateShortcut "$DESKTOP\${APP_NAME}.lnk" "$INSTDIR\${APP_EXE}"

	WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${APP_ID}" "DisplayName" "${APP_NAME}"
	WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${APP_ID}" "DisplayVersion" "${APP_VERSION}"
	WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${APP_ID}" "Publisher" "${APP_PUBLISHER}"
	WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${APP_ID}" "DisplayIcon" "$INSTDIR\${APP_EXE}"
	WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${APP_ID}" "UninstallString" "$INSTDIR\uninstall.exe"
	WriteRegDWORD HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${APP_ID}" "NoModify" 1
	WriteRegDWORD HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${APP_ID}" "NoRepair" 1
SectionEnd

Section "Uninstall"
	Delete "$SMPROGRAMS\${APP_NAME}.lnk"
	Delete "$DESKTOP\${APP_NAME}.lnk"
	Delete "$INSTDIR\uninstall.exe"
	Delete "$INSTDIR\${APP_EXE}"
	RMDir "$INSTDIR"

	DeleteRegKey HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${APP_ID}"
	DeleteRegKey HKLM "Software\${APP_ID}"
SectionEnd
