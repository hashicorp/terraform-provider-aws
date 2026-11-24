// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package objectvalidator

import (
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/validators/internal"
)

// RequiredTogetherChildren checks that a set of path.Expression
// either has all known or all null values.
func RequiredTogetherChildren(expressions ...path.Expression) validator.Object {
	return internal.RequiredTogetherValidator(expressions...)
}
