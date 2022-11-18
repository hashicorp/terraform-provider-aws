package validators

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestValidateInt64String(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		request       tfsdk.ValidateAttributeRequest
		expectedInt64 int64
		expectedOk    bool
	}{
		"invalid-type": {
			request: tfsdk.ValidateAttributeRequest{
				AttributeConfig:         types.BoolValue(true),
				AttributePath:           path.Root("test"),
				AttributePathExpression: path.MatchRoot("test"),
			},
			expectedInt64: 0,
			expectedOk:    false,
		},
		"string-null": {
			request: tfsdk.ValidateAttributeRequest{
				AttributeConfig:         types.Int64Null(),
				AttributePath:           path.Root("test"),
				AttributePathExpression: path.MatchRoot("test"),
			},
			expectedInt64: 0,
			expectedOk:    false,
		},
		"string-valid-value": {
			request: tfsdk.ValidateAttributeRequest{
				AttributeConfig:         types.StringValue("43"),
				AttributePath:           path.Root("test"),
				AttributePathExpression: path.MatchRoot("test"),
			},
			expectedInt64: 43,
			expectedOk:    true,
		},
		"string-invalid-value": {
			request: tfsdk.ValidateAttributeRequest{
				AttributeConfig:         types.StringValue("test-value"),
				AttributePath:           path.Root("test"),
				AttributePathExpression: path.MatchRoot("test"),
			},
			expectedInt64: 0,
			expectedOk:    false,
		},
		"string-unknown": {
			request: tfsdk.ValidateAttributeRequest{
				AttributeConfig:         types.Int64Unknown(),
				AttributePath:           path.Root("test"),
				AttributePathExpression: path.MatchRoot("test"),
			},
			expectedInt64: 0,
			expectedOk:    false,
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			gotInt64, gotOk := validateInt64String(context.Background(), testCase.request, &tfsdk.ValidateAttributeResponse{})

			if diff := cmp.Diff(gotInt64, testCase.expectedInt64); diff != "" {
				t.Errorf("unexpected int64-string difference: %s", diff)
			}

			if diff := cmp.Diff(gotOk, testCase.expectedOk); diff != "" {
				t.Errorf("unexpected ok difference: %s", diff)
			}
		})
	}
}

func TestValidateString(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		request        tfsdk.ValidateAttributeRequest
		expectedString string
		expectedOk     bool
	}{
		"invalid-type": {
			request: tfsdk.ValidateAttributeRequest{
				AttributeConfig:         types.BoolValue(true),
				AttributePath:           path.Root("test"),
				AttributePathExpression: path.MatchRoot("test"),
			},
			expectedString: "",
			expectedOk:     false,
		},
		"string-null": {
			request: tfsdk.ValidateAttributeRequest{
				AttributeConfig:         types.Int64Null(),
				AttributePath:           path.Root("test"),
				AttributePathExpression: path.MatchRoot("test"),
			},
			expectedString: "",
			expectedOk:     false,
		},
		"string-value": {
			request: tfsdk.ValidateAttributeRequest{
				AttributeConfig:         types.StringValue("test-value"),
				AttributePath:           path.Root("test"),
				AttributePathExpression: path.MatchRoot("test"),
			},
			expectedString: "test-value",
			expectedOk:     true,
		},
		"string-unknown": {
			request: tfsdk.ValidateAttributeRequest{
				AttributeConfig:         types.Int64Unknown(),
				AttributePath:           path.Root("test"),
				AttributePathExpression: path.MatchRoot("test"),
			},
			expectedString: "",
			expectedOk:     false,
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			gotString, gotOk := validateString(context.Background(), testCase.request, &tfsdk.ValidateAttributeResponse{})

			if diff := cmp.Diff(gotString, testCase.expectedString); diff != "" {
				t.Errorf("unexpected string difference: %s", diff)
			}

			if diff := cmp.Diff(gotOk, testCase.expectedOk); diff != "" {
				t.Errorf("unexpected ok difference: %s", diff)
			}
		})
	}
}
