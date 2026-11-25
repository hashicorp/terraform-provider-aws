// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package objectvalidator

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/validators/internal"
)

// AtMostOneOfChildren checks that of a set of path.Expression,
// at most one has a non-null value.
func AtMostOneOfChildren(expressions ...path.Expression) validator.Object {
	return internal.AtMostOneOfValidator(validatordiag.InvalidAttributeCombinationDiagnostic, expressions...)
}
