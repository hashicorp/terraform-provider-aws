// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"flag"
	"log"
)

func main() {
	serviceDataDir := flag.String("out", "service-data", "directory to write service reference files to")
	serviceIndexFile := flag.String("index", "service-urls-cache.json", "file to write the service name to URL index to")
	sdkMapFile := flag.String("sdkmap", "../../sdkmap.json", "sdkmap output to read")
	permissionsMapFile := flag.String("permissions-map", "../../permissions_map.json", "file to write the sdkmap output with IAM permissions to")
	skipSync := flag.Bool("skip-sync", false, "use the existing service reference files without checking for updates")
	flag.Parse()

	if !*skipSync {
		if err := SyncServiceRefs(*serviceDataDir, *serviceIndexFile); err != nil {
			log.Fatal(err)
		}
	}

	if err := GeneratePermissionsMap(*serviceDataDir, *sdkMapFile, *permissionsMapFile); err != nil {
		log.Fatal(err)
	}
}
