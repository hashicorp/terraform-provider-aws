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
	"net/http"
	"os"
	"path/filepath"
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

func main() {
	outDir := flag.String("out", "service-data", "directory to write service reference files to")
	indexFile := flag.String("index", "service-urls.json", "file to write the service name to URL index to")
	flag.Parse()

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
		fetched []ServiceURL
	)
	for _, su := range urls {
		wg.Go(func() {
			err := GetServiceRef(su, *outDir)

			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				errs = append(errs, err)
				return
			}
			fetched = append(fetched, su)
		})
	}
	wg.Wait()

	fmt.Printf("#### Service urls: %d, Files written: %d\n", urlCount, len(fetched))

	// Only services whose data was fetched and written are indexed, so a
	// service's presence in the index means its file exists.
	if err := WriteServiceIndex(fetched, *indexFile); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("#### Saved service index (%d services) to %s\n", len(fetched), *indexFile)

	if err := errors.Join(errs...); err != nil {
		log.Fatal(err)
	}
}