// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package dynamodb

import (
	"context"
	"fmt"

	awstypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
)

var _ validator.List = globalSecondaryIndexKeySchemaListValidator{}

type globalSecondaryIndexKeySchemaListValidator struct {
}

func (v globalSecondaryIndexKeySchemaListValidator) Description(_ context.Context) string {
	return ""
}

func (v globalSecondaryIndexKeySchemaListValidator) MarkdownDescription(_ context.Context) string {
	return ""
}

func (v globalSecondaryIndexKeySchemaListValidator) ValidateList(ctx context.Context, request validator.ListRequest, response *validator.ListResponse) {
	if request.ConfigValue.IsNull() || request.ConfigValue.IsUnknown() {
		response.Diagnostics.Append(validatordiag.InvalidAttributeValueDiagnostic(
			request.Path,
			fmt.Sprintf(`must contain at least %d and at most %d elements with a "key_type" of %q`, minNumberOfHashes, maxNumberOfHashes, awstypes.KeyTypeHash),
			"0",
		))
		return
	}

	typ := fwtypes.NewListNestedObjectTypeOf[keySchemaModel](ctx)
	keySchemaAttr := fwdiag.Must(typ.ValueFromList(ctx, request.ConfigValue)).(fwtypes.ListNestedObjectValueOf[keySchemaModel])

	keySchemas := fwdiag.Must(keySchemaAttr.ToSlice(ctx))

	if keySchemas[0].KeyType.ValueEnum() != awstypes.KeyTypeHash {
		elementPath := request.Path.AtListIndex(0)
		response.Diagnostics.Append(diag.NewAttributeErrorDiagnostic(
			elementPath,
			"Invalid Attribute Value",
			fmt.Sprintf(`The first element of %s must have a "key_type" of "`+string(awstypes.KeyTypeHash)+`", got %q`, request.Path, keySchemas[0].KeyType.ValueEnum()),
		))
		return
	}

	var hashCount, rangeCount int
	var lastKeyType awstypes.KeyType
	for i, v := range keySchemas {
		switch v.KeyType.ValueEnum() {
		case awstypes.KeyTypeHash:
			if lastKeyType == awstypes.KeyTypeRange {
				elementPath := request.Path.AtListIndex(i)
				response.Diagnostics.Append(diag.NewAttributeErrorDiagnostic(
					elementPath,
					"Invalid Attribute Value",
					fmt.Sprintf(`All elements of %s with "key_type" of "`+string(awstypes.KeyTypeHash)+`" must be before elements with "key_type" of "`+string(awstypes.KeyTypeRange)+`"`, request.Path),
				))
			}
			hashCount++

		case awstypes.KeyTypeRange:
			rangeCount++
		}
		lastKeyType = v.KeyType.ValueEnum()
	}

	if hashCount < minNumberOfHashes || hashCount > maxNumberOfHashes {
		response.Diagnostics.Append(validatordiag.InvalidAttributeValueDiagnostic(
			request.Path,
			fmt.Sprintf(`must contain at least %d and at most %d elements with a "key_type" of %q`, minNumberOfHashes, maxNumberOfHashes, awstypes.KeyTypeHash),
			fmt.Sprintf("%d", hashCount),
		))
	}

	if rangeCount > maxNumberOfRanges {
		response.Diagnostics.Append(validatordiag.InvalidAttributeValueDiagnostic(
			request.Path,
			fmt.Sprintf(`must contain at most %d elements with a "key_type" of %q`, maxNumberOfRanges, awstypes.KeyTypeRange),
			fmt.Sprintf("%d", rangeCount),
		))
	}
}
