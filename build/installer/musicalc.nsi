Unicode true
!include "MUI2.nsh"
!include "x64.nsh"
!include "WinVer.nsh"

!ifndef AppVersion
  !error "AppVersion must be supplied by build-nsis-installer.ps1"
!endif
!ifndef NumericVersion
  !error "NumericVersion must be supplied by build-nsis-installer.ps1"
!endif
!ifndef SourceRoot
  !error "SourceRoot must be supplied by build-nsis-installer.ps1"
!endif
!ifndef AppExe
  !error "AppExe must be supplied by build-nsis-installer.ps1"
!endif
!ifndef InstallerOutput
  !error "InstallerOutput must be supplied by build-nsis-installer.ps1"
!endif

!define AppName "MusiCalc"
!define AppKey "Software\MusiCalc\NSIS"
!define UninstallKey "Software\Microsoft\Windows\CurrentVersion\Uninstall\MusiCalc.NSIS"
!define InnoKey "Software\Microsoft\Windows\CurrentVersion\Uninstall\{8F9A3D2E-5B6C-4A7E-9D8F-1C2E3A4B5C6D}_is1"

Name "${AppName} ${AppVersion}"
OutFile "${InstallerOutput}"
InstallDir "$PROGRAMFILES64\${AppName}"
RequestExecutionLevel admin
SetCompressor /SOLID lzma
SetOverwrite on
ShowInstDetails show
ShowUninstDetails show

VIProductVersion "${NumericVersion}"
VIAddVersionKey /LANG=1033 "ProductName" "${AppName}"
VIAddVersionKey /LANG=1033 "ProductVersion" "${AppVersion}"
VIAddVersionKey /LANG=1033 "FileVersion" "${AppVersion}"
VIAddVersionKey /LANG=1033 "FileDescription" "${AppName} Windows x64 Setup"
VIAddVersionKey /LANG=1033 "CompanyName" "B. Zeiss"
VIAddVersionKey /LANG=1033 "LegalCopyright" "Copyright (c) 2026 B. Zeiss"

!define MUI_ICON "${SourceRoot}\icons\appicon.ico"
!define MUI_UNICON "${SourceRoot}\icons\appicon.ico"
!define MUI_ABORTWARNING
!define MUI_LANGDLL_REGISTRY_ROOT "HKLM"
!define MUI_LANGDLL_REGISTRY_KEY "${AppKey}"
!define MUI_LANGDLL_REGISTRY_VALUENAME "InstallerLanguage"
!insertmacro MUI_PAGE_WELCOME
!insertmacro MUI_PAGE_LICENSE "${SourceRoot}\LICENSE"
!insertmacro MUI_PAGE_COMPONENTS
!insertmacro MUI_PAGE_DIRECTORY
!insertmacro MUI_PAGE_INSTFILES
!insertmacro MUI_PAGE_FINISH
!insertmacro MUI_UNPAGE_CONFIRM
!insertmacro MUI_UNPAGE_INSTFILES
!insertmacro MUI_LANGUAGE "English"
!insertmacro MUI_LANGUAGE "German"

LangString AppSection ${LANG_ENGLISH} "MusiCalc (required)"
LangString AppSection ${LANG_GERMAN} "MusiCalc (erforderlich)"
LangString DesktopShortcut ${LANG_ENGLISH} "Desktop shortcut"
LangString DesktopShortcut ${LANG_GERMAN} "Desktop-Verknüpfung"
LangString Unsupported ${LANG_ENGLISH} "This package requires Windows 10 or later on an x64 processor."
LangString Unsupported ${LANG_GERMAN} "Dieses Paket benötigt Windows 10 oder neuer auf einem x64-Prozessor."
LangString InnoConflict ${LANG_ENGLISH} "Uninstall the existing Inno Setup installation of MusiCalc before using this NSIS package."
LangString InnoConflict ${LANG_GERMAN} "Deinstallieren Sie die vorhandene Inno-Setup-Installation von MusiCalc, bevor Sie dieses NSIS-Paket verwenden."
LangString CloseApp ${LANG_ENGLISH} "Close MusiCalc before uninstalling, then try again."
LangString CloseApp ${LANG_GERMAN} "Schließen Sie MusiCalc vor der Deinstallation und versuchen Sie es erneut."

Function .onInit
  SetRegView 64
  SetShellVarContext all
  !insertmacro MUI_LANGDLL_DISPLAY
  ${IfNot} ${AtLeastWin10}
  ${OrIfNot} ${IsNativeAMD64}
    MessageBox MB_OK|MB_ICONSTOP "$(Unsupported)" /SD IDOK
    SetErrorLevel 1633
    Abort
  ${EndIf}
  ReadRegStr $0 HKLM "${InnoKey}" "UninstallString"
  ${If} $0 != ""
    MessageBox MB_OK|MB_ICONSTOP "$(InnoConflict)" /SD IDOK
    SetErrorLevel 1638
    Abort
  ${EndIf}
  ReadRegStr $0 HKLM "${AppKey}" "InstallDir"
  ${If} $0 != ""
    StrCpy $INSTDIR $0
  ${EndIf}
FunctionEnd

Function un.onInit
  SetRegView 64
  SetShellVarContext all
  !insertmacro MUI_UNGETLANGUAGE
FunctionEnd

Section "$(AppSection)"
  SectionIn RO
  SetOutPath "$INSTDIR"
  File /oname=musicalc.exe "${AppExe}"
  File /oname=LICENSE.txt "${SourceRoot}\LICENSE"
  File "${SourceRoot}\icons\appicon.ico"
  SetOutPath "$INSTDIR\icons"
  File "${SourceRoot}\icons\appicon.png"
  WriteUninstaller "$INSTDIR\Uninstall.exe"
  SetOutPath "$INSTDIR"
  CreateShortcut "$SMPROGRAMS\MusiCalc.lnk" "$INSTDIR\musicalc.exe" "" "$INSTDIR\appicon.ico"
  WriteRegStr HKLM "${AppKey}" "InstallDir" "$INSTDIR"
  WriteRegStr HKLM "${UninstallKey}" "DisplayName" "${AppName}"
  WriteRegStr HKLM "${UninstallKey}" "DisplayVersion" "${AppVersion}"
  WriteRegStr HKLM "${UninstallKey}" "Publisher" "B. Zeiss"
  WriteRegStr HKLM "${UninstallKey}" "URLInfoAbout" "https://github.com/tphzz/musicalc"
  WriteRegStr HKLM "${UninstallKey}" "DisplayIcon" "$INSTDIR\appicon.ico"
  WriteRegStr HKLM "${UninstallKey}" "InstallLocation" "$INSTDIR"
  WriteRegStr HKLM "${UninstallKey}" "UninstallString" '$\"$INSTDIR\Uninstall.exe$\"'
  WriteRegStr HKLM "${UninstallKey}" "QuietUninstallString" '$\"$INSTDIR\Uninstall.exe$\" /S'
  WriteRegDWORD HKLM "${UninstallKey}" "NoModify" 1
  WriteRegDWORD HKLM "${UninstallKey}" "NoRepair" 1
SectionEnd

Section /o "$(DesktopShortcut)"
  CreateShortcut "$DESKTOP\MusiCalc.lnk" "$INSTDIR\musicalc.exe" "" "$INSTDIR\appicon.ico"
SectionEnd

Section "Uninstall"
  ClearErrors
  Delete "$INSTDIR\musicalc.exe"
  ${If} ${Errors}
    MessageBox MB_OK|MB_ICONSTOP "$(CloseApp)" /SD IDOK
    SetErrorLevel 1
    Abort
  ${EndIf}
  Delete "$INSTDIR\LICENSE.txt"
  Delete "$INSTDIR\appicon.ico"
  Delete "$INSTDIR\icons\appicon.png"
  Delete "$INSTDIR\Uninstall.exe"
  RMDir "$INSTDIR\icons"
  RMDir "$INSTDIR"
  Delete "$SMPROGRAMS\MusiCalc.lnk"
  Delete "$DESKTOP\MusiCalc.lnk"
  DeleteRegKey HKLM "${UninstallKey}"
  DeleteRegKey HKLM "${AppKey}"
SectionEnd
