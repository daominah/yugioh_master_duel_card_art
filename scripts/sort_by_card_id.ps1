# sort_by_card_id.ps1
#
# If you hit "running scripts is disabled on this system", the execution policy
# blocks the script before its first line runs, so it cannot bypass itself.
# Launch it one of these ways instead:
#   powershell -ExecutionPolicy Bypass -File scripts\sort_by_card_id.ps1
# or set it once for your user (then plain .\sort_by_card_id.ps1 works):
#   Set-ExecutionPolicy -Scope CurrentUser -ExecutionPolicy RemoteSigned
#
# Rewrites Creation/LastWrite timestamps on card images so that, when sorted by
# date, the images appear in card id order and every image of the same card id
# stays adjacent (full art, cut-in, transparent, profile icon, ...).
#
# Replaces the old sort_censored.ps1, which handled exactly one tcg plus one ocg
# pair per card; here a card id can have any number of images:
#   card index N, image index M within the card -> base + N minutes + 2M seconds
# One whole minute per card id, so even tools that resolve dates only to minutes
# never interleave neighbouring cards.

[CmdletBinding()]
param(
#    [string]$Dir = 'D:\syncthing\Master_Duel_art_full\MD_different_censored',
#    [string]$Dir = 'D:\syncthing\Master_Duel_art_full\upscayled_2048',
    [string]$Dir = 'D:\syncthing\Master_Duel_art_full\MD_art_renamed',
    [datetime]$BaseTime = ([datetime]'2001-01-01T00:00:00'),
    [string[]]$Extensions = @('.png', '.jpg', '.jpeg', '.webp'),
    [int64]$MinCardId = 1000,
    [switch]$DryRun
)

if (-not (Test-Path -LiteralPath $Dir))
{
    throw "Directory not found: $Dir"
}

$files = Get-ChildItem -LiteralPath $Dir -File |
        Where-Object { $Extensions -contains $_.Extension.ToLower() }

# Card id = last token that is all digits, optionally prefixed by "p" as the
# cut-in images are named (accesscode_talker_p15032_cutin2048).
# This skips digits glued to a word (up2048, cutin2048, art2, a01) and digits
# that are part of the card name (number_99__utopia_dragonar_16471 -> 16471).
# An alt art id wins over the base card id when it comes last
# (dogmatika_ecclesia..._15239_alt22186 has no trailing pure-digit token, so it
# keeps 15239, while eldlich..._p3423_cutin2048 sorts as card 3423).
# Numbers below $MinCardId are rejected as too small to be a real card id, which
# drops the last false positives: coin_01, number_39__utopia, number_99__utopia.
function Get-CardId([string]$fileName)
{
    $stem = [System.IO.Path]::GetFileNameWithoutExtension($fileName)
    $id = $null
    foreach ($t in ($stem -split '_'))
    {
        if ($t -match '^p?(?<num>\d+)$')
        {
            $candidate = [int64]$Matches['num']
            if ($candidate -ge $MinCardId)
            {
                $id = $candidate
            }
        }
    }
    return $id
}

# Group by card id so that every image of one card shares a minute. Files with
# no card id keep their own minute each and sort last, by name.
$groups = @{ }   # sort key -> list of FileInfo
$noIdCount = 0
foreach ($f in $files)
{
    $id = Get-CardId $f.Name
    if ($null -eq $id)
    {
        Write-Warning "No card id found in: $( $f.Name ) (sorted last)"
        $key = "2`t$( $f.Name )"
        $noIdCount++
    }
    else
    {
        $key = "1`t$($id.ToString('D12') )"
    }
    if (-not $groups.ContainsKey($key))
    {
        $groups[$key] = New-Object System.Collections.ArrayList
    }
    [void]$groups[$key].Add($f)
}

$n = 0
foreach ($key in ($groups.Keys | Sort-Object))
{
    $group = $groups[$key] | Sort-Object Name
    if ($group.Count -gt 29)
    {
        Write-Warning "Card group $key has $( $group.Count ) images, more than the 29 that fit in one minute"
    }

    $m = 0
    foreach ($file in $group)
    {
        $time = $BaseTime.AddMinutes($n).AddSeconds(2 * $m)
        if ($DryRun)
        {
            Write-Host ("{0:yyyy-MM-dd HH:mm:ss}  {1}" -f $time, $file.Name)
        }
        else
        {
            $file.CreationTime = $time
            $file.LastWriteTime = $time
            $file.LastAccessTime = $time
        }
        $m++
    }

    $n++
}

$mode = if ($DryRun)
{
    'Dry run'
}
else
{
    'Done'
}
Write-Host "$mode on $Dir"
Write-Host "Processed $n card group(s) across $( $files.Count ) file(s), $noIdCount without a card id."
