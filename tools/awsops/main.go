package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"

	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/hashicorp/terraform-provider-aws/tools/awsops/awsops"
)

type output struct {
	Resources []resourceBlock `hcl:"resource,block"`
}

type resourceBlock struct {
	Type   string   `hcl:"type,label"`
	Create []string `hcl:"create,optional"`
	Read   []string `hcl:"read,optional"`
	Update []string `hcl:"update,optional"`
	Delete []string `hcl:"delete,optional"`
}

func main() {
	log.SetFlags(0)

	providerDir := flag.String("provider-dir", ".", "Root directory of the Terraform AWS provider")
	outputPath := flag.String("output", "", "Output file path (default: stdout)")
	flag.Parse()

	serviceDir := filepath.Join(*providerDir, "internal", "service")

	results, err := awsops.Analyze(serviceDir)
	if err != nil {
		log.Fatalf("analysis failed: %s", err)
	}

	hcl := formatHCL(results)

	if *outputPath != "" {
		if err := os.WriteFile(*outputPath, hcl, 0644); err != nil {
			log.Fatalf("writing output: %s", err)
		}
	} else {
		fmt.Print(string(hcl))
	}
}

func formatHCL(results map[string]awsops.ResourceOps) []byte {
	names := make([]string, 0, len(results))
	for name := range results {
		names = append(names, name)
	}
	sort.Strings(names)

	out := output{
		Resources: make([]resourceBlock, 0, len(names)),
	}
	for _, name := range names {
		ops := results[name]
		sort.Strings(ops.Create)
		sort.Strings(ops.Read)
		sort.Strings(ops.Update)
		sort.Strings(ops.Delete)

		out.Resources = append(out.Resources, resourceBlock{
			Type:   name,
			Create: ops.Create,
			Read:   ops.Read,
			Update: ops.Update,
			Delete: ops.Delete,
		})
	}

	f := hclwrite.NewEmptyFile()
	gohcl.EncodeIntoBody(&out, f.Body())
	return f.Bytes()
}
