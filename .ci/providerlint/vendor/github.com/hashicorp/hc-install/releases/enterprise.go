// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package releases

import "fmt"

type EnterpriseOptions struct {
	// LicenseDir represents directory path where to install license files (required)
	LicenseDir string

	// Meta represents optional version metadata (e.g. hsm, fips1402)
	Meta string
}

func enterpriseVersionMetadata(eo *EnterpriseOptions) string {
	if eo == nil {
		return ""
	}

	metadata := "ent"
	if eo.Meta != "" {
		metadata += "." + eo.Meta
	}
	return metadata
}

func validateEnterpriseOptions(eo *EnterpriseOptions) error {
	if eo == nil {
		return nil
	}

	if eo.LicenseDir == "" {
		return fmt.Errorf("LicenseDir must be provided when requesting enterprise versions")
	}

	return nil
}
