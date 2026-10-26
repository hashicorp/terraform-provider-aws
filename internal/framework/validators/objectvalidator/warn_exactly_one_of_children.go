// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package objectvalidator

import (
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/validators/internal"
)

// WarnExactlyOneOfChildren acts similarly to `objectvalidator.ExactlyOneOf` except that it returns a Warning and
// that it doesn't include the Object in the count of matched attributes.
func WarnExactlyOneOfChildren(expressions ...path.Expression) validator.Object {
	return internal.ExactlyOneOfValidator(fwdiag.WarningInvalidAttributeCombinationDiagnostic, expressions...)
}
