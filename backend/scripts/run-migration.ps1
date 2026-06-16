param(
  [string]$HostName = "127.0.0.1",
  [string]$Port = "3306",
  [string]$User = "root",
  [string]$Password = ""
)

# Requires mysql client in PATH
$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$sqlPath = Join-Path $scriptDir "..\migrations\001_create_transactions.sql"

$mysqlArgs = @("-h", $HostName, "-P", $Port, "-u", $User)
if ($Password -ne "") {
  $mysqlArgs += "-p$Password"
}

Get-Content $sqlPath | & mysql @mysqlArgs
