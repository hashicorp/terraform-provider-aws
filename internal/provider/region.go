// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// verifyRegionInConfiguredPartition is a CustomizeDiff function that verifies that the value of
// the top-level `region` attribute is in the configured AWS partition.
func verifyRegionInConfiguredPartition(ctx context.Context, d *schema.ResourceDiff, meta any) error {
	if v, ok := d.GetOk(names.AttrRegion); ok {
		region := v.(string)
		if got, want := names.PartitionForRegion(region).ID(), meta.(*conns.AWSClient).Partition(ctx); got != want {
			return fmt.Errorf("partition (%s) for per-resource Region (%s) is not the provider's configured partition (%s)", got, region, want)
		}
	}

	return nil
}

// forceNewIfRegionChanges is a CustomizeDiff function that forces resource replacement
// if the value of the top-level `region` attribute changes.
func forceNewIfRegionChanges(ctx context.Context, d *schema.ResourceDiff, meta any) error {
	if d.Id() != "" && d.HasChange(names.AttrRegion) {
		providerRegion := meta.(*conns.AWSClient).AwsConfig(ctx).Region
		o, n := d.GetChange(names.AttrRegion)
		if o, n := o.(string), n.(string); (o == "" && n == providerRegion) || (o == providerRegion && n == "") {
			return nil
		}
		return d.ForceNew(names.AttrRegion)
	}

	return nil
}
