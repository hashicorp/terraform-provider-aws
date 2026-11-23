// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"sync"
	"time"
)

const serviceRefURL = "https://servicereference.us-east-1.amazonaws.com/"

var (
	// httpClient's timeout covers the whole request, including reading the
	// body, so a stalled response can't block a sync indefinitely.
	httpClient = &http.Client{Timeout: 60 * time.Second}

	// serviceNameRegexp matches service names that are safe to use as file
	// names: no path separators and no "." or ".." components.
	serviceNameRegexp = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]*$`)
)

type ServiceURL struct {
	Service  string `json:"service"`
	URL      string `json:"url"`
	Modified int    `json:"modified"`
}

// SyncServiceRefs brings the service reference files in dir and the index at
// indexPath up to date with the remote service reference list. Only services
// that are new or modified since the last sync are downloaded, and files for
// services no longer listed remotely are removed.
func SyncServiceRefs(dir, indexPath string) error {
	return syncServiceRefs(serviceRefURL, dir, indexPath)
}

func syncServiceRefs(indexURL, dir, indexPath string) error {
	cache := LoadCache(dir, indexPath)
	fmt.Printf("#### Loaded %d cached services\n", len(cache))

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	fmt.Println("#### Fetching service reference urls")
	urls, err := GetServiceRefURLs(indexURL)
	fmt.Println("#### Completed fetching service reference urls")
	if err != nil {
		return err
	}

	urlCount := len(urls)
	fmt.Printf("service urls found: %d\n", urlCount)

	var (
		wg      sync.WaitGroup
		mu      sync.Mutex
		errs    []error
		current = make(map[string]ServiceURL, urlCount) // entries whose file on disk is valid
		fetched int
		cached  int
	)
	for _, su := range urls {
		wg.Go(func() {
			if !needsFetch(su, cache, dir) {
				mu.Lock()
				defer mu.Unlock()
				current[su.Service] = cache[su.Service]
				cached++
				return
			}

			err := GetServiceRef(su, dir)

			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				errs = append(errs, err)
				// A failed refresh leaves the older file in place, so keep its entry.
				if old, ok := cache[su.Service]; ok {
					current[su.Service] = old
				}
				return
			}
			current[su.Service] = su
			fetched++
		})
	}
	wg.Wait()

	fmt.Printf("#### Service urls: %d, Fetched: %d, Cached (skipped): %d, Failed: %d\n", urlCount, fetched, cached, len(errs))

	// current only holds services from the remote list, so removed services
	// are already gone from the index; delete their files too.
	pruned, err := PruneStale(dir, urls)
	if err != nil {
		errs = append(errs, err)
	}
	for _, name := range pruned {
		fmt.Printf("pruned %s\n", name)
	}
	fmt.Printf("#### Pruned: %d\n", len(pruned))

	// Only services with a file on disk are indexed, so a service's presence
	// in the index means its data exists.
	if err := WriteServiceIndex(slices.Collect(maps.Values(current)), indexPath); err != nil {
		return err
	}
	fmt.Printf("#### Saved service index (%d services) to %s\n", len(current), indexPath)

	return errors.Join(errs...)
}

// GetServiceRefURLs returns the list of service reference links at indexURL.
// It fails, rather than return a partial or empty list, if the response is not
// a valid, non-empty list, so a bad index never prunes the cache.
func GetServiceRefURLs(indexURL string) ([]ServiceURL, error) {
	b, err := fetch(indexURL)
	if err != nil {
		return nil, err
	}

	var urls []ServiceURL
	if err := json.Unmarshal(b, &urls); err != nil {
		return nil, fmt.Errorf("parsing service index: %w", err)
	}
	if len(urls) == 0 {
		return nil, errors.New("service index lists no services")
	}
	for _, su := range urls {
		if err := validateServiceURL(su); err != nil {
			return nil, fmt.Errorf("service index: %w", err)
		}
	}

	return urls, nil
}

// validateServiceURL checks that su's service name is safe to use as a file
// name and that its URL uses HTTPS.
func validateServiceURL(su ServiceURL) error {
	if !serviceNameRegexp.MatchString(su.Service) {
		return fmt.Errorf("invalid service name %q", su.Service)
	}
	if u, err := url.Parse(su.URL); err != nil || u.Scheme != "https" {
		return fmt.Errorf("%s: service reference URL %q is not HTTPS", su.Service, su.URL)
	}
	return nil
}

// fetch returns the body of a GET of rawURL, which must respond 200 OK.
func fetch(rawURL string) ([]byte, error) {
	resp, err := httpClient.Get(rawURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET %s: unexpected status %s", rawURL, resp.Status)
	}

	return io.ReadAll(resp.Body)
}

// WriteServiceIndex writes urls to path as a JSON object keyed by service name
func WriteServiceIndex(urls []ServiceURL, path string) error {
	index := make(map[string]ServiceURL, len(urls))
	for _, su := range urls {
		index[su.Service] = su
	}

	b, err := json.MarshalIndent(index, "", "  ")
	if err != nil {
		return err
	}

	return writeFileAtomic(path, b)
}

// LoadServiceIndex reads a file written by WriteServiceIndex. The returned map
// supports constant-time lookup of a service by name:
//
//	index, err := LoadServiceIndex("service-urls.json")
//	su, ok := index["ec2"]
func LoadServiceIndex(path string) (map[string]ServiceURL, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var index map[string]ServiceURL
	if err := json.Unmarshal(b, &index); err != nil {
		return nil, err
	}

	return index, nil
}

// GetServiceRef downloads the service reference for su and writes it to
// <dir>/<service>.json. The existing file is replaced only if the download is
// a valid service reference.
func GetServiceRef(su ServiceURL, dir string) error {
	if err := validateServiceURL(su); err != nil {
		return err
	}

	fmt.Printf("fetching %s...\n", su.Service)
	b, err := fetch(su.URL)
	if err != nil {
		return fmt.Errorf("%s: %w", su.Service, err)
	}
	if _, err := parseServiceRef(b); err != nil {
		return fmt.Errorf("%s: invalid service reference: %w", su.Service, err)
	}

	path := filepath.Join(dir, su.Service+".json")
	if err := writeFileAtomic(path, b); err != nil {
		return fmt.Errorf("%s: %w", su.Service, err)
	}
	fmt.Printf("saved %s\n", path)

	return nil
}

// LoadCache returns the index of previously fetched services. It returns an
// empty index, forcing a full fetch, if the index file is missing or
// unreadable, or if dir does not exist or is empty.
func LoadCache(dir, indexPath string) map[string]ServiceURL {
	entries, err := os.ReadDir(dir)
	if err != nil || len(entries) == 0 {
		return map[string]ServiceURL{}
	}

	index, err := LoadServiceIndex(indexPath)
	if err != nil {
		return map[string]ServiceURL{}
	}

	return index
}

// needsFetch reports whether the data for su must be (re)fetched: it is not in
// the cache, its file is missing, or the remote copy is newer than the cached one.
func needsFetch(su ServiceURL, cache map[string]ServiceURL, dir string) bool {
	cached, ok := cache[su.Service]
	if !ok {
		return true
	}

	if _, err := os.Stat(filepath.Join(dir, su.Service+".json")); err != nil {
		return true
	}

	return su.Modified > cached.Modified
}

// PruneStale deletes <service>.json files in dir for services that are no
// longer in urls and returns the names of the services pruned.
func PruneStale(dir string, urls []ServiceURL) ([]string, error) {
	remote := make(map[string]struct{}, len(urls))
	for _, su := range urls {
		remote[su.Service] = struct{}{}
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var pruned []string
	var errs []error
	for _, e := range entries {
		name, ok := strings.CutSuffix(e.Name(), ".json")
		if e.IsDir() || !ok {
			continue
		}
		if _, ok := remote[name]; ok {
			continue
		}
		if err := os.Remove(filepath.Join(dir, e.Name())); err != nil {
			errs = append(errs, err)
			continue
		}
		pruned = append(pruned, name)
	}

	return pruned, errors.Join(errs...)
}

// writeFileAtomic writes data to path through a temporary file in the same
// directory, renamed into place only after a successful write and close. A
// failed or interrupted write leaves any existing file at path unchanged.
func writeFileAtomic(path string, data []byte) error {
	tmp, err := os.CreateTemp(filepath.Dir(path), filepath.Base(path)+".*.tmp")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name()) // no-op after a successful rename

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Chmod(tmp.Name(), 0o644); err != nil {
		return err
	}

	return os.Rename(tmp.Name(), path)
}
