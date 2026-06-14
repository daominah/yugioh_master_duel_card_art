# extract_md_animator.ps1
#
# Exports Animator assets from Master Duel as FBX (mode -m animator).
# This is a separate workflow from extract_md_assets_chunks.ps1, which handles
# self-contained image/sprite assets. Animators are different in three ways:
#
#   1. Mode, not type. Animators export via "-m animator" (FBX output),
#      there is no "-t animator" value.
#   2. Cross-bundle dependencies. An Animator references its AnimationClips,
#      Mesh, skeleton, and model object, which may live in other bundles.
#      So we cannot chunk per hex bucket the way the assets script does:
#      a bucket loaded alone would miss its dependencies and produce broken FBX.
#      Each source here is loaded whole instead.
#   3. Readability. Animator controllers and components are MonoBehaviour-backed.
#      For readable data, point -AssemblyFolder at IL2CPP dummy DLLs
#      (generated from GameAssembly.dll via Il2CppDumper). Without it, the
#      MonoBehaviour fields are unreadable, but the FBX geometry still exports.
#
# Where animators live (measured with -m info -t all):
#   masterduel_Data\data.unity3d                 ~2784 animators (UI, effects)
#   LocalData\<account>\0000\<bucket>            a few per bucket (card related)
#   StreamingAssets\AssetBundle                   none
# So data.unity3d is the main source and the default below.
#
# Memory note: data.unity3d is a single ~158 MB bundle and loads fine in one run.
# Adding the whole LocalData store (~13 GB) as a source can run out of memory,
# because animator binding holds the parsed asset graph in RAM and cannot be
# chunked safely. Add LocalData only if you accept that risk.
#
# Launch (execution policy blocks scripts before line 1, so bypass at launch):
#   powershell -ExecutionPolicy Bypass -File .\scripts\extract_md_animator.ps1

[CmdletBinding()]
param(
    [string]$CliExe = 'D:\opt\AssetStudioModCLI_net472_win32_64\AssetStudioModCLI.exe',

    # Input sources, each loaded whole (no per-bucket chunking, see header).
    # Default is data.unity3d, where almost all animators live.
    # To also pull card-related animators, append the LocalData root, but mind
    # the memory note above:
    #   'D:\game\SteamLibrary\steamapps\common\Yu-Gi-Oh!  Master Duel\LocalData\70102374\0000'
    [string[]]$Sources = @(
        'D:\game\SteamLibrary\steamapps\common\Yu-Gi-Oh!  Master Duel\masterduel_Data\data.unity3d'
    ),

    [string]$OutputDir = 'D:\tmp_process_MD_file_by_path_animator',

    # Optional path to IL2CPP dummy DLLs for readable MonoBehaviour data.
    # Leave '' to export without them (FBX geometry still works).
    [string]$AssemblyFolder = '',

    # Concurrent export tasks. Lower means less peak memory.
    [int]$MaxExportTasks = 2
)

if (-not (Test-Path -LiteralPath $CliExe)) {
    throw "AssetStudioModCLI.exe not found at: $CliExe  (edit the -CliExe path)"
}
New-Item -ItemType Directory -Force -Path $OutputDir | Out-Null

# One animator export run over a single input source (loaded whole).
# Returns $true on success, $false on non-zero exit.
function Invoke-AnimatorExport([string]$inputPath) {
    $cliArgs = @(
        $inputPath,
        '-m', 'animator',
        '-g', 'container',
        '--max-export-tasks', $MaxExportTasks,
        '-o', $OutputDir
    )
    if ($AssemblyFolder) {
        $cliArgs += @('--assembly-folder', $AssemblyFolder)
    }
    & $CliExe @cliArgs
    return ($LASTEXITCODE -eq 0)
}

$total = $Sources.Count
Write-Host "Animator export, $total source(s)."
Write-Host "Output -> $OutputDir`n"

$i = 0
$failed = @()
$swTotal = [System.Diagnostics.Stopwatch]::StartNew()
foreach ($source in $Sources) {
    $i++
    Write-Host ("[{0}/{1}] {2} ..." -f $i, $total, $source)

    if (-not (Test-Path -LiteralPath $source)) {
        Write-Warning "source not found, skipping: $source"
        $failed += $source
        continue
    }
    if (-not (Invoke-AnimatorExport $source)) {
        Write-Warning "source exited with code $LASTEXITCODE`: $source"
        $failed += $source
    }
}
$swTotal.Stop()

Write-Host ("`nDone in {0:hh\:mm\:ss}. Processed {1} source(s)." -f $swTotal.Elapsed, $total)
if ($failed.Count -gt 0) {
    Write-Warning "Sources with problems: $($failed -join ', ')"
} else {
    Write-Host "All sources exited cleanly."
}
