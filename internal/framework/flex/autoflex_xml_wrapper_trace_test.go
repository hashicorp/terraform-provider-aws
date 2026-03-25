// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package flex

import (
	"bytes"
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflogtest"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
)

// Test to capture trace logs for Rule 2 expansion
func TestTraceRule2Expansion(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// AWS types (Rule 2: Items + Quantity + Flying)
	type Parrots struct {
		Flying   *bool
		Quantity *int32
		Items    []string
	}
	type Birds struct {
		Parrots *Parrots
	}

	// TF models
	type parrotModel struct {
		Flying types.Bool                        `tfsdk:"flying"`
		Items  fwtypes.ListValueOf[types.String] `tfsdk:"items" autoflex:",xmlwrapper=Items"`
	}
	type birdModel struct {
		Parrot fwtypes.ListNestedObjectValueOf[parrotModel] `tfsdk:"parrot"`
	}

	source := &birdModel{
		Parrot: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []parrotModel{
			{
				Flying: types.BoolValue(true),
				Items: fwtypes.NewListValueOfMust[types.String](ctx, []attr.Value{
					types.StringValue("macaw"),
				}),
			},
		}),
	}

	target := &Birds{}

	var buf bytes.Buffer
	ctx = tflogtest.RootLogger(ctx, &buf)
	ctx = registerTestingLogger(ctx)

	diags := Expand(ctx, source, target)
	if diags.HasError() {
		t.Fatalf("Expand failed: %v", diags)
	}

	// Print logs
	t.Logf("\n=== CAPTURED LOGS ===\n%s\n", buf.String())

	// Check result
	if target.Parrots == nil {
		t.Fatal("Expected non-nil Parrots")
	}
	if target.Parrots.Quantity == nil {
		t.Error("Expected Quantity to be set, got nil")
	} else if *target.Parrots.Quantity != 1 {
		t.Errorf("Expected Quantity=1, got %d", *target.Parrots.Quantity)
	}

	expected := &Birds{Parrots: &Parrots{
		Flying:   aws.Bool(true),
		Items:    []string{"macaw"},
		Quantity: aws.Int32(1),
	}}

	t.Logf("\nExpected: %+v", expected.Parrots)
	t.Logf("Got:      %+v", target.Parrots)
}
