// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"flag"
	"fmt"
	"log"
)

func main() {
	serviceDataDir := flag.String("out", "service-data", "directory to write service reference files to")
	serviceIndexFile := flag.String("index", "service-urls-cache.json", "file to write the service name to URL index to")
	flag.Parse()

	if err := SyncServiceRefs(*serviceDataDir, *serviceIndexFile); err != nil {
		log.Fatal(err)
	}

	refs, err := LoadServiceRefs(*serviceDataDir)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("#### Parsed %d service references\n", len(refs))
}
