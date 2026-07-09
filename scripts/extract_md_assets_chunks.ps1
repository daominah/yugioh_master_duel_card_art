# extract_md_assets_chunks.ps1
#
# Runs AssetStudioModCLI once per hex bucket of the Master Duel LocalData store.
# Each invocation loads only ~150 bundles, exports,
# then exits and frees all memory before the next bucket.
# This keeps peak RAM small and avoids the out-of-memory crash
# that happens when AssetStudio loads all ~38k bundles (13 GB) at once.
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
#   output missing         -> moved in as is.
#   identical content       -> dropped (skip).
#   same path, new content  -> kept as "<name>_dYYYYMMDD<ext>",
#                              so a second account's different art for the same
#                              card id is preserved next to the first.
# This mode re-extracts every file each run (it gives up the fast file-name
# skip), so it is slower. Use it for cross-account comparison runs,
# for example extracting a Japanese account after an English one.

[CmdletBinding()]
param(
    [string]$CliExe = 'D:\opt\AssetStudioModCLI_net472_win32_64\AssetStudioModCLI.exe',

    # English (minahdao):
    # [string]$InputRoot = 'D:\game\SteamLibrary\steamapps\common\Yu-Gi-Oh!  Master Duel\LocalData\70102374\0000',
    # # Japanese OCG (bixuzofa):
    [string]$InputRoot = 'D:\game\SteamLibrary\steamapps\common\Yu-Gi-Oh!  Master Duel\LocalData\2509bcbc\0000',  # Japanese OCG (bixuzofa)

    # Second, smaller asset store shipped with the game
    # (the bootstrap/tutorial set, ~100 MB).
    # It also contains card art under an "assets/resources/card/..." container,
    # so it is exported in one final pass after the LocalData buckets.
    # Set to '' to skip this pass.
    [string]$StreamingAssetsRoot = 'D:\game\SteamLibrary\steamapps\common\Yu-Gi-Oh!  Master Duel\masterduel_Data\StreamingAssets\AssetBundle',

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
    [int]$MaxExportTasks = 2,

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
        $rel  = $_.FullName.Substring($prefix.Length)
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
        $ext  = [System.IO.Path]::GetExtension($dest)
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

# Export one input, either straight to the output tree,
# or through the staging folder when keeping content-different variants.
# Returns $true on success, $false on non-zero CLI exit.
function Export-OneInput([string]$inputPath) {
    if (-not $KeepDifferentVariants) {
        return (Invoke-Export $inputPath $OutputDir)
    }
    if (Test-Path -LiteralPath $StageRoot) {
        Remove-Item -Recurse -Force -LiteralPath $StageRoot
    }
    New-Item -ItemType Directory -Force -Path $StageRoot | Out-Null
    $ok = Invoke-Export $inputPath $StageRoot
    Merge-Staging $StageRoot $OutputDir $Stamp
    return $ok
}

# Each immediate subfolder of InputRoot is one chunk (00..ff and root).
$buckets = Get-ChildItem -LiteralPath $InputRoot -Directory | Sort-Object Name
$total = $buckets.Count
Write-Host "Found $total bucket(s) under `"$InputRoot`""
Write-Host "Output -> $OutputDir`n"

$i = 0
$failed = @()
$swTotal = [System.Diagnostics.Stopwatch]::StartNew()
foreach ($bucket in $buckets) {
    $i++
    Write-Host ("[{0,3}/{1}] bucket {2} ..." -f $i, $total, $bucket.Name)

    if (-not (Export-OneInput $bucket.FullName)) {
        Write-Warning "bucket $($bucket.Name) exited with a non-zero code"
        $failed += $bucket.Name
    }
}

# Second pass: the StreamingAssets store is small (~100 MB),
# so it loads fine in one run, no chunking needed.
# Its card art shares the same "assets/resources/card/..." container as LocalData,
# so it merges into the same output tree;
# without -r, existing files are kept, so it only fills gaps.
if ($StreamingAssetsRoot -and (Test-Path -LiteralPath $StreamingAssetsRoot)) {
    Write-Host "`n[StreamingAssets] $StreamingAssetsRoot ..."
    if (-not (Export-OneInput $StreamingAssetsRoot)) {
        Write-Warning "StreamingAssets pass exited with a non-zero code"
        $failed += 'StreamingAssets'
    }
}
elseif ($StreamingAssetsRoot) {
    Write-Warning "StreamingAssets root not found, skipping: $StreamingAssetsRoot"
}
$swTotal.Stop()

# Drop the staging folder if -KeepDifferentVariants created one.
if (Test-Path -LiteralPath $StageRoot) {
    Remove-Item -Recurse -Force -LiteralPath $StageRoot
}

Write-Host ("`nDone in {0:hh\:mm\:ss}. Processed {1} bucket(s) + StreamingAssets." -f $swTotal.Elapsed, $total)
if ($failed.Count -gt 0) {
    Write-Warning "Passes with non-zero exit: $($failed -join ', ')"
} else {
    Write-Host "All passes exited cleanly."
}
