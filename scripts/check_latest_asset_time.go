package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// this program reports, for each first-level child of the input dir (one per
// Steam account under LocalData), the newest file found anywhere below it, plus
// that child's file count and total size. Run it before an extraction to see
// which accounts have new assets; point --input at the output tree for the
// other side of that comparison. Rows are sorted newest first.
//
// Example output, taken 2026-07-28, the day after Selection Pack
// "Call of Eternity" released:
//
//	2509bcbc  2026-07-28 13:32     38883 files   12.60 GiB  0000/32/32782cb5
//	70102374  2026-07-27 21:28     38805 files   12.56 GiB  0000/32/32782cb5
//
// The rows sit a day apart because a pack downloads in-game, into whichever
// account is logged in: 70102374 got it on release night, 2509bcbc only when
// opened the next afternoon. Check both rows before extracting, or the run
// captures the pack in one region only.
//
// Usage:
//
//	go run scripts/check_latest_asset_time.go
//	go run scripts/check_latest_asset_time.go --input D:\tmp_process_MD_file_by_path\assets

// Inputs (edit for each run, or pass --input):
const (
	// One child per Steam account. Note the two spaces in "Yu-Gi-Oh!  Master Duel".
	dirInputDefault = `D:\game\SteamLibrary\steamapps\common\Yu-Gi-Oh!  Master Duel\LocalData`
)

func main() {
	dirInput := dirInputDefault
	// Exit codes mirror the shell script: 2 usage error, 1 missing input dir.
	for i := 1; i < len(os.Args); i++ {
		switch os.Args[i] {
		case "--input":
			if i+1 >= len(os.Args) {
				usage()
			}
			dirInput = os.Args[i+1]
			i++
		case "--help":
			usage()
		default:
			fmt.Fprintf(os.Stderr, "unknown option: %v\n", os.Args[i])
			usage()
		}
	}

	children, err := os.ReadDir(dirInput)
	if err != nil {
		fmt.Fprintf(os.Stderr, "input dir not found: %v: %v\n", dirInput, err)
		os.Exit(1)
	}
	if len(children) == 0 {
		fmt.Fprintf(os.Stderr, "no children under: %v\n", dirInput)
		os.Exit(1)
	}

	rows := make([]ChildStat, 0, len(children))
	for _, child := range children {
		row, err := scanChild(filepath.Join(dirInput, child.Name()), child)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error scanChild %v: %v\n", child.Name(), err)
			os.Exit(1)
		}
		rows = append(rows, row)
	}

	sort.Slice(rows, func(i, j int) bool {
		return rows[i].NewestTime.After(rows[j].NewestTime)
	})

	// Pad the name column so the times line up.
	nameWidth := 0
	for _, row := range rows {
		if len(row.Name) > nameWidth {
			nameWidth = len(row.Name)
		}
	}

	fmt.Println(dirInput)
	fmt.Println()
	for _, row := range rows {
		when := "(no files)"
		if row.FileCount > 0 {
			when = row.NewestTime.Format("2006-01-02 15:04")
		}
		fmt.Printf("%-*s  %-16s  %8d files  %10s  %s\n",
			nameWidth, row.Name, when, row.FileCount,
			humanSize(row.TotalBytes), row.NewestPath)
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: go run scripts/check_latest_asset_time.go [--input <input_dir>]\n")
	os.Exit(2)
}

// scanChild walks one first-level child for its newest file, file count and
// total size. d.Info() costs nothing: on Windows those values already come with
// the directory entry.
func scanChild(path string, entry fs.DirEntry) (ChildStat, error) {
	row := ChildStat{Name: entry.Name()}

	// A plain file child stands for itself.
	if !entry.IsDir() {
		info, err := entry.Info()
		if err != nil {
			return row, fmt.Errorf("error entry.Info: %w", err)
		}
		row.FileCount, row.TotalBytes = 1, info.Size()
		row.NewestTime, row.NewestPath = info.ModTime(), entry.Name()
		return row, nil
	}

	err := filepath.WalkDir(path, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("error WalkDir %v: %w", p, err)
		}
		if d.IsDir() {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return fmt.Errorf("error d.Info %v: %w", p, err)
		}
		row.FileCount++
		row.TotalBytes += info.Size()
		if info.ModTime().After(row.NewestTime) {
			row.NewestTime = info.ModTime()
			relative, err := filepath.Rel(path, p)
			if err != nil {
				return fmt.Errorf("error filepath.Rel %v: %w", p, err)
			}
			// Forward slashes to match the shell version's output.
			row.NewestPath = filepath.ToSlash(relative)
		}
		return nil
	})
	if err != nil {
		return row, fmt.Errorf("error filepath.WalkDir: %w", err)
	}
	return row, nil
}

func humanSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	value, exponent := float64(bytes)/unit, 0
	for value >= unit && exponent < 3 {
		value /= unit
		exponent++
	}
	return fmt.Sprintf("%.2f %ciB", value, "KMGT"[exponent])
}

// ChildStat is the aggregate for one first-level child of the input directory.
type ChildStat struct {
	Name       string
	FileCount  int
	TotalBytes int64
	NewestTime time.Time
	NewestPath string
}
