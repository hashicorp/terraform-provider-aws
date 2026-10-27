// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"maps"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"sync"
)

const serviceRefURL = "http://servicereference.us-east-1.amazonaws.com/"

type ServiceURL struct {
	Service  string `json:"service"`
	URL      string `json:"url"`
	Modified int    `json:"modified"`
}

// GetServiceRefURLs returns the current list of service reference links
func GetServiceRefURLs() ([]ServiceURL, error) {
	resp, err := http.Get(serviceRefURL)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var urls []ServiceURL
	err = json.Unmarshal(b, &urls)

	return urls, err
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

	return os.WriteFile(path, b, 0o644)
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
// <dir>/<service>.json
func GetServiceRef(su ServiceURL, dir string) error {
	fmt.Printf("fetching %s...\n", su.Service)
	resp, err := http.Get(su.URL)
	if err != nil {
		return fmt.Errorf("%s: %w", su.Service, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s: unexpected status %s", su.Service, resp.Status)
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("%s: %w", su.Service, err)
	}

	path := filepath.Join(dir, su.Service+".json")
	if err := os.WriteFile(path, b, 0o644); err != nil {
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

func main() {
	outDir := flag.String("out", "service-data", "directory to write service reference files to")
	indexFile := flag.String("index", "service-urls-cache.json", "file to write the service name to URL index to")
	flag.Parse()

	cache := LoadCache(*outDir, *indexFile)
	fmt.Printf("#### Loaded %d cached services\n", len(cache))

	if err := os.MkdirAll(*outDir, 0o755); err != nil {
		log.Fatal(err)
	}

	fmt.Println("#### Fetching service reference urls")
	urls, err := GetServiceRefURLs()
	fmt.Println("#### Completed fetching service reference urls")
	if err != nil {
		log.Fatal(err)
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
			if !needsFetch(su, cache, *outDir) {
				mu.Lock()
				defer mu.Unlock()
				current[su.Service] = cache[su.Service]
				cached++
				return
			}

			err := GetServiceRef(su, *outDir)

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

	// Only services with a file on disk are indexed, so a service's presence
	// in the index means its data exists.
	if err := WriteServiceIndex(slices.Collect(maps.Values(current)), *indexFile); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("#### Saved service index (%d services) to %s\n", len(current), *indexFile)

	if err := errors.Join(errs...); err != nil {
		log.Fatal(err)
	}
}
