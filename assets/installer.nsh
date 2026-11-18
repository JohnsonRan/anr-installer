; Steps of the app's own in the installer (mygo.json, windows.installerInclude).

!define RUN_KEY "Software\Microsoft\Windows\CurrentVersion\Run"
!define APPROVED_KEY "Software\Microsoft\Windows\CurrentVersion\Explorer\StartupApproved\Run"
!define LOGIN_VALUE "io.github.johnsonran.anrinstaller"

; Start with Windows is a Run value MyGo writes under the app's identifier. Uninstalling removes it, else Windows
; would keep starting a program that is gone. The portable version writes the same value: it is removed only when
; it starts this copy, compared without case as MyGo compares it.
!macro customUnInstall
  ReadRegStr $0 HKCU "${RUN_KEY}" "${LOGIN_VALUE}"
  StrCmp $0 '"$INSTDIR\ANR FG Installer.exe" --mygo-opened-at-login' 0 login_kept
    DeleteRegValue HKCU "${RUN_KEY}" "${LOGIN_VALUE}"
    DeleteRegValue HKCU "${APPROVED_KEY}" "${LOGIN_VALUE}"
  login_kept:
!macroend
