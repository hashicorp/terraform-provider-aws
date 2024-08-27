// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sdkdiag_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

func TestErrors(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		testName string
		diags    diag.Diagnostics
		want     int
	}{
		{
			testName: "nil Diagnostics",
		},
		{
			testName: "single warning Diagnostics",
			diags:    diag.Diagnostics{diag.Diagnostic{Severity: diag.Warning, Summary: "summary", Detail: "detail"}},
		},
		{
			testName: "single error Diagnostics",
			diags:    diag.Diagnostics{diag.Diagnostic{Severity: diag.Error, Summary: "summary", Detail: "detail"}},
			want:     1,
		},
		{
			testName: "mixed warning and error Diagnostics",
			diags: diag.Diagnostics{
				diag.Diagnostic{Severity: diag.Warning, Summary: "summary1", Detail: "detail1"},
				diag.Diagnostic{Severity: diag.Error, Summary: "summary2", Detail: "detail2"},
			},
			want: 1,
		},
		{
			testName: "multiple error Diagnostics",
			diags: diag.Diagnostics{
				diag.Diagnostic{Severity: diag.Error, Summary: "summary1", Detail: "detail1"},
				diag.Diagnostic{Severity: diag.Error, Summary: "summary2", Detail: "detail2"},
			},
			want: 2,
		},
		{
			testName: "multiple warning Diagnostics",
			diags: diag.Diagnostics{
				diag.Diagnostic{Severity: diag.Warning, Summary: "summary1", Detail: "detail1"},
				diag.Diagnostic{Severity: diag.Warning, Summary: "summary2", Detail: "detail2"},
			},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.testName, func(t *testing.T) {
			t.Parallel()

			output := sdkdiag.Errors(testCase.diags)

			if got, want := len(output), testCase.want; got != want {
				t.Errorf("Errors = %v, want %v", got, want)
			}
		})
	}
}

func TestWarnings(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		testName string
		diags    diag.Diagnostics
		want     int
	}{
		{
			testName: "nil Diagnostics",
		},
		{
			testName: "single warning Diagnostics",
			diags:    diag.Diagnostics{diag.Diagnostic{Severity: diag.Warning, Summary: "summary", Detail: "detail"}},
			want:     1,
		},
		{
			testName: "single error Diagnostics",
			diags:    diag.Diagnostics{diag.Diagnostic{Severity: diag.Error, Summary: "summary", Detail: "detail"}},
		},
		{
			testName: "mixed warning and error Diagnostics",
			diags: diag.Diagnostics{
				diag.Diagnostic{Severity: diag.Warning, Summary: "summary1", Detail: "detail1"},
				diag.Diagnostic{Severity: diag.Error, Summary: "summary2", Detail: "detail2"},
			},
			want: 1,
		},
		{
			testName: "multiple error Diagnostics",
			diags: diag.Diagnostics{
				diag.Diagnostic{Severity: diag.Error, Summary: "summary1", Detail: "detail1"},
				diag.Diagnostic{Severity: diag.Error, Summary: "summary2", Detail: "detail2"},
			},
		},
		{
			testName: "multiple warning Diagnostics",
			diags: diag.Diagnostics{
				diag.Diagnostic{Severity: diag.Warning, Summary: "summary1", Detail: "detail1"},
				diag.Diagnostic{Severity: diag.Warning, Summary: "summary2", Detail: "detail2"},
			},
			want: 2,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.testName, func(t *testing.T) {
			t.Parallel()

			output := sdkdiag.Warnings(testCase.diags)

			if got, want := len(output), testCase.want; got != want {
				t.Errorf("Warnings = %v, want %v", got, want)
			}
		})
	}
}

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
			diags:    diag.Diagnostics{diag.Diagnostic{Severity: diag.Warning, Summary: "summary", Detail: "detail"}},
		},
		{
			testName: "single error Diagnostics",
			diags:    diag.Diagnostics{diag.Diagnostic{Severity: diag.Error, Summary: "summary", Detail: "detail"}},
			wantErr:  true,
		},
		{
			testName: "mixed warning and error Diagnostics",
			diags: diag.Diagnostics{
				diag.Diagnostic{Severity: diag.Warning, Summary: "summary1", Detail: "detail1"},
				diag.Diagnostic{Severity: diag.Error, Summary: "summary2", Detail: "detail2"},
			},
			wantErr: true,
		},
		{
			testName: "multiple error Diagnostics",
			diags: diag.Diagnostics{
				diag.Diagnostic{Severity: diag.Error, Summary: "summary1", Detail: "detail1"},
				diag.Diagnostic{Severity: diag.Error, Summary: "summary2", Detail: "detail2"},
			},
			wantErr: true,
		},
		{
			testName: "multiple warning Diagnostics",
			diags: diag.Diagnostics{
				diag.Diagnostic{Severity: diag.Warning, Summary: "summary1", Detail: "detail1"},
				diag.Diagnostic{Severity: diag.Warning, Summary: "summary2", Detail: "detail2"},
			},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.testName, func(t *testing.T) {
			t.Parallel()

			err := sdkdiag.DiagnosticsError(testCase.diags)
			gotErr := err != nil

			if gotErr != testCase.wantErr {
				t.Errorf("gotErr = %v, wantErr = %v", gotErr, testCase.wantErr)
			}
		})
	}
}
