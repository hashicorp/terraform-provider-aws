// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sdkv2

import (
	"fmt"
	"strings"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
)

func DeprecatedEnvVarDiag(envvar, replacement string) diag.Diagnostic {
	return errs.NewWarningDiagnostic(
		"Deprecated Environment Variable",
		fmt.Sprintf(`The environment variable "%s" is deprecated. Use environment variable "%s" instead.`, envvar, replacement),
	)
}

func ConflictingEndpointsWarningDiag(elementPath cty.Path, attrs ...string) diag.Diagnostic {
	attrPaths := make([]string, len(attrs))
	for i, attr := range attrs {
		path := elementPath.GetAttr(attr)
		attrPaths[i] = `"` + errs.PathString(path) + `"`
	}
	return errs.NewAttributeWarningDiagnostic(
		elementPath,
		"Invalid Attribute Combination",
		fmt.Sprintf("Only one of the following attributes should be set: %s"+
			"\n\nThis will be an error in a future release.",
			strings.Join(attrPaths, ", ")),
	)
}
