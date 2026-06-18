// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package elasticache

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/elasticache/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
)

// TestReservedCacheNodeFlexOptsMapsID is a regression test for
// https://github.com/hashicorp/terraform-provider-aws/issues/47523.
//
// flexOpts() returns flex.WithFieldNamePrefix("ReservedCacheNode"), which is
// what makes AutoFlex map the AWS SDK field ReservedCacheNodeId to the model's
// ID field. Without it, data.ID stays null after Flatten and the resource is
// silently removed from state on the next refresh.
func TestReservedCacheNodeFlexOptsMapsID(t *testing.T) {
	t.Parallel()

	const reservationID = "tf-reservation-test"

	reservation := &awstypes.ReservedCacheNode{
		ReservedCacheNodeId: aws.String(reservationID),
	}

	r := &reservedCacheNodeResource{}

	t.Run("with flexOpts", func(t *testing.T) {
		t.Parallel()

		var data reservedCacheNodeResourceModel
		if diags := flex.Flatten(context.Background(), reservation, &data, r.flexOpts()...); diags.HasError() {
			t.Fatalf("Flatten returned diagnostics: %s", diags)
		}
		if got, want := data.ID.ValueString(), reservationID; got != want {
			t.Errorf("data.ID = %q, want %q (regression of #47523)", got, want)
		}
	})

	t.Run("without flexOpts", func(t *testing.T) {
		t.Parallel()

		var data reservedCacheNodeResourceModel
		if diags := flex.Flatten(context.Background(), reservation, &data); diags.HasError() {
			t.Fatalf("Flatten returned diagnostics: %s", diags)
		}
		// Documents the broken state: without the prefix option AutoFlex finds
		// no field for ReservedCacheNodeId, so data.ID stays null. If a future
		// AutoFlex change ever makes this case work without flexOpts, this
		// assertion fails and the regression test should be revisited.
		if !data.ID.IsNull() {
			t.Errorf("data.ID = %q, want null (without flexOpts the prefix mapping is gone)", data.ID.ValueString())
		}
	})
}
