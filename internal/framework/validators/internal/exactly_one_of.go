// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package internal

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

type always struct{}

func (always) Eval(context.Context, attr.Value) bool {
	return true
}

func (always) String() string {
	return ""
}

func ExactlyOneOfValidator(diagFactory invalidAttributeCombinationDiagnosticFunc, expressions ...path.Expression) exactlyOneOfWhenValidator {
	return exactlyOneOfWhenValidator{
		conditionalMatchedPathCountValidator{
			when:            always{},
			pathExpressions: expressions,
		},
		diagFactory,
	}
}
