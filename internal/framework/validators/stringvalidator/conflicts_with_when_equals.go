// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package stringvalidator

import (
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/validators/internal"
)

// ConflictsWithWhenEquals checks that each path.Expression has a null
// configuration value when the stringy attribute being validated has the known
// specified value.
//
// Relative path.Expressions are resolved using the attribute being
// validated.
func ConflictsWithWhenEquals[T ~string](value T, expressions ...path.Expression) validator.String {
	return internal.ConflictsWithWhenValidator(whenEquals[T]{value: value}, expressions...)
}
