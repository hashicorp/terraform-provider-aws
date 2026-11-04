// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package objectvalidator

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/validators/internal"
)

// AtLeastOneOfChildren checks that of a set of path.Expression,
// at least one has a non-null value.
func AtLeastOneOfChildren(expressions ...path.Expression) validator.Object {
	return internal.AtLeastOneOfValidator(validatordiag.InvalidAttributeCombinationDiagnostic, expressions...)
}
