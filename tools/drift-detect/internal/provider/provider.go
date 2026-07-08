// Copyright IBM Corp. 2026, 2026
// SPDX-License-Identifier: MPL-2.0

// Package provider automates the generation of a `terraform providers schema
// -json` output file from a local provider source directory.
//
// The typical call sequence is:
//
//  1. Call GenerateSchema(providerDir, providerSource) to build the provider
//     binary, run `terraform init`, run `terraform providers schema -json`, and
//     receive the path to the generated schema.json.
//  2. Use the returned path as input to tfschema.LoadFile.
//  3. Call CleanupSchema(schemaPath) (usually via defer) to remove the temp
//     directory.
//
// If a persistent cached copy is preferred, use GenerateSchemaTo instead.
//
// The package requires both `go` and `terraform` to be present in PATH.
package provider

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// GenerateSchema builds the provider binary from providerDir, runs
// `terraform init` and `terraform providers schema -json`, and returns the
// path to the generated schema.json inside a temporary directory.
//
// The caller is responsible for calling CleanupSchema when done.
func GenerateSchema(providerDir, providerSource string) (string, error) {
	if err := requireTerraform(); err != nil {
		return "", err
	}

	workDir, err := os.MkdirTemp("", "drift-detect-*")
	if err != nil {
		return "", fmt.Errorf("creating temp dir: %w", err)
	}

	schemaPath, err := generateSchemaInDir(workDir, providerDir, providerSource)
	if err != nil {
		os.RemoveAll(workDir)
		return "", err
	}
	return schemaPath, nil
}

// GenerateSchemaTo builds the provider, generates the schema, and writes it
// to destPath.  The temporary working directory is removed automatically.
func GenerateSchemaTo(providerDir, providerSource, destPath string) error {
	schemaPath, err := GenerateSchema(providerDir, providerSource)
	if err != nil {
		return err
	}
	defer CleanupSchema(schemaPath)

	src, err := os.ReadFile(schemaPath)
	if err != nil {
		return fmt.Errorf("reading generated schema: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
		return fmt.Errorf("creating destination directory: %w", err)
	}
	if err := os.WriteFile(destPath, src, 0o644); err != nil {
		return fmt.Errorf("writing schema to %s: %w", destPath, err)
	}
	return nil
}

// CleanupSchema removes the temporary directory that was created by
// GenerateSchema.  Safe to call with an empty string (no-op).
func CleanupSchema(schemaPath string) {
	if schemaPath != "" {
		os.RemoveAll(filepath.Dir(schemaPath))
	}
}

// ---------------------------------------------------------------------------
// Internal helpers
// ---------------------------------------------------------------------------

// generateSchemaInDir performs the full build + init + schema cycle inside
// workDir.  On any failure it returns an error; the caller owns cleanup.
func generateSchemaInDir(workDir, providerDir, providerSource string) (string, error) {
	// 1. Validate provider source format.
	parts := strings.Split(providerSource, "/")
	if len(parts) < 3 {
		return "", fmt.Errorf(
			"provider-source must be in format registry.terraform.io/namespace/name, got %q",
			providerSource,
		)
	}
	providerName := parts[len(parts)-1]

	// 2. Build the provider binary.
	binaryPath := filepath.Join(workDir, "terraform-provider")
	fmt.Fprintf(os.Stderr, "drift-detect: building provider (%s)...\n", providerDir)
	buildCmd := exec.Command("go", "build", "-o", binaryPath, ".")
	buildCmd.Dir = providerDir
	buildCmd.Stderr = os.Stderr
	if err := buildCmd.Run(); err != nil {
		return "", fmt.Errorf("building provider: %w", err)
	}

	// 3. Place the binary inside the Terraform plugin-dir layout.
	platform := runtime.GOOS + "_" + runtime.GOARCH
	pluginDir := filepath.Join(workDir, "plugin-dir", providerSource, "99.99.99", platform)
	if err := os.MkdirAll(pluginDir, 0o755); err != nil {
		return "", fmt.Errorf("creating plugin dir: %w", err)
	}
	destBinary := filepath.Join(pluginDir, fmt.Sprintf("terraform-provider-%s_v99.99.99", providerName))
	if err := os.Rename(binaryPath, destBinary); err != nil {
		return "", fmt.Errorf("moving provider binary: %w", err)
	}

	// 4. Write a minimal .tf file sufficient for `terraform init`.
	//    For the AWS provider the canonical stub is the partition data source.
	//    For other providers we fall back to a terraform{} block with
	//    required_providers so init can resolve the local plugin.
	var tfContent string
	if providerName == "aws" {
		tfContent = `data "aws_partition" "example" {}`
	} else {
		tfContent = fmt.Sprintf(`terraform {
  required_providers {
    %s = {
      source = "%s"
    }
  }
}
`, providerName, providerSource)
	}
	tfFile := filepath.Join(workDir, "main.tf")
	if err := os.WriteFile(tfFile, []byte(tfContent), 0o644); err != nil {
		return "", fmt.Errorf("writing tf stub: %w", err)
	}

	// 5. terraform init
	fmt.Fprintf(os.Stderr, "drift-detect: running terraform init...\n")
	initCmd := exec.Command("terraform", "init", "-plugin-dir", filepath.Join(workDir, "plugin-dir"))
	initCmd.Dir = workDir
	initCmd.Stderr = os.Stderr
	if err := initCmd.Run(); err != nil {
		return "", fmt.Errorf("terraform init: %w", err)
	}

	// 6. terraform providers schema -json
	fmt.Fprintf(os.Stderr, "drift-detect: running terraform providers schema -json...\n")
	schemaPath := filepath.Join(workDir, "schema.json")
	schemaFile, err := os.Create(schemaPath)
	if err != nil {
		return "", fmt.Errorf("creating schema output file: %w", err)
	}
	schemaCmd := exec.Command("terraform", "providers", "schema", "-json")
	schemaCmd.Dir = workDir
	schemaCmd.Stdout = schemaFile
	schemaCmd.Stderr = os.Stderr
	runErr := schemaCmd.Run()
	schemaFile.Close()
	if runErr != nil {
		return "", fmt.Errorf("terraform providers schema: %w", runErr)
	}

	fmt.Fprintf(os.Stderr, "drift-detect: schema written to %s\n", schemaPath)
	return schemaPath, nil
}

// requireTerraform returns an error when the `terraform` binary cannot be
// found in PATH, giving the user an actionable message.
func requireTerraform() error {
	if _, err := exec.LookPath("terraform"); err != nil {
		return fmt.Errorf("terraform not found in PATH: install Terraform to use --provider-dir")
	}
	return nil
}
