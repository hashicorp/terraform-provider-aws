// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// drift-detect detects schema drift between AWS API models and the Terraform
// AWS Provider schemas.
//
// # Current status: Phase 1 – TF + AWS Schema Extraction
//
// The tool can currently:
//   - Load a pre-generated `terraform providers schema -json` output file, or
//     build the provider locally and generate it on the fly.
//   - Parse the TF JSON into a Phase 1 IR (top-level primitive fields only).
//   - Resolve Terraform resources to AWS Smithy model files using the provider
//     source and the resource mapping file.
//   - Extract an AWS IR
//   - Print TF and AWS IR side-by-side for a given resource with a preliminary
//     "present in AWS but missing in TF" field list.
//
// # Usage
//
//  Need to enter the terraform-provider-aws/tools/drift-detect directory
// 
//  # --provider-dir is required if the schema-json file does not already exist.
//	# --resource is required and must be in the format aws_<service>_<resource>:
//
//	./drift-detect 	--provider-dir ../..
// 					--resource aws_sqs_queue \
//	             	--schema-json .cache/schema.json
//
//	./drift-detect 	--resource aws_sqs_queue \
//	          		--schema-json .cache/schema.json \
//	            	--mappings tools/drift-detect/mappings/aws_resources.yaml
//
//	# JSON output:
//	./drift-detect 	--resource aws_sqs_queue ... --json
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/hashicorp/terraform-provider-aws/tools/drift-detect/internal/awsmapping"
	"github.com/hashicorp/terraform-provider-aws/tools/drift-detect/internal/awsschema"
	"github.com/hashicorp/terraform-provider-aws/tools/drift-detect/internal/provider"
	"github.com/hashicorp/terraform-provider-aws/tools/drift-detect/internal/serviceindex"
	"github.com/hashicorp/terraform-provider-aws/tools/drift-detect/internal/smithymodel"
	"github.com/hashicorp/terraform-provider-aws/tools/drift-detect/internal/tfschema"
)

const (
	defaultAPIModelsBaseURL = "https://raw.githubusercontent.com/aws/api-models-aws/main"
	defaultProviderSource   = "registry.terraform.io/hashicorp/aws"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "drift-detect: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	var (
		schemaJSON     = flag.String("schema-json", ".cache/schema.json", "path to terraform providers schema -json output file")
		providerDir    = flag.String("provider-dir", "", "path to provider source directory (builds provider and generates schema)")
		mappingsFile   = flag.String("mappings", "mappings/aws_resources.yaml", "path to aws_resources.yaml mapping file")
		resource       = flag.String("resource", "", "required: AWS resource name in the format aws_<service>_<resource> (e.g. aws_sqs_queue)")
		outputJSON     = flag.Bool("json", false, "output results as JSON")
		refreshSchema  = flag.Bool("refresh-schema", false, "regenerate cached schema even if schema-json file already exists")
	)
	flag.Parse()

	if err := validateResource(*resource); err != nil {
		return err
	}

	// --- TF schema ---
	schemaPath, cleanup, err := resolveSchema(*schemaJSON, *providerDir, *refreshSchema)
	if err != nil {
		return err
	}
	if cleanup != nil {
		defer cleanup()
	}

	ps, err := tfschema.LoadFile(schemaPath, defaultProviderSource)
	if err != nil {
		return fmt.Errorf("loading TF schema: %w", err)
	}

	// --- AWS mappings ---
	var mappings *awsmapping.File
	if *mappingsFile != "" {
		mappings, err = awsmapping.LoadFile(*mappingsFile)
		if err != nil {
			return fmt.Errorf("loading mappings: %w", err)
		}
	}

	// --- Resource-Service Index ---
	resourceServiceIndex, err := serviceindex.LoadResourceServiceIndex(*providerDir, *refreshSchema, ps, mappings)
	if err != nil {
		return fmt.Errorf("loading resource service index: %w", err)
	}

	resourceNames := []string{*resource} // FIXME: support multiple resources

	// --- Build output ---
	reports := buildReports(ps, mappings, resourceServiceIndex, defaultAPIModelsBaseURL, resourceNames)
	if *outputJSON {
		return outputJSONReport(reports)
	}
	return outputTextReport(reports)
}

// ---------------------------------------------------------------------------
// Output types
// ---------------------------------------------------------------------------

// resourceReport is the per-resource section of the output.
type resourceReport struct {
	Resource string      `json:"resource"`
	TF       *sideReport `json:"terraform"`
	AWS      *sideReport `json:"aws,omitempty"`
	// MissingInTF lists field names present in the AWS IR but absent in the TF IR.
	// Phase 3 will expand this into a full structured drift result.
	MissingInTF []string `json:"missing_in_tf,omitempty"`
}

// sideReport is the per-source field list.
type sideReport struct {
	Source     string            `json:"source"`
	FieldCount int               `json:"field_count"`
	Fields     []*tfschema.Field `json:"fields"`
}

// ---------------------------------------------------------------------------
// Text output
// ---------------------------------------------------------------------------

func outputTextReport(reports []resourceReport) error {
	for _, r := range reports {
		fmt.Printf("\n══ %s ══\n", r.Resource)
		printSideText("terraform", r.TF)
		if r.AWS != nil {
			printSideText("aws", r.AWS)
			if len(r.MissingInTF) > 0 {
				fmt.Printf("\n  [missing in TF: %d field(s)]\n", len(r.MissingInTF))
				for _, f := range r.MissingInTF {
					fmt.Printf("    - %s\n", f)
				}
			} else {
				fmt.Printf("\n  [no fields missing in TF for compared set]\n")
			}
		}
	}
	fmt.Fprintf(os.Stderr, "\n%d resource(s) compared.\n", len(reports))
	return nil
}

func printSideText(label string, s *sideReport) {
	if s == nil {
		return
	}
	fmt.Printf("  ── %s (%d fields) ──\n", label, s.FieldCount)
	for _, f := range s.Fields {
		flags := fieldFlags(f)
		fmt.Printf("    %-42s %-8s %s\n", f.Name, f.Type, flags)
	}
}

// ---------------------------------------------------------------------------
// JSON output
// ---------------------------------------------------------------------------

func outputJSONReport(reports []resourceReport) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(reports)
}

// ---------------------------------------------------------------------------
// Report builder
// ---------------------------------------------------------------------------

func buildReports(ps *tfschema.ProviderSchema, mappings *awsmapping.File, rsIdx map[string]*serviceindex.ResourceServiceInfo, apiModelsBaseURL string, resourceNames []string) []resourceReport {
	var reports []resourceReport

	for _, name := range resourceNames {
		tfIR, ok := ps.Resources[name]
		if !ok {
			fmt.Fprintf(os.Stderr, "warning: TF resource %q not found\n", name)
			continue
		}
		tfFields := sortedFields(tfIR.Fields)

		r := resourceReport{
			Resource: name,
			TF: &sideReport{
				Source:     "terraform",
				FieldCount: len(tfFields),
				Fields:     tfFields,
			},
		}

		m, awsModel, namespace, err := resolveResourceInfo(name, mappings, rsIdx, apiModelsBaseURL)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: %s: %v\n", name, err)
			reports = append(reports, r)
			continue
		}

		awsIR, err := awsschema.ExtractResource(name, m, awsModel, namespace)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: extracting %s from AWS model: %v\n", name, err)
		} else {
			awsFields := sortedFields(awsIR.Fields)
			r.AWS = &sideReport{
				Source:     "aws",
				FieldCount: len(awsFields),
				Fields:     awsFields,
			}
			r.MissingInTF = missingInTF(tfIR, awsIR)
		}

		reports = append(reports, r)
	}
	return reports
}

func resolveResourceInfo(
	name string,
	mappings *awsmapping.File,
	rsIdx map[string]*serviceindex.ResourceServiceInfo,
	apiModelsBaseURL string,
) (
	*awsmapping.ResourceMapping,
	*smithymodel.Model,
	string,
	error,
) {
	resourceInfo, ok := rsIdx[name]
	if !ok {
		return nil, nil, "", fmt.Errorf("resource %q not found in ResourceServiceIndex", name)
	}

	if resourceInfo.Skipped != nil {
		return nil, nil, "", fmt.Errorf("resource %q is skipped: %s", name, resourceInfo.Skipped.Reason)
	}

	if resourceInfo.Error != "" {
		return nil, nil, "", fmt.Errorf("resource service index: %s", resourceInfo.Error)
	}

	awsModel, err := awsschema.LoadModel(apiModelsBaseURL, resourceInfo.AWSFile)
	if err != nil {
		return nil, nil, "", fmt.Errorf("loading model: %w", err)
	}

	var m *awsmapping.ResourceMapping

	if mappings != nil {
		if existing, ok := mappings.Resources[name]; ok {
			m = existing
		}
	}

	if m == nil {
		m = &awsmapping.ResourceMapping{}
	}

	if mappings != nil && mappings.SuppressFields != nil {
		m.SuppressFields = append(m.SuppressFields,mappings.SuppressFields...)
	}

	namespace := resourceInfo.AWSNamespace

	if m.SmithyResource == "" && awsModel != nil && awsModel.HasResourceShapes() {
		candidate := resourceInfo.TFResourceSegment
		resourceID := namespace + "#" + candidate

		if s := awsModel.Shape(resourceID); s != nil && s.Kind == smithymodel.KindResource {
			m.SmithyResource = candidate
		} else {
			candidate2 := candidate + "Resource"
			resourceID2 := namespace + "#" + candidate2
			if s2 := awsModel.Shape(resourceID2); s2 != nil &&
				s2.Kind == smithymodel.KindResource {
				m.SmithyResource = candidate2
			}
		}
	}

	return m, awsModel, namespace, nil
}

// missingInTF returns a sorted list of AWS IR field names not present in the
// TF IR. This is the preliminary Phase 1 drift signal.
func missingInTF(tfIR *tfschema.ResourceIR, awsIR *tfschema.ResourceIR) []string {
	var missing []string
	for name := range awsIR.Fields {
		if _, ok := tfIR.Fields[name]; !ok {
			missing = append(missing, name)
		}
	}
	sort.Strings(missing)
	return missing
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func sortedFields(fields map[string]*tfschema.Field) []*tfschema.Field {
	result := make([]*tfschema.Field, 0, len(fields))
	for _, f := range fields {
		result = append(result, f)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	return result
}

func fieldFlags(f *tfschema.Field) string {
	var parts []string
	if f.Required {
		parts = append(parts, "required")
	}
	if f.Optional {
		parts = append(parts, "optional")
	}
	if f.Computed {
		parts = append(parts, "computed")
	}
	result := ""
	for i, p := range parts {
		if i > 0 {
			result += ","
		}
		result += p
	}
	return result
}

func resolveSchema(schemaJSON, providerDir string, refresh bool) (string, func(), error) {
	switch {
	case schemaJSON != "" && providerDir != "":
		if !filepath.IsAbs(schemaJSON) {
			schemaJSON = filepath.Join(providerDir, schemaJSON)
		}
		if refresh || !fileExists(schemaJSON) {
			fmt.Fprintf(os.Stderr, "Building schema (this may take a few minutes)...\n")
			fmt.Fprintf(os.Stderr, "Subsequent runs will reuse the cached schema at %s\n", schemaJSON)
			fmt.Fprintf(os.Stderr, "Use --refresh-schema to regenerate after provider changes.\n")
			if err := provider.GenerateSchemaTo(providerDir, defaultProviderSource, schemaJSON); err != nil {
				return "", nil, fmt.Errorf("generating schema: %w", err)
			}
		}
		return schemaJSON, nil, nil

	case providerDir != "" && schemaJSON == "":
		path, err := provider.GenerateSchema(providerDir, defaultProviderSource)
		if err != nil {
			return "", nil, fmt.Errorf("generating schema: %w", err)
		}
		return path, func() { provider.CleanupSchema(path) }, nil

	case schemaJSON != "":
		return schemaJSON, nil, nil

	default:
		return "", nil, fmt.Errorf(
			"one of --schema-json or --provider-dir is required\n\n" +
				"Examples:\n" +
				"  drift-detect --resource aws_sqs_queue --schema-json schema.json\n" +
				"  drift-detect --resource aws_sqs_queue --schema-json schema.json --mappings mappings/aws_resources.yaml",
		)
	}
}
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// validateResource checks that r is non-empty and uses the aws_ resource-name
// prefix. Detailed resource lookup and service resolution happen later.
func validateResource(r string) error {
    if r == "" {
        return fmt.Errorf("--resource is required")
    }

    if !strings.HasPrefix(r, "aws_") {
        return fmt.Errorf("--resource %q must start with aws_", r)
    }

    return nil
}
