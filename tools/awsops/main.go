package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/hashicorp/terraform-provider-aws/tools/awsops/awsops"
)

func main() {
	log.SetFlags(0)

	providerDir := flag.String("provider-dir", ".", "Root directory of the Terraform AWS provider")
	output := flag.String("output", "", "Output file path (default: stdout)")
	flag.Parse()

	serviceDir := filepath.Join(*providerDir, "internal", "service")

	results, err := awsops.Analyze(serviceDir)
	if err != nil {
		log.Fatalf("analysis failed: %s", err)
	}

	hcl := formatHCL(results)

	if *output != "" {
		if err := os.WriteFile(*output, []byte(hcl), 0644); err != nil {
			log.Fatalf("writing output: %s", err)
		}
	} else {
		fmt.Print(hcl)
	}
}

func formatHCL(results map[string]awsops.ResourceOps) string {
	var b strings.Builder

	names := make([]string, 0, len(results))
	for name := range results {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		ops := results[name]
		b.WriteString(fmt.Sprintf("resource %q {\n", name))
		for _, method := range []string{"create", "read", "update", "delete"} {
			var methodOps []string
			switch method {
			case "create":
				methodOps = ops.Create
			case "read":
				methodOps = ops.Read
			case "update":
				methodOps = ops.Update
			case "delete":
				methodOps = ops.Delete
			}
			if len(methodOps) > 0 {
				sort.Strings(methodOps)
				b.WriteString(fmt.Sprintf("  %s = [\n", method))
				for _, op := range methodOps {
					b.WriteString(fmt.Sprintf("    %q,\n", op))
				}
				b.WriteString("  ]\n")
			}
		}
		b.WriteString("}\n\n")
	}

	return b.String()
}
