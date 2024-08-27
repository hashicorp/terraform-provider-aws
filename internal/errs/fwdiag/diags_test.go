// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fwdiag_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
)

func TestDiagnosticsError(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		testName string
		diags    diag.Diagnostics
		wantErr  bool
	}{
		{
			testName: "nil Diagnostics",
		},
		{
			testName: "single warning Diagnostics",
			diags:    diag.Diagnostics{diag.NewWarningDiagnostic("summary", "detail")},
		},
		{
			testName: "single error Diagnostics",
			diags:    diag.Diagnostics{diag.NewErrorDiagnostic("summary", "detail")},
			wantErr:  true,
		},
		{
			testName: "mixed warning and error Diagnostics",
			diags: diag.Diagnostics{
				diag.NewWarningDiagnostic("summary1", "detail1"),
				diag.NewErrorDiagnostic("summary2", "detail2"),
			},
			wantErr: true,
		},
		{
			testName: "multiple error Diagnostics",
			diags: diag.Diagnostics{
				diag.NewErrorDiagnostic("summary1", "detail1"),
				diag.NewErrorDiagnostic("summary2", "detail2"),
			},
			wantErr: true,
		},
		{
			testName: "multiple warning Diagnostics",
			diags: diag.Diagnostics{
				diag.NewWarningDiagnostic("summary1", "detail1"),
				diag.NewWarningDiagnostic("summary2", "detail2"),
			},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.testName, func(t *testing.T) {
			t.Parallel()

			err := fwdiag.DiagnosticsError(testCase.diags)
			gotErr := err != nil

			if gotErr != testCase.wantErr {
				t.Errorf("gotErr = %v, wantErr = %v", gotErr, testCase.wantErr)
			}
		})
	}
}
