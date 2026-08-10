# sort_censored.ps1
#
# If you hit "running scripts is disabled on this system", the execution policy
# blocks the script before its first line runs, so it cannot bypass itself.
# Launch it one of these ways instead:
#   powershell -ExecutionPolicy Bypass -File D:\syncthing\Master_Duel_art_full\sort_censored.ps1
# or set it once for your user (then plain .\sort_censored.ps1 works):
#   Set-ExecutionPolicy -Scope CurrentUser -ExecutionPolicy RemoteSigned
#
# Rewrites Creation/LastWrite timestamps on the censored card images so that,
# when sorted by date, each card's _tcg and _ocg images stay adjacent and the
# cards appear in card-number order.
#
# For card index N (cards sorted by numeric card id, then base name):
#   <name>_tcg.png -> base + N minutes
#   <name>_ocg.png -> base + N minutes + 30 seconds
# One whole minute per card, so even tools that resolve dates only to minutes
# keep each pair together and never interleave neighbouring cards. The 30s gives
# tcg a stable lead over ocg at full (sub-minute) precision.

[CmdletBinding()]
param(
    [string]$Dir = 'D:\syncthing\Master_Duel_art_full\MD_different_censored',
    [datetime]$BaseTime = ([datetime]'2001-01-01T00:00:00')
)

if (-not (Test-Path -LiteralPath $Dir)) {
    throw "Directory not found: $Dir"
}

# Collect *.png and group by base name (everything before the _tcg/_ocg suffix).
$files = Get-ChildItem -LiteralPath $Dir -Filter *.png -File

$groups = @{}   # baseName -> @{ tcg = FileInfo; ocg = FileInfo }
foreach ($f in $files) {
    if ($f.Name -match '^(?<base>.+)_(?<kind>tcg|ocg)\.png$') {
        $base = $Matches['base']
        $kind = $Matches['kind']
        if (-not $groups.ContainsKey($base)) { $groups[$base] = @{} }
        $groups[$base][$kind] = $f
    }
    else {
        Write-Warning "Skipping (no tcg/ocg suffix): $($f.Name)"
    }
}

# Card number = last pure-digit token in the base name. This skips digits that
# are part of the name (7_completed_4806 -> 4806, allure_queen_lv7_6872 -> 6872)
# and ignores alt suffixes (eldlich_..._15123_alt3423 -> 15123).
function Get-CardNumber([string]$base) {
    $num = $null
    foreach ($t in ($base -split '_')) {
        if ($t -match '^\d+$') { $num = [int64]$t }
    }
    if ($null -eq $num) {
        Write-Warning "No card number found in: $base (sorted last)"
        return [int64]::MaxValue
    }
    return $num
}

# Order cards by numeric card id, then base name as a stable tiebreaker
# (keeps alt variants that share a number adjacent and deterministic).
$baseNames = $groups.Keys |
    Sort-Object @{ Expression = { Get-CardNumber $_ } }, @{ Expression = { $_ } }

$n = 0
foreach ($base in $baseNames) {
    $g = $groups[$base]

    foreach ($kind in 'tcg', 'ocg') {
        $file = $g[$kind]
        if ($null -eq $file) {
            Write-Warning "Missing $kind for: $base"
            continue
        }
        $time = $BaseTime.AddMinutes($n)
        if ($kind -eq 'ocg') { $time = $time.AddSeconds(30) }

        $file.CreationTime   = $time
        $file.LastWriteTime  = $time
        $file.LastAccessTime = $time
    }

    $n++
}

Write-Host "Done. Processed $n card(s) across $($files.Count) file(s)."
