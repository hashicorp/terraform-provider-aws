// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package objectvalidator

import (
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/validators/internal"
)

// WarnAtMostOneOfChildren is equivalent to AtMostOneOfChildren, but returns a Warning instead of an error
func WarnAtMostOneOfChildren(expressions ...path.Expression) validator.Object {
	return internal.AtMostOneOfValidator(fwdiag.WarningInvalidAttributeCombinationDiagnostic, expressions...)
}
