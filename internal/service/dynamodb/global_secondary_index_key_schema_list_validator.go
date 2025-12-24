// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package dynamodb

import (
	"context"
	"fmt"

	awstypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
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
	keySchema := fwdiag.Must(typ.ValueFromList(ctx, request.ConfigValue)).(fwtypes.ListNestedObjectValueOf[keySchemaModel])

	var hashCount, rangeCount int
	for _, v := range fwdiag.Must(keySchema.ToSlice(ctx)) {
		switch v.KeyType.ValueEnum() {
		case awstypes.KeyTypeHash:
			hashCount++

		case awstypes.KeyTypeRange:
			rangeCount++
		}
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
