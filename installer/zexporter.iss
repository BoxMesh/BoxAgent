; Inno Setup script: installs ZExporter as a Windows service and asks for the BoxServer gRPC address.
; Run the installer again to repair or to change the address: the box is pre-filled with the saved one.
; Silent: zexporter-windows.exe /VERYSILENT /SERVER=host:port
; Build: ISCC.exe installer\zexporter.iss  (expects dist\zexporter-windows-amd64.exe; VERSION env var sets the version)

#define Version GetEnv("VERSION")
#if Version == ""
  #define Version "0.0.0"
#endif

[Setup]
AppId={{B4D2C1E6-7A3F-4C58-9E21-5F0A8D6C3B17}
AppName=ZExporter
AppVersion={#Version}
AppPublisher=Zin
DefaultDirName={autopf}\ZExporter
DisableDirPage=yes
DisableProgramGroupPage=yes
PrivilegesRequired=admin
ArchitecturesAllowed=x64
ArchitecturesInstallIn64BitMode=x64
OutputDir=..\dist
OutputBaseFilename=zexporter-windows
Compression=lzma2
SolidCompression=yes
UninstallDisplayName=ZExporter

[Files]
Source: "..\dist\zexporter-windows-amd64.exe"; DestDir: "{app}"; DestName: "zexporter.exe"; Flags: ignoreversion

[Registry]
Root: HKLM; Subkey: "Software\ZExporter"; ValueType: string; ValueName: "Server"; ValueData: "{code:GetServer}"; Flags: uninsdeletekey

[Code]
var
  ServerPage: TInputQueryWizardPage;

function GetServer(Param: String): String;
begin
  Result := Trim(ServerPage.Values[0]);
end;

// host:port made only of characters that are safe to put on a command line.
function ValidServer(S: String): Boolean;
var
  I: Integer;
begin
  Result := (Pos(':', S) > 1) and (Pos(':', S) < Length(S));
  for I := 1 to Length(S) do
    if Pos(S[I], 'abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789.-_:') = 0 then
      Result := False;
end;

function Sc(Args: String): Integer;
var
  Code: Integer;
begin
  if not Exec(ExpandConstant('{sys}\sc.exe'), Args, '', SW_HIDE, ewWaitUntilTerminated, Code) then
    Code := -1;
  Result := Code;
end;

procedure InitializeWizard;
var
  Saved: String;
begin
  Saved := ExpandConstant('{param:SERVER|}');
  if Saved = '' then
    RegQueryStringValue(HKLM, 'Software\ZExporter', 'Server', Saved);
  ServerPage := CreateInputQueryPage(wpWelcome, 'BoxServer address',
    'Where should the agent report to?',
    'Enter the BoxServer gRPC address as host:port, for example 192.168.1.10:8081. Run this installer again to change it.');
  ServerPage.Add('Server address:', False);
  ServerPage.Values[0] := Saved;
end;

function NextButtonClick(CurPageID: Integer): Boolean;
begin
  Result := True;
  if CurPageID = ServerPage.ID then
    if not ValidServer(GetServer('')) then
    begin
      MsgBox('Enter the address as host:port (letters, digits, dots and dashes only).', mbError, MB_OK);
      Result := False;
    end;
end;

// Also covers silent installs, which skip the page above.
function PrepareToInstall(var NeedsRestart: Boolean): String;
begin
  Result := '';
  if not ValidServer(GetServer('')) then
    Result := 'The server address must be host:port, for example 192.168.1.10:8081.'
  else if Sc('stop ZExporter') = 0 then
    Sleep(3000); // let the running service release zexporter.exe before it is replaced
end;

procedure CurStepChanged(CurStep: TSetupStep);
var
  Bin, Action: String;
begin
  if CurStep <> ssPostInstall then Exit;
  Bin := '"\"' + ExpandConstant('{app}\zexporter.exe') + '\" --server ' + GetServer('') + '"';
  if Sc('query ZExporter') = 0 then Action := 'config' else Action := 'create';
  Sc(Action + ' ZExporter binPath= ' + Bin + ' start= auto DisplayName= "ZExporter"');
  Sc('failure ZExporter reset= 86400 actions= restart/5000/restart/5000/restart/5000');
  Sc('failureflag ZExporter 1');
  Sc('start ZExporter');
end;

procedure CurUninstallStepChanged(CurUninstallStep: TUninstallStep);
begin
  if CurUninstallStep = usUninstall then
  begin
    if Sc('stop ZExporter') = 0 then Sleep(3000);
    Sc('delete ZExporter');
  end;
end;
