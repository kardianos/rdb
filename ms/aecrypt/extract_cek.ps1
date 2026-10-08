# extract_cek.ps1 prints the plain form of SQL Server Always Encrypted column
# encryption keys, as "column_key=<fingerprint>.<key>" for an rdb DSN
# (package github.com/kardianos/rdb/ms/aecrypt).
#
# Run it in Windows PowerShell on the machine that holds the column master key
# certificate, as a Windows user that can read the certificate's private key.
# It only decrypts with the private key, so the key need not be exportable.
#
# Read every key of a database (Windows login; add -Credential for a SQL login):
#
#   powershell -ExecutionPolicy Bypass -File extract_cek.ps1 -ServerInstance "(local)\EVOLUTION" -Database MatchX
#
# Or pass one key's stored value, copied from SSMS:
#
#   powershell -ExecutionPolicy Bypass -File extract_cek.ps1 -EncryptedValue 0x01...
#
#   select
#   	EncryptionKey = cek.name,
#   	cekv.encrypted_value
#   from
#   	sys.column_encryption_keys cek
#   	join sys.column_encryption_key_values cekv on cekv.column_encryption_key_id = cek.column_encryption_key_id
#   ;
#
# The output decrypts every column the key protects. Treat it as a password.
param(
	[string] $ServerInstance,
	[string] $Database,
	[PSCredential] $Credential,
	[string] $EncryptedValue
)
$ErrorActionPreference = 'Stop'

function ConvertFrom-Hex([string] $text) {
	$hex = $text.Trim() -replace '^0[xX]', ''
	if ($hex.Length -eq 0 -or $hex.Length % 2 -ne 0) {
		throw 'EncryptedValue is not hex: want 0x01 followed by an even number of hex digits'
	}
	$b = New-Object byte[] ($hex.Length / 2)
	for ($i = 0; $i -lt $b.Length; $i++) {
		$b[$i] = [Convert]::ToByte($hex.Substring($i * 2, 2), 16)
	}
	, $b
}

# Get-StoredKey returns each column encryption key value of the database.
function Get-StoredKey {
	$cs = New-Object System.Data.SqlClient.SqlConnectionStringBuilder
	$cs['Data Source'] = $ServerInstance
	$cs['Initial Catalog'] = $Database
	$cs['Integrated Security'] = ($null -eq $Credential)
	$conn = New-Object System.Data.SqlClient.SqlConnection $cs.ConnectionString
	if ($null -ne $Credential) {
		$pw = $Credential.Password.Copy()
		$pw.MakeReadOnly()
		$conn.Credential = New-Object System.Data.SqlClient.SqlCredential($Credential.UserName, $pw)
	}
	$conn.Open()
	try {
		$cmd = $conn.CreateCommand()
		$cmd.CommandText = @'
select
	EncryptionKey = cek.name,
	cekv.encrypted_value
from
	sys.column_encryption_keys cek
	join sys.column_encryption_key_values cekv on cekv.column_encryption_key_id = cek.column_encryption_key_id
order by
	cek.name
;
'@
		$r = $cmd.ExecuteReader()
		while ($r.Read()) {
			[pscustomobject]@{ Name = $r.GetString(0); Value = [byte[]]$r.GetValue(1) }
		}
		$r.Close()
	} finally {
		$conn.Close()
	}
}

# Get-ColumnKey unwraps one stored key value with its certificate and returns
# the DSN text.
function Get-ColumnKey([byte[]] $blob) {
	# Layout: version 0x01 | key path length (2) | ciphertext length (2) |
	# key path (UTF-16LE) | RSA-OAEP ciphertext of the key | signature.
	if ($blob[0] -ne 1) {
		throw "Stored key value has version $($blob[0]), want 1"
	}
	$pathLen = [BitConverter]::ToUInt16($blob, 1)
	$ctLen = [BitConverter]::ToUInt16($blob, 3)
	$keyPath = [Text.Encoding]::Unicode.GetString($blob, 5, $pathLen)
	$ct = [byte[]]($blob[(5 + $pathLen)..(4 + $pathLen + $ctLen)])

	# The key path names the certificate: <CurrentUser|LocalMachine>/<store>/<thumbprint>.
	$location, $store, $thumb = $keyPath.Split('/')
	$cert = Get-Item "Cert:\$location\$store\$thumb"
	Write-Host "  certificate $($cert.Thumbprint) $($cert.Subject), expires $($cert.NotAfter)"

	$rsa = [Security.Cryptography.X509Certificates.RSACertificateExtensions]::GetRSAPrivateKey($cert)
	if ($null -eq $rsa) {
		throw "No access to the private key of certificate $thumb"
	}
	$cek = $rsa.Decrypt($ct, [Security.Cryptography.RSAEncryptionPadding]::OaepSHA1)
	if ($cek.Length -ne 32) {
		throw "Unwrapped key is $($cek.Length) bytes, want 32"
	}
	# Fingerprint: the first 8 bytes of SHA-256(stored value || key).
	$fp = [Security.Cryptography.SHA256]::Create().ComputeHash([byte[]]($blob + $cek))[0..7]
	'column_key=' + (ConvertTo-Base64Url $fp) + '.' + (ConvertTo-Base64Url $cek)
}

function ConvertTo-Base64Url([byte[]] $b) {
	[Convert]::ToBase64String($b).TrimEnd('=').Replace('+', '-').Replace('/', '_')
}

if ($EncryptedValue) {
	$keys = @([pscustomobject]@{ Name = 'EncryptedValue'; Value = (ConvertFrom-Hex $EncryptedValue) })
} elseif ($ServerInstance -and $Database) {
	$keys = @(Get-StoredKey)
	if ($keys.Count -eq 0) {
		throw "Database $Database has no column encryption keys"
	}
} else {
	throw 'Pass -ServerInstance and -Database to read the keys, or -EncryptedValue 0x01...; see the top of extract_cek.ps1'
}

# A key has one stored value per column master key it is wrapped under; each
# value gives its own column_key, and any of them works.
foreach ($k in $keys) {
	Write-Host "$($k.Name):"
	try {
		Get-ColumnKey $k.Value
	} catch {
		Write-Warning "$($k.Name): $_"
	}
}
