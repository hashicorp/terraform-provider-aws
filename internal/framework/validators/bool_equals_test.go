// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package validators_test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	fwvalidators "github.com/hashicorp/terraform-provider-aws/internal/framework/validators"
)

func TestBoolEqualsValidator(t *testing.T) {
	t.Parallel()

	type testCase struct {
		in        types.Bool
		validator validator.Bool
		expErrors int
	}

	testCases := map[string]testCase{
		"simple-match": {
			in:        types.BoolValue(true),
			validator: fwvalidators.BoolEquals(true),
			expErrors: 0,
		},
		"simple-mismatch": {
			in:        types.BoolValue(false),
			validator: fwvalidators.BoolEquals(true),
			expErrors: 1,
		},
		"skip-validation-on-null": {
			in:        types.BoolNull(),
			validator: fwvalidators.BoolEquals(true),
			expErrors: 0,
		},
		"skip-validation-on-unknown": {
			in:        types.BoolUnknown(),
			validator: fwvalidators.BoolEquals(true),
			expErrors: 0,
		},
	}

	for name, test := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			req := validator.BoolRequest{
				ConfigValue: test.in,
			}
			res := validator.BoolResponse{}
			test.validator.ValidateBool(context.TODO(), req, &res)

			if test.expErrors > 0 && !res.Diagnostics.HasError() {
				t.Fatalf("expected %d error(s), got none", test.expErrors)
			}

			if test.expErrors > 0 && test.expErrors != res.Diagnostics.ErrorsCount() {
				t.Fatalf("expected %d error(s), got %d: %v", test.expErrors, res.Diagnostics.ErrorsCount(), res.Diagnostics)
			}

			if test.expErrors == 0 && res.Diagnostics.HasError() {
				t.Fatalf("expected no error(s), got %d: %v", res.Diagnostics.ErrorsCount(), res.Diagnostics)
			}
		})
	}
}
