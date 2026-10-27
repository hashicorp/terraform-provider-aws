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
	"sync/atomic"
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
	flag.Parse()

	if err := os.MkdirAll(*outDir, 0o755); err != nil {
		log.Fatal(err)
	}

	fmt.Println("fetching service reference urls")
	urls, err := GetServiceRefURLs()
	fmt.Println("Completed fetching service reference urls")
	if err != nil {
		log.Fatal(err)
	}

	urlCount := len(urls)
	fmt.Printf("service urls found: %d\n", urlCount)

	var (
		wg        sync.WaitGroup
		mu        sync.Mutex
		errs      []error
		fileCount atomic.Int64
	)
	for _, su := range urls {
		wg.Go(func() {
			if err := GetServiceRef(su, *outDir); err != nil {
				mu.Lock()
				defer mu.Unlock()
				errs = append(errs, err)
				return
			}
			fileCount.Add(1)
		})
	}
	wg.Wait()

	fmt.Printf("service urls: %d, files written: %d\n", urlCount, fileCount.Load())

	if err := errors.Join(errs...); err != nil {
		log.Fatal(err)
	}
}