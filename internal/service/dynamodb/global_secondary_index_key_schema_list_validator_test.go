// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package dynamodb

import (
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
)

func TestGlobalSecondaryIndexKeySchemaListValidator(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		value       fwtypes.ListNestedObjectValueOf[keySchemaModel]
		expectError bool
	}{
		"unknown": {
			value:       fwtypes.NewListNestedObjectValueOfUnknown[keySchemaModel](t.Context()),
			expectError: false,
		},
		"null": {
			// Null value is handled by `listvalidator.IsRequired()`
			value:       fwtypes.NewListNestedObjectValueOfNull[keySchemaModel](t.Context()),
			expectError: false,
		},
		"empty": {
			// Empty value is handled by `Required` on element attributes and `listvalidator.SizeAtLeast(1)`
			value:       fwtypes.NewListNestedObjectValueOfValueSliceMust(t.Context(), []keySchemaModel{}),
			expectError: false,
		},

		"fully known": {
			value: fwtypes.NewListNestedObjectValueOfValueSliceMust(t.Context(), []keySchemaModel{
				{
					AttributeName: types.StringValue("attribute_name"),
					AttributeType: fwtypes.StringEnumValue(awstypes.ScalarAttributeTypeS),
					KeyType:       fwtypes.StringEnumValue(awstypes.KeyTypeHash),
				},
			}),
			expectError: false,
		},

		"single unknown type": {
			value: fwtypes.NewListNestedObjectValueOfValueSliceMust(t.Context(), []keySchemaModel{
				{
					AttributeName: types.StringValue("attribute_name"),
					AttributeType: fwtypes.StringEnumValue(awstypes.ScalarAttributeTypeS),
					KeyType:       fwtypes.StringEnumUnknown[awstypes.KeyType](),
				},
			}),
			expectError: false,
		},

		"multiple unknown type": {
			value: fwtypes.NewListNestedObjectValueOfValueSliceMust(t.Context(), []keySchemaModel{
				{
					AttributeName: types.StringValue("attribute_name1"),
					AttributeType: fwtypes.StringEnumValue(awstypes.ScalarAttributeTypeS),
					KeyType:       fwtypes.StringEnumValue(awstypes.KeyTypeHash),
				},
				{
					AttributeName: types.StringValue("attribute_name2"),
					AttributeType: fwtypes.StringEnumValue(awstypes.ScalarAttributeTypeS),
					KeyType:       fwtypes.StringEnumUnknown[awstypes.KeyType](),
				},
			}),
			expectError: false,
		},
	}
	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			request := validator.ListRequest{
				Path:           path.Root("test"),
				PathExpression: path.MatchRoot("test"),
				ConfigValue:    testCase.value.ListValue,
			}
			response := validator.ListResponse{}

			globalSecondaryIndexKeySchemaListValidator{}.ValidateList(t.Context(), request, &response)

			if !response.Diagnostics.HasError() && testCase.expectError {
				t.Fatal("expected error, got no error")
			}

			if response.Diagnostics.HasError() && !testCase.expectError {
				t.Fatalf("got unexpected error: %s", response.Diagnostics)
			}
		})
	}
}
