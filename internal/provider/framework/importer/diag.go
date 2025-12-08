// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package importer

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

const (
	InvalidResourceImportIDValue = "Invalid Resource Import ID Value"

	InvalidIdentityAttributeValue = "Invalid Identity Attribute Value"
)

func InvalidResourceImportIDError(description string) diag.Diagnostic {
	return diag.NewErrorDiagnostic(
		InvalidResourceImportIDValue,
		"The import ID "+description,
	)
}

func InvalidIdentityAttributeError(path path.Path, description string) diag.Diagnostic {
	return diag.NewAttributeErrorDiagnostic(
		path,
		InvalidIdentityAttributeValue,
		fmt.Sprintf("Identity attribute %q ", path)+description,
	)
}
