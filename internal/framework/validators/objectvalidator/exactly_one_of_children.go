// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package objectvalidator

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/validators/internal"
)

// ExactlyOneOfChildren acts similarly to `objectvalidator.ExactlyOneOf` except that
// it  doesn't include the Object in the count of matched attributes.
// https://github.com/hashicorp/terraform-plugin-framework-validators/issues/274
func ExactlyOneOfChildren(expressions ...path.Expression) validator.Object {
	return internal.ExactlyOneOfValidator(validatordiag.InvalidAttributeCombinationDiagnostic, expressions...)
}
