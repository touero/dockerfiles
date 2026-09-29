// Command plot renders Docker Hub pull-count trends from the collected CSV.
//
// It writes modern SVG charts into OUT_DIR, in a light and a dark variant:
//
//	pull-counts.svg / pull-counts-dark.svg              cumulative pulls per repo
//	pull-counts-daily.svg / pull-counts-daily-dark.svg  new pulls per day
//
// Environment:
//
//	DATA_FILE  input CSV path, default "metrics/data/pull_counts.csv"
//	OUT_DIR    output directory, default "docs"
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/touero/dockerfiles/metrics/internal/chart"
	"github.com/touero/dockerfiles/metrics/internal/store"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "plot:", err)
		os.Exit(1)
	}
}

func run() error {
	dataFile := env("DATA_FILE", "metrics/data/pull_counts.csv")
	outDir := env("OUT_DIR", "docs")

	records, err := store.Load(dataFile)
	if err != nil {
		return fmt.Errorf("load %s: %w", dataFile, err)
	}
	if len(records) == 0 {
		return fmt.Errorf("no data in %s", dataFile)
	}

	series := group(records)
	repos := make([]string, 0, len(series))
	for repo := range series {
		repos = append(repos, repo)
	}
	sort.Strings(repos)

	cumulative := make([]chart.Series, 0, len(repos))
	for _, repo := range repos {
		cumulative = append(cumulative, chart.Series{Name: repo, Points: points(series[repo], false)})
	}
	daily := make([]chart.Series, 0, len(repos))
	for _, repo := range repos {
		if pts := points(series[repo], true); len(pts) > 0 {
			daily = append(daily, chart.Series{Name: repo, Points: pts})
		}
	}

	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}

	variants := []struct {
		suffix string
		dark   bool
	}{{"", false}, {"-dark", true}}

	for _, v := range variants {
		cum := chart.Render(chart.Options{
			Title:     "Docker Hub pulls (cumulative)",
			YLabel:    "pull count",
			Series:    cumulative,
			LogY:      true,
			EndLabels: true,
			Dark:      v.dark,
		})
		if err := write(filepath.Join(outDir, "pull-counts"+v.suffix+".svg"), cum); err != nil {
			return err
		}
		if len(daily) == 0 {
			continue
		}
		dly := chart.Render(chart.Options{
			Title:    "Docker Hub new pulls per day",
			YLabel:   "new pulls",
			Series:   daily,
			ZeroBase: true,
			Dark:     v.dark,
		})
		if err := write(filepath.Join(outDir, "pull-counts-daily"+v.suffix+".svg"), dly); err != nil {
			return err
		}
	}

	if len(daily) == 0 {
		fmt.Println("plot: not enough history for the daily chart, skipping")
	}
	fmt.Printf("charts written to %s\n", outDir)
	return nil
}

// group turns the flat record list into per-repo, time-ordered series.
func group(records []store.Record) map[string][]store.Record {
	series := map[string][]store.Record{}
	for _, rec := range records {
		series[rec.Repo] = append(series[rec.Repo], rec)
	}
	for repo := range series {
		points := series[repo]
		sort.Slice(points, func(i, j int) bool { return points[i].Date.Before(points[j].Date) })
		series[repo] = points
	}
	return series
}

// points converts records to chart points. When delta is true the values are
// the day-over-day increase instead of the cumulative count.
func points(records []store.Record, delta bool) []chart.Point {
	if !delta {
		pts := make([]chart.Point, 0, len(records))
		for _, rec := range records {
			pts = append(pts, chart.Point{X: rec.Date, Y: float64(rec.PullCount)})
		}
		return pts
	}
	if len(records) < 2 {
		return nil
	}
	pts := make([]chart.Point, 0, len(records)-1)
	for i := 1; i < len(records); i++ {
		d := records[i].PullCount - records[i-1].PullCount
		if d < 0 { // counters should never decrease, guard anyway
			d = 0
		}
		pts = append(pts, chart.Point{X: records[i].Date, Y: float64(d)})
	}
	return pts
}

func write(path, content string) error {
	if content == "" {
		return fmt.Errorf("empty chart for %s", path)
	}
	return os.WriteFile(path, []byte(content), 0o644)
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
