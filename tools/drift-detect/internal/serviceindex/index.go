// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

// Package serviceindex fetches and caches the list of AWS service directory
// names from the aws/api-models-aws GitHub repository and resolves a
// Terraform service name to the correct model file path.
//
// The service directory list is cached at <cacheDir>/service-index.json and
// is considered fresh for 24 hours. Passing refresh=true to Load deletes the
// cache file so it is regenerated on the next run.
package serviceindex

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"time"
)

const (
	cacheTTL        = 24 * time.Hour
	cacheFileName   = "service-index.json"
	githubAPIBase   = "https://api.github.com/repos/aws/api-models-aws/contents"
	httpTimeout     = 30 * time.Second
)

// Index holds the cached list of AWS service directory names and the
// services rename map from the mapping file.
type Index struct {
	dirs     []string          // sorted list of directory names under models/
	services map[string]string // TF service name → AWS directory name overrides
}

// githubEntry is one element of the GitHub Contents API response.
type githubEntry struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// Load loads (or refreshes) the service index.
//
//   - If refresh is true the existing cache file is deleted before proceeding.
//   - If the cache file exists and its modification time is within 24 hours it
//     is parsed and returned without any network call.
//   - Otherwise the GitHub Contents API is queried, the result is written to
//     the cache file, and the populated Index is returned.
//
// cacheDir is typically ".cache"; apiBaseURL is typically the raw GitHub
// base URL (used only to satisfy the interface — directory listing always
// uses the GitHub API, not the raw URL).
// services is the rename map from awsmapping.File.Services.
func Load(cacheDir, _ string, services map[string]string, refresh bool) (*Index, error) {
	cachePath := filepath.Join(cacheDir, cacheFileName)

	if refresh {
		_ = os.Remove(cachePath)
	}

	dirs, err := loadFromCacheOrFetch(cachePath)
	if err != nil {
		return nil, err
	}

	return &Index{dirs: dirs, services: services}, nil
}

// loadFromCacheOrFetch returns the directory list from the cache file if it
// is fresh, or fetches it from the GitHub API and writes a new cache file.
func loadFromCacheOrFetch(cachePath string) ([]string, error) {
	if info, err := os.Stat(cachePath); err == nil {
		if time.Since(info.ModTime()) < cacheTTL {
			return readCache(cachePath)
		}
	}
	return fetchAndCache(cachePath)
}

// readCache parses the cached JSON file and returns the directory list.
func readCache(cachePath string) ([]string, error) {
	data, err := os.ReadFile(cachePath)
	if err != nil {
		return nil, fmt.Errorf("reading service index cache: %w", err)
	}
	var dirs []string
	if err := json.Unmarshal(data, &dirs); err != nil {
		return nil, fmt.Errorf("parsing service index cache: %w", err)
	}
	return dirs, nil
}

// fetchAndCache calls the GitHub Contents API to list models/ directories,
// writes the result to cachePath, and returns the directory names.
func fetchAndCache(cachePath string) ([]string, error) {
	dirs, err := fetchDirs(githubAPIBase + "/models")
	if err != nil {
		return nil, err
	}

	if err := writeCache(cachePath, dirs); err != nil {
		// Non-fatal: log but continue — the index is still usable in memory.
		fmt.Fprintf(os.Stderr, "warning: could not write service index cache: %v\n", err)
	}
	return dirs, nil
}

// fetchDirs calls the GitHub Contents API at url and returns the names of all
// entries whose type is "dir".
func fetchDirs(url string) ([]string, error) {
	client := &http.Client{Timeout: httpTimeout}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("building request for %s: %w", url, err)
	}
	// Request JSON from the GitHub API.
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching service index: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("service index URL not found: %s", url)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetching service index: unexpected status %s", resp.Status)
	}

	var entries []githubEntry
	if err := json.NewDecoder(resp.Body).Decode(&entries); err != nil {
		return nil, fmt.Errorf("decoding service index response: %w", err)
	}

	var dirs []string
	for _, e := range entries {
		if e.Type == "dir" {
			dirs = append(dirs, e.Name)
		}
	}
	sort.Strings(dirs)
	return dirs, nil
}

// writeCache serialises dirs as JSON and writes it to cachePath, creating
// the parent directory if necessary.
func writeCache(cachePath string, dirs []string) error {
	if err := os.MkdirAll(filepath.Dir(cachePath), 0o755); err != nil {
		return fmt.Errorf("creating cache directory: %w", err)
	}
	data, err := json.Marshal(dirs)
	if err != nil {
		return fmt.Errorf("marshalling service index: %w", err)
	}
	return os.WriteFile(cachePath, data, 0o644)
}

// Resolve maps a TF service name to the matching AWS directory name.
//
// Resolution order:
//  1. If the services rename map has an entry for tfServiceName, that value is
//     used as the candidate AWS name.
//  2. Otherwise tfServiceName is used as the candidate.
//
// The candidate is then checked against the directory list. An error is
// returned if it is not present.
func (idx *Index) Resolve(tfServiceName string) (string, error) {
	candidate := tfServiceName
	if renamed, ok := idx.services[tfServiceName]; ok {
		candidate = renamed
	}

	for _, d := range idx.dirs {
		if d == candidate {
			return candidate, nil
		}
	}

	if candidate != tfServiceName {
		return "", fmt.Errorf(
			"service %q (mapped from TF name %q) not found in AWS model directory listing",
			candidate, tfServiceName,
		)
	}
	return "", fmt.Errorf(
		"service %q not found in AWS model directory listing; add a rename entry to the services: block in aws_resources.yaml if the AWS directory name differs",
		tfServiceName,
	)
}

// ResolveModelPath queries the GitHub Contents API for the date subfolder of
// awsServiceName and returns the relative model file path and derived
// Smithy namespace.
//
// When multiple date folders exist the lexicographically last one (most
// recent) is used.
func ResolveModelPath(awsServiceName, apiBaseURL string) (modelPath, namespace string, err error) {
	_ = apiBaseURL // the date-folder query always uses the GitHub API
	url := fmt.Sprintf("%s/models/%s/service", githubAPIBase, awsServiceName)

	dates, err := fetchDirs(url)
	if err != nil {
		return "", "", fmt.Errorf("resolving model path for %q: %w", awsServiceName, err)
	}
	if len(dates) == 0 {
		return "", "", fmt.Errorf("no service date folders found for %q at %s", awsServiceName, url)
	}

	// fetchDirs already sorts; the last element is the most recent date.
	date := dates[len(dates)-1]

	modelPath = fmt.Sprintf("models/%s/service/%s/%s-%s.json", awsServiceName, date, awsServiceName, date)
	namespace = fmt.Sprintf("com.amazonaws.%s", awsServiceName)
	return modelPath, namespace, nil
}
