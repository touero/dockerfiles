// Package store persists pull-count snapshots as a CSV file.
//
// The file uses "long" format so that adding or removing repositories never
// changes the schema:
//
//	date,repo,pull_count,star_count
//	2025-01-01,fileserver,100,2
//	2025-01-01,gitweb,50,1
package store

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"time"
)

// DateFormat is the "day" granularity used for snapshots.
const DateFormat = "2006-01-02"

var header = []string{"date", "repo", "pull_count", "star_count"}

// Record is a single snapshot of one repository on one day.
type Record struct {
	Date      time.Time
	Repo      string
	PullCount int64
	StarCount int64
}

// Load reads all records from path. A missing file is not an error: it simply
// means there is no history yet.
func Load(path string) ([]Record, error) {
	f, err := os.Open(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer f.Close()

	r := csv.NewReader(f)
	first := true
	var records []Record
	for {
		row, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if first { // skip header
			first = false
			continue
		}
		if len(row) < 4 {
			continue
		}
		date, err := time.Parse(DateFormat, row[0])
		if err != nil {
			continue
		}
		pulls, err := strconv.ParseInt(row[2], 10, 64)
		if err != nil {
			continue
		}
		stars, err := strconv.ParseInt(row[3], 10, 64)
		if err != nil {
			stars = 0
		}
		records = append(records, Record{Date: date, Repo: row[1], PullCount: pulls, StarCount: stars})
	}
	return records, nil
}

// Save writes records to path, sorted by date then repo.
func Save(path string, records []Record) error {
	sort.Slice(records, func(i, j int) bool {
		if !records[i].Date.Equal(records[j].Date) {
			return records[i].Date.Before(records[j].Date)
		}
		return records[i].Repo < records[j].Repo
	})

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	w := csv.NewWriter(f)
	if err := w.Write(header); err != nil {
		return err
	}
	for _, rec := range records {
		row := []string{
			rec.Date.Format(DateFormat),
			rec.Repo,
			strconv.FormatInt(rec.PullCount, 10),
			strconv.FormatInt(rec.StarCount, 10),
		}
		if err := w.Write(row); err != nil {
			return err
		}
	}
	w.Flush()
	return w.Error()
}

// Upsert replaces the records for the given day (for the given repos) and
// returns the merged result, so re-running on the same day is idempotent.
func Upsert(existing []Record, day time.Time, fresh []Record, repos []string) []Record {
	wanted := make(map[string]bool, len(repos))
	for _, r := range repos {
		wanted[r] = true
	}
	day = truncate(day)

	merged := make([]Record, 0, len(existing)+len(fresh))
	for _, rec := range existing {
		if truncate(rec.Date).Equal(day) && wanted[rec.Repo] {
			continue
		}
		merged = append(merged, rec)
	}
	for _, rec := range fresh {
		rec.Date = day
		merged = append(merged, rec)
	}
	return merged
}

func truncate(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}

// String is a small helper for logging a record.
func (r Record) String() string {
	return fmt.Sprintf("%s %s: pulls=%d stars=%d", r.Date.Format(DateFormat), r.Repo, r.PullCount, r.StarCount)
}
