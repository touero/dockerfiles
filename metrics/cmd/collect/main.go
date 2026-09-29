// Command collect records today's Docker Hub pull/star counts into a CSV file.
//
// Environment:
//
//	DOCKERHUB_USERNAME (or DOCKERHUB_NAMESPACE)  required, Docker Hub namespace
//	REPOS        space separated repos, default "fileserver gitweb wechat"
//	DATA_FILE    output CSV path, default "metrics/data/pull_counts.csv"
package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/touero/dockerfiles/metrics/internal/hub"
	"github.com/touero/dockerfiles/metrics/internal/store"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "collect:", err)
		os.Exit(1)
	}
}

func run() error {
	namespace := firstNonEmpty(os.Getenv("DOCKERHUB_NAMESPACE"), os.Getenv("DOCKERHUB_USERNAME"))
	if namespace == "" {
		return fmt.Errorf("DOCKERHUB_USERNAME (or DOCKERHUB_NAMESPACE) is required")
	}
	repos := strings.Fields(env("REPOS", "fileserver gitweb wechat"))
	dataFile := env("DATA_FILE", "metrics/data/pull_counts.csv")
	if len(repos) == 0 {
		return fmt.Errorf("REPOS is empty")
	}

	existing, err := store.Load(dataFile)
	if err != nil {
		return fmt.Errorf("load %s: %w", dataFile, err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	client := hub.New()
	day := time.Now().UTC()
	var fresh []store.Record
	var errs []string
	for _, repo := range repos {
		meta, err := client.Repository(ctx, namespace, repo)
		if err != nil {
			fmt.Fprintf(os.Stderr, "collect: skip %s: %v\n", repo, err)
			errs = append(errs, repo)
			continue
		}
		rec := store.Record{Date: day, Repo: repo, PullCount: meta.PullCount, StarCount: meta.StarCount}
		fresh = append(fresh, rec)
		fmt.Println(rec.String())
	}

	if len(fresh) == 0 {
		return fmt.Errorf("no data collected (failed: %s)", strings.Join(errs, ", "))
	}

	merged := store.Upsert(existing, day, fresh, repos)
	if err := store.Save(dataFile, merged); err != nil {
		return fmt.Errorf("save %s: %w", dataFile, err)
	}
	fmt.Printf("wrote %d rows to %s\n", len(merged), dataFile)
	return nil
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
