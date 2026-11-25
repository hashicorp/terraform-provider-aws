// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"flag"
	"log"
	"strings"

	"github.com/hashicorp/terraform-provider-aws/tools/injectr/docs"
	"github.com/hashicorp/terraform-provider-aws/tools/injectr/permissions"
	"github.com/hashicorp/terraform-provider-aws/tools/injectr/sdkmap"
	"github.com/hashicorp/terraform-provider-aws/tools/injectr/serviceref"
)

func main() {
	log.SetFlags(0)
	root := flag.String("root", "../..", "path to the root of the terraform-provider-aws repository")
	serviceDataDir := flag.String("out", "service-data", "directory to write service reference files to")
	serviceIndexFile := flag.String("index", "service-urls-cache.json", "file to write the service name to URL index to")
	permissionsMapFile := flag.String("permissions-map", "../../permissions_map.json", "file to write the permissions map to")
	skipSync := flag.Bool("skip-sync", false, "use the existing service reference files without checking for updates")

	var insert docs.InsertOptions
	flag.Var((*stringList)(&insert.Files), "insert-file", "comma-separated documentation files, relative to -root, to add IAM permissions to")
	flag.Var((*stringList)(&insert.Services), "insert-service", "comma-separated service packages to add IAM permissions to, e.g. s3,sqs")
	flag.BoolVar(&insert.All, "insert-all", false, "add IAM permissions to the documentation of every resource and data source")
	flag.BoolVar(&insert.DryRun, "dry-run", false, "report documentation changes without writing them")
	flag.Parse()
	insert.Root = *root

	m, err := sdkmap.Generate(*root)
	if err != nil {
		log.Fatal(err)
	}

	if !*skipSync {
		if err := serviceref.SyncServiceRefs(*serviceDataDir, *serviceIndexFile); err != nil {
			log.Fatal(err)
		}
	}

	if err := permissions.GeneratePermissionsMap(*serviceDataDir, m, *permissionsMapFile); err != nil {
		log.Fatal(err)
	}

	if insert.Enabled() {
		if err := docs.InsertIAMPermissions(m, insert); err != nil {
			log.Fatal(err)
		}
	}
}

// stringList is a flag holding a comma-separated list, e.g. "s3,sqs".
type stringList []string

func (l *stringList) String() string { return strings.Join(*l, ",") }

func (l *stringList) Set(v string) error {
	for item := range strings.SplitSeq(v, ",") {
		if item = strings.TrimSpace(item); item != "" {
			*l = append(*l, item)
		}
	}
	return nil
}
