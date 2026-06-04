// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package lakeformation

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
)

func TestDataCellsFilterPreserveEmptyColumnWildcardExcludedColumnNames(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	emptySet := fwtypes.NewSetValueOfMust[types.String](ctx, []attr.Value{})
	nonEmptySet := fwtypes.NewSetValueOfMust[types.String](ctx, []attr.Value{
		types.StringValue("column_name"),
	})

	t.Run("preserves empty prior set", func(t *testing.T) {
		current := tableData{
			ColumnWildcard: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &columnWildcard{
				ExcludedColumnNames: fwtypes.NewSetValueOfNull[types.String](ctx),
			}),
		}
		prior := tableData{
			ColumnWildcard: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &columnWildcard{
				ExcludedColumnNames: emptySet,
			}),
		}

		diags := current.preserveEmptyColumnWildcardExcludedColumnNames(ctx, prior)
		if diags.HasError() {
			t.Fatalf("unexpected diagnostics: %s", diags.Errors())
		}

		currentColumnWildcard, diags := current.ColumnWildcard.ToPtr(ctx)
		if diags.HasError() {
			t.Fatalf("unexpected diagnostics: %s", diags.Errors())
		}

		if !currentColumnWildcard.ExcludedColumnNames.Equal(emptySet) {
			t.Fatalf("expected empty excluded column names to be preserved, got %s", currentColumnWildcard.ExcludedColumnNames.String())
		}
	})

	t.Run("ignores non-empty prior set", func(t *testing.T) {
		current := tableData{
			ColumnWildcard: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &columnWildcard{
				ExcludedColumnNames: fwtypes.NewSetValueOfNull[types.String](ctx),
			}),
		}
		prior := tableData{
			ColumnWildcard: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &columnWildcard{
				ExcludedColumnNames: nonEmptySet,
			}),
		}

		diags := current.preserveEmptyColumnWildcardExcludedColumnNames(ctx, prior)
		if diags.HasError() {
			t.Fatalf("unexpected diagnostics: %s", diags.Errors())
		}

		currentColumnWildcard, diags := current.ColumnWildcard.ToPtr(ctx)
		if diags.HasError() {
			t.Fatalf("unexpected diagnostics: %s", diags.Errors())
		}

		if !currentColumnWildcard.ExcludedColumnNames.IsNull() {
			t.Fatalf("expected non-empty prior excluded column names not to be copied, got %s", currentColumnWildcard.ExcludedColumnNames.String())
		}
	})

	t.Run("keeps current set", func(t *testing.T) {
		current := tableData{
			ColumnWildcard: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &columnWildcard{
				ExcludedColumnNames: nonEmptySet,
			}),
		}
		prior := tableData{
			ColumnWildcard: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &columnWildcard{
				ExcludedColumnNames: emptySet,
			}),
		}

		diags := current.preserveEmptyColumnWildcardExcludedColumnNames(ctx, prior)
		if diags.HasError() {
			t.Fatalf("unexpected diagnostics: %s", diags.Errors())
		}

		currentColumnWildcard, diags := current.ColumnWildcard.ToPtr(ctx)
		if diags.HasError() {
			t.Fatalf("unexpected diagnostics: %s", diags.Errors())
		}

		if !currentColumnWildcard.ExcludedColumnNames.Equal(nonEmptySet) {
			t.Fatalf("expected current excluded column names to be preserved, got %s", currentColumnWildcard.ExcludedColumnNames.String())
		}
	})
}
