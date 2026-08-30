# extract_md_assets_chunks.ps1
#
# Runs AssetStudioModCLI once per hex bucket of the Master Duel LocalData store.
# Each invocation loads only ~150 bundles, exports,
# then exits and frees all memory before the next bucket.
# This keeps peak RAM small and avoids the out-of-memory crash
# that happens when AssetStudio loads all ~38k bundles (13 GB) at once.
#
# The game keeps assets in three stores, and this script covers all three:
#   1. LocalData\<account>\0000: the downloaded store (~13 GB),
#      processed one hex bucket per run.
#   2. masterduel_Data\StreamingAssets\AssetBundle: the bootstrap set (~100 MB).
#   3. masterduel_Data\data.unity3d: the store built into the Unity player (~122 MB).
#
# Launch (execution policy blocks scripts before line 1, so bypass at launch):
#   powershell -ExecutionPolicy Bypass -File .\scripts\extract_md_assets_chunks.ps1
#
# Memory levers used:
#   -t tex2d,sprite,..    Everything except MonoBehaviour and Animator.
#                         - MonoBehaviour: excluded because it is a huge pile
#                           of human-unreadable data assets.
#                         - Animator: excluded because its FBX export path is error-prone.
#                         Skipping MonoBehaviour .
#                         - AnimationClip has no CLI -t token: it only exports
#                           via -m animator, the very path we are avoiding.
#   --max-export-tasks 2  Cap concurrent texture decodes (each decodes to RGBA).
#   per-bucket loop       Only one bucket's assets are resident at a time.
#
# We do not pass --decompress-to-disk: per-bucket loading already bounds memory
# (about 50 MB per bucket), so keeping decompression in RAM is faster.
#
# -g container (default) preserves the ocg/tcg container path,
# so the output keeps the card/images/illust/{ocg,tcg}/.. structure
# even though we chunk.
#
# The CLI skips files that already exist, so this script is resumable:
# rerun it and it continues where it stopped.
# Pointed at an existing extraction,
# it only writes the assets a previous out-of-memory crash missed.
#
# -KeepDifferentVariants changes how an already existing output file is handled.
# Instead of the CLI skipping by file name, each input is exported to a staging
# folder first, then every file is reconciled into the output tree by content:
# - output missing: moved in as is.
# - identical content: dropped (skip).
# - same path, new content: kept as "<name>_dYYYYMMDD<ext>",
#   so a second account's different art for the same
#   card id is preserved next to the first.
# This mode re-extracts every file each run (it gives up the fast file-name
# skip), so it is slower. Use it for cross-account comparison runs,
# for example extracting a Japanese account after an English one.

[CmdletBinding()]
param(
    [string]$CliExe = 'D:\opt\AssetStudioModCLI_net472_win32_64\AssetStudioModCLI.exe',

# English (minahdao):
# [string]$InputRoot = 'D:\game\SteamLibrary\steamapps\common\Yu-Gi-Oh!  Master Duel\LocalData\70102374\0000',
# # Japanese OCG (bixuzofa):
    [string]$InputRoot = 'D:\game\SteamLibrary\steamapps\common\Yu-Gi-Oh!  Master Duel\LocalData\2509bcbc\0000', # Japanese OCG (bixuzofa)

# Second, smaller asset store shipped with the game
# (the bootstrap/tutorial set, ~100 MB).
# It also contains card art under an "assets/resources/card/..." container,
# so it is exported in one final pass after the LocalData buckets.
# Set to '' to skip this pass.
    [string]$StreamingAssetsRoot = 'D:\game\SteamLibrary\steamapps\common\Yu-Gi-Oh!  Master Duel\masterduel_Data\StreamingAssets\AssetBundle',

# Third, account-independent store: the assets compiled into the Unity player.
# It holds what the game needs before any bundle loads,
# including the per card type frames (card_frame00..card_frame19, card_frame_ext),
# each shipped at 704x1024 and 480x700.
# The sibling "resources.resource" holds the streamed texture bytes;
# AssetStudio finds it on its own because it sits next to data.unity3d.
# These assets carry no container path, so they export flat.
# The pass therefore writes into a "data_unity3d" subfolder of the output
# instead of scattering ~1800 loose files across the output root.
# Set to '' to skip this pass.
    [string]$GameDataFile = 'D:\game\SteamLibrary\steamapps\common\Yu-Gi-Oh!  Master Duel\masterduel_Data\data.unity3d',

# Output root: point this at the parent of the "assets" folder,
# not at "...\assets" itself.
# The CLI container paths already begin with "assets/",
# so AssetStudio builds the "assets\..." tree under this root on its own.
# Pointing at "...\tmp_process_MD_file_by_path" yields a single
# "assets\resources\card\..." tree;
# pointing at "...\assets" would nest a redundant "assets\assets\..." level.
#
# Fills the existing extraction in place:
# without -r, files that already exist are skipped,
# so only the assets missed by a previous out-of-memory crash get written.
    [string]$OutputDir = 'D:\tmp_process_MD_file_by_path',

# Comma or semicolon separated asset types (see CLI -t help).
# Default: all types present in the data except MonoBehaviour and Animator.
    [string]$Types = 'tex2d,sprite,textAsset,audio,mesh',

# Concurrent texture decodes per run. Lower means less peak memory.
# 0 means auto: use every logical processor.
# Must NOT exceed the logical CPU count (physical cores times threads-per-core,
# i.e. what [Environment]::ProcessorCount reports, hyperthreads included),
# because the CLI validates against Environment.ProcessorCount.
# Worse, on a too-high value the CLI prints a parse error but still exits 0,
# so an over-large number would silently extract nothing.
# The value is validated and clamped to that ceiling below.
    [int]$MaxExportTasks = 0,

# When set, keep content-different variants instead of skipping by file name.
# See the header for the staging and reconcile behavior.
    [switch]$KeepDifferentVariants
)

if (-not (Test-Path -LiteralPath $CliExe)) {
    throw "AssetStudioModCLI.exe not found at: $CliExe  (edit the -CliExe path)"
}
if (-not (Test-Path -LiteralPath $InputRoot)) {
    throw "Input root not found: $InputRoot"
}
New-Item -ItemType Directory -Force -Path $OutputDir | Out-Null

# Clamp the decode task count to the logical CPU count.
# See the -MaxExportTasks note above: the CLI rejects a larger value with a
# parse error but still exits 0, which would look like a clean run that
# extracted nothing.
$cores = [Environment]::ProcessorCount
if ($MaxExportTasks -le 0) {
    $MaxExportTasks = $cores
}
elseif ($MaxExportTasks -gt $cores) {
    Write-Warning "MaxExportTasks $MaxExportTasks exceeds logical CPUs ($cores); clamping to $cores."
    $MaxExportTasks = $cores
}

# Staging folder for -KeepDifferentVariants, kept on the same drive as the
# output so that reconciling moves files (a rename) instead of copying them.
$StageRoot = "${OutputDir}_staging"
$Stamp = Get-Date -Format 'yyyyMMdd'

# One CLI export run over a single input path
# (a bucket folder or the whole StreamingAssets store) into a chosen folder.
# Returns $true on success, $false on non-zero exit.
function Invoke-Export([string]$inputPath, [string]$outDir) {
    & $CliExe $inputPath `
        -m export `
        -t $Types `
        -g container `
        --max-export-tasks $MaxExportTasks `
        -o $outDir
    return ($LASTEXITCODE -eq 0)
}

# Reconcile a freshly exported staging tree into the output tree by content.
# See the header for the three cases (missing, identical, different).
function Merge-Staging([string]$stageDir, [string]$destRoot, [string]$stamp) {
    $prefix = (Resolve-Path -LiteralPath $stageDir).Path.TrimEnd('\') + '\'
    Get-ChildItem -LiteralPath $stageDir -Recurse -File | ForEach-Object {
        $rel = $_.FullName.Substring($prefix.Length)
        $dest = Join-Path $destRoot $rel
        $destDir = Split-Path -Parent $dest

        if (-not (Test-Path -LiteralPath $dest)) {
            New-Item -ItemType Directory -Force -Path $destDir | Out-Null
            Move-Item -LiteralPath $_.FullName -Destination $dest
            return
        }

        $newHash = (Get-FileHash -LiteralPath $_.FullName -Algorithm MD5).Hash
        if ($newHash -eq (Get-FileHash -LiteralPath $dest -Algorithm MD5).Hash) {
            Remove-Item -LiteralPath $_.FullName -Force   # identical, skip
            return
        }

        # Different content at the same path:
        # keep it as a dated variant, sidestepping same-day collisions.
        $base = [System.IO.Path]::GetFileNameWithoutExtension($dest)
        $ext = [System.IO.Path]::GetExtension($dest)
        $variant = Join-Path $destDir ("{0}_d{1}{2}" -f $base, $stamp, $ext)
        $n = 1
        while (Test-Path -LiteralPath $variant) {
            if ($newHash -eq (Get-FileHash -LiteralPath $variant -Algorithm MD5).Hash) {
                Remove-Item -LiteralPath $_.FullName -Force   # this variant already saved
                return
            }
            $variant = Join-Path $destDir ("{0}_d{1}_{2}{3}" -f $base, $stamp, $n, $ext)
            $n++
        }
        Move-Item -LiteralPath $_.FullName -Destination $variant
    }
}

# AssetStudio appends "_#<pathID>" when two assets share a container name,
# and that number changes between runs,
# so the skip-by-name check cannot stop a re-run from writing
# a byte-identical copy of a file already there under a different suffix.
# This collapses each "<name>_#<pathID>" group down to one file per content.
# Only files this run wrote are ever deleted,
# and only groups this run wrote into are hashed,
# so the cost follows the run rather than the size of the tree.
# Length is compared first, so a hash is computed only for a same-size pair.
# Returns the number of files removed.
function Remove-DuplicateVariants([string]$root, [datetime]$since) {
    $sw = [System.Diagnostics.Stopwatch]::StartNew()
    $files = @(Get-ChildItem -LiteralPath $root -Recurse -File -ErrorAction SilentlyContinue)
    Write-Host ("[dedupe] scanned {0} file(s) in {1:mm\:ss}" -f $files.Count, $sw.Elapsed)

    # Group by name with the "_#<pathID>" suffix stripped,
    # tracking which groups this run wrote into.
    # Only those are hashed, so the cost follows what the run produced,
    # not the size of the whole tree.
    $byGroup = @{ }
    $touched = @{ }
    foreach ($file in $files) {
        $key = $file.DirectoryName + '\' + [regex]::Replace($file.Name, '_#\d+(?=(\.[^.\\]+)?$)', '')
        $group = $byGroup[$key]
        if ($null -eq $group) {
            $group = New-Object System.Collections.ArrayList
            $byGroup[$key] = $group
        }
        [void]$group.Add($file)
        if ($file.LastWriteTime -ge $since) { $touched[$key] = $true }
    }

    $candidates = @($touched.Keys | Where-Object { $byGroup[$_].Count -gt 1 })
    Write-Host ("[dedupe] {0} group(s) touched this run, {1} of them hold more than one file" -f $touched.Count, $candidates.Count)

    $removed = 0
    $hashed = 0
    $done = 0
    foreach ($key in $candidates) {
        $done++
        if ($done % 500 -eq 0) {
            Write-Host ("[dedupe] {0}/{1} group(s), {2} hashed, {3} removed" -f $done, $candidates.Count, $hashed, $removed)
        }
        foreach ($sameSize in ($byGroup[$key] | Group-Object Length)) {
            if ($sameSize.Count -lt 2) { continue }

            # Plain names first, then oldest first,
            # so the copy that survives is the one already in the tree.
            $ordered = $sameSize.Group |
                Sort-Object @{ Expression = { $_.Name -match '_#\d+' } }, LastWriteTime
            $seen = @{}
            foreach ($file in $ordered) {
                $hash = (Get-FileHash -LiteralPath $file.FullName -Algorithm MD5).Hash
                $hashed++
                if (-not $seen.ContainsKey($hash)) {
                    $seen[$hash] = $true
                    continue
                }
                if ($file.LastWriteTime -ge $since) {
                    Remove-Item -LiteralPath $file.FullName -Force
                    $removed++
                }
            }
        }
    }
    Write-Host ("[dedupe] hashed {0} file(s), removed {1}, in {2:mm\:ss}" -f $hashed, $removed, $sw.Elapsed)
    return $removed
}

# Export one input, either straight to the output tree,
# or through the staging folder when keeping content-different variants.
# Returns $true on success, $false on non-zero CLI exit.
function Export-OneInput([string]$inputPath, [string]$destRoot = $OutputDir) {
    if (-not $KeepDifferentVariants) {
        return (Invoke-Export $inputPath $destRoot)
    }
    if (Test-Path -LiteralPath $StageRoot) {
        Remove-Item -Recurse -Force -LiteralPath $StageRoot
    }
    New-Item -ItemType Directory -Force -Path $StageRoot | Out-Null
    $ok = Invoke-Export $inputPath $StageRoot
    Merge-Staging $StageRoot $destRoot $Stamp
    return $ok
}

# Each immediate subfolder of InputRoot is one chunk (00..ff and root).
$buckets = Get-ChildItem -LiteralPath $InputRoot -Directory | Sort-Object Name
$total = $buckets.Count
Write-Host "Found $total bucket(s) under `"$InputRoot`""
Write-Host "Output: $OutputDir"
Write-Host "MaxExportTasks: $MaxExportTasks (logical CPUs: $cores)`n"

$i = 0
$failed = @()
# Cutoff for counting files written during this run (see the summary below).
$runStart = Get-Date
# Duration of each step, printed as a table at the end.
$stepTimes = [ordered]@{ }
$swTotal = [System.Diagnostics.Stopwatch]::StartNew()

$swStep = [System.Diagnostics.Stopwatch]::StartNew()
foreach ($bucket in $buckets) {
    $i++
    # A bucket run prints nothing until the CLI starts,
    # so carry the elapsed time and a projection from the average bucket so far.
    if ($i -gt 1) {
        $left = [TimeSpan]::FromTicks([long]($swStep.Elapsed.Ticks / ($i - 1) * ($total - $i + 1)))
    }
    else {
        $left = [TimeSpan]::Zero
    }
    Write-Host ("[{0,3}/{1}] elapsed {2:hh\:mm\:ss}, left ~{3:hh\:mm\:ss}, bucket {4} ..." -f $i, $total, $swStep.Elapsed, $left, $bucket.Name)

    if (-not (Export-OneInput $bucket.FullName)) {
        Write-Warning "bucket $( $bucket.Name ) exited with a non-zero code"
        $failed += $bucket.Name
    }
}
$swStep.Stop()
if ($total -gt 0) { $stepTimes['buckets'] = $swStep.Elapsed }

# Second pass: the StreamingAssets store is small (~100 MB),
# so it loads fine in one run, no chunking needed.
# Its card art shares the same "assets/resources/card/..." container as LocalData,
# so it merges into the same output tree;
# without -r, existing files are kept, so it only fills gaps.
if ($StreamingAssetsRoot -and (Test-Path -LiteralPath $StreamingAssetsRoot)) {
    Write-Host "`n[StreamingAssets] $StreamingAssetsRoot ..."
    $swStep = [System.Diagnostics.Stopwatch]::StartNew()
    if (-not (Export-OneInput $StreamingAssetsRoot)) {
        Write-Warning "StreamingAssets pass exited with a non-zero code"
        $failed += 'StreamingAssets'
    }
    $swStep.Stop()
    $stepTimes['StreamingAssets'] = $swStep.Elapsed
    Write-Host ("[StreamingAssets] done in {0:hh\:mm\:ss}" -f $swStep.Elapsed)
}
elseif ($StreamingAssetsRoot) {
    Write-Warning "StreamingAssets root not found, skipping: $StreamingAssetsRoot"
}

# Third pass: the store built into the Unity player.
# One file, small enough to load in one run.
# Its assets have no container path, so they land flat
# in their own subfolder rather than in the "assets\..." tree.
if ($GameDataFile -and (Test-Path -LiteralPath $GameDataFile)) {
    $gameDataOut = Join-Path $OutputDir 'data_unity3d'
    New-Item -ItemType Directory -Force -Path $gameDataOut | Out-Null
    Write-Host "`n[data.unity3d] $GameDataFile ..."
    $swStep = [System.Diagnostics.Stopwatch]::StartNew()
    if (-not (Export-OneInput $GameDataFile $gameDataOut)) {
        Write-Warning "data.unity3d pass exited with a non-zero code"
        $failed += 'data.unity3d'
    }
    $swStep.Stop()
    $stepTimes['data.unity3d'] = $swStep.Elapsed
    Write-Host ("[data.unity3d] done in {0:hh\:mm\:ss}, {1} file(s) in {2}" -f
            $swStep.Elapsed, (Get-ChildItem -LiteralPath $gameDataOut -File).Count, $gameDataOut)
}
elseif ($GameDataFile) {
    Write-Warning "data.unity3d file not found, skipping: $GameDataFile"
}

# Drop the staging folder if -KeepDifferentVariants created one.
if (Test-Path -LiteralPath $StageRoot) {
    Remove-Item -Recurse -Force -LiteralPath $StageRoot
}

Write-Host "`n[dedupe] removing byte-identical copies written this run ..."
$swStep = [System.Diagnostics.Stopwatch]::StartNew()
$dupes = Remove-DuplicateVariants $OutputDir $runStart
$swStep.Stop()
$stepTimes['dedupe'] = $swStep.Elapsed
$swTotal.Stop()

# Count files written to the output tree during this run: new extractions and,
# in -KeepDifferentVariants mode, kept variants. Move-Item preserves each file's
# write time from when the CLI created it in staging, so a single timestamp
# cutoff catches both direct writes and reconciled moves.
$newCount = (Get-ChildItem -LiteralPath $OutputDir -Recurse -File -ErrorAction SilentlyContinue |
    Where-Object { $_.LastWriteTime -ge $runStart } | Measure-Object).Count

# The step table below names the passes that actually ran,
# so the headline stays true when a pass is skipped.
Write-Host ("`nDone in {0:hh\:mm\:ss}, {1} bucket(s):" -f $swTotal.Elapsed, $total)
foreach ($step in $stepTimes.Keys) {
    Write-Host ("  {0,-16} {1:hh\:mm\:ss}" -f $step, $stepTimes[$step])
}
Write-Host ("  {0,-16} {1:hh\:mm\:ss}" -f 'total', $swTotal.Elapsed)
Write-Host "New files extracted this run: $newCount"
Write-Host "Duplicate copies removed: $dupes"
if ($failed.Count -gt 0) {
    Write-Warning "Passes with non-zero exit: $( $failed -join ', ' )"
}
else {
    Write-Host "All passes exited cleanly."
}
