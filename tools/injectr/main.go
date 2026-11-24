// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"flag"
	"log"
	"strings"

	"github.com/hashicorp/terraform-provider-aws/tools/injectr/docs"
	"github.com/hashicorp/terraform-provider-aws/tools/injectr/permissions"
	"github.com/hashicorp/terraform-provider-aws/tools/injectr/serviceref"
)

func main() {
	serviceDataDir := flag.String("out", "service-data", "directory to write service reference files to")
	serviceIndexFile := flag.String("index", "service-urls-cache.json", "file to write the service name to URL index to")
	sdkMapFile := flag.String("sdkmap", "../../sdkmap.json", "sdkmap output to read")
	permissionsMapFile := flag.String("permissions-map", "../../permissions_map.json", "file to write the sdkmap output with IAM permissions to")
	skipSync := flag.Bool("skip-sync", false, "use the existing service reference files without checking for updates")
	skipGenerate := flag.Bool("skip-generate", false, "use the existing permissions map without regenerating it")

	var insert docs.InsertOptions
	flag.StringVar(&insert.Root, "root", "../..", "path to the root of the terraform-provider-aws repository")
	flag.Var((*stringList)(&insert.Files), "insert-file", "comma-separated documentation files, relative to -root, to add IAM permissions to")
	flag.Var((*stringList)(&insert.Services), "insert-service", "comma-separated service packages to add IAM permissions to, e.g. s3,sqs")
	flag.BoolVar(&insert.All, "insert-all", false, "add IAM permissions to the documentation of every resource and data source")
	flag.BoolVar(&insert.DryRun, "dry-run", false, "report documentation changes without writing them")
	flag.Parse()

	if !*skipSync {
		if err := serviceref.SyncServiceRefs(*serviceDataDir, *serviceIndexFile); err != nil {
			log.Fatal(err)
		}
	}

	if !*skipGenerate {
		if err := permissions.GeneratePermissionsMap(*serviceDataDir, *sdkMapFile, *permissionsMapFile); err != nil {
			log.Fatal(err)
		}
	}

	if insert.Enabled() {
		if err := docs.InsertIAMPermissions(*permissionsMapFile, insert); err != nil {
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
