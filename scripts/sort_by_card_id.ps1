# sort_by_card_id.ps1
#
# If you hit "running scripts is disabled on this system", the execution policy
# blocks the script before its first line runs, so it cannot bypass itself.
# Launch it one of these ways instead:
#   powershell -ExecutionPolicy Bypass -File sort_by_card_id.ps1
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
    [string]$Dir = 'D:\syncthing\Master_Duel_art_full\upscayled_2048',
#    [string]$Dir = 'D:\syncthing\Master_Duel_art_full\yugioh_card_editor\card_result',
#    [string]$Dir = 'D:\syncthing\Master_Duel_art_full\MD_art_renamed',
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

# File names end with the ids, in one of these shapes (suffix words are skipped):
#   {name}_{originalCardID}_alt{n}_{suffix}     alt art of a card
#   {name}_{originalCardID}_token{n}_{suffix}   token of a card
#   {name}_{originalCardID}_p{n}_cutin2048      alt art cut-in
#   {name}_p{n}_cutin2048                       cut-in, p{n} is the card id itself
#   {name}_{originalCardID}_{suffix}            plain card
# Scan the tokens from the end and look at up to 2 id patterns:
# alt{n}, token{n}, p{n}, {n}.
# The primary id is the originalCardID, the bare {n}, so an alt art or token
# stays next to its original card; without a bare {n} the p{n} is the primary id.
# When the name has both, the number of the alt{n}, token{n} or p{n} is the
# secondary id: the sort key inside the card, after the images without one.
# Digits glued to a word (up2048, cutin2048, art2) are not id patterns,
# and scanning stops at the bare {n}, so digits inside the card name
# (number_99__utopia_dragonar_16471) are never reached.
# Bare {n} and p{n} below $MinCardId are rejected as too small to be a real
# card id (coin_01, number_39__utopia); alt{n} and token{n} have no such limit (alt3, alt4).
function Get-CardId([string]$fileName)
{
    $stem = [System.IO.Path]::GetFileNameWithoutExtension($fileName)
    $tokens = $stem -split '_'
    $patternCount = 0
    $secondary = $null
    $printedId = $null
    for ($i = $tokens.Count - 1; $i -ge 0 -and $patternCount -lt 2; $i--)
    {
        $t = $tokens[$i]
        if ($t -match '^(alt|token)(?<num>\d+)$')
        {
            $patternCount++
            if ($null -eq $secondary)
            {
                $secondary = [int64]$Matches['num']
            }
        }
        elseif ($t -match '^p(?<num>\d+)$' -and [int64]$Matches['num'] -ge $MinCardId)
        {
            $patternCount++
            if ($null -eq $secondary)
            {
                $secondary = [int64]$Matches['num']
            }
            if ($null -eq $printedId)
            {
                $printedId = [int64]$Matches['num']
            }
        }
        elseif ($t -match '^\d+$' -and [int64]$t -ge $MinCardId)
        {
            return [pscustomobject]@{ Id = [int64]$t; Secondary = $secondary }
        }
    }
    if ($null -eq $printedId)
    {
        return $null
    }
    return [pscustomobject]@{ Id = $printedId; Secondary = $null }
}

# Group by card id so that every image of one card shares a minute. Files with
# no card id are skipped, so their timestamps stay untouched.
$groups = @{ }   # sort key -> list of @{ File; Secondary }
$noIdCount = 0
foreach ($f in $files)
{
    $ids = Get-CardId $f.Name
    if ($null -eq $ids)
    {
        Write-Warning "No card id found in: $( $f.Name ) (skipped)"
        $noIdCount++
        continue
    }
    $key = $ids.Id.ToString('D12')
    if (-not $groups.ContainsKey($key))
    {
        $groups[$key] = New-Object System.Collections.ArrayList
    }
    [void]$groups[$key].Add([pscustomobject]@{ File = $f; Secondary = $ids.Secondary })
}

$n = 0
foreach ($key in ($groups.Keys | Sort-Object))
{
    # Images without a secondary id first, then by secondary id, then by name.
    $group = $groups[$key] |
            Sort-Object @{ Expression = { if ($null -eq $_.Secondary)
            {
                -1
            }
            else
            {
                $_.Secondary
            } } },
            @{ Expression = { $_.File.Name } } |
            ForEach-Object { $_.File }
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
Write-Host "Processed $n card group(s) across $( $files.Count - $noIdCount ) file(s), skipped $noIdCount without a card id."
