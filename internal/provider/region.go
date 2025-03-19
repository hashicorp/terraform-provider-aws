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
		if err := validateRegionInPartition(ctx, meta.(*conns.AWSClient), v.(string)); err != nil {
			return err
		}
	}

	return nil
}

func validateRegionInPartition(ctx context.Context, c *conns.AWSClient, region string) error {
	if got, want := names.PartitionForRegion(region).ID(), c.Partition(ctx); got != want {
		return fmt.Errorf("partition (%s) for per-resource Region (%s) is not the provider's configured partition (%s)", got, region, want)
	}

	return nil
}

// defaultRegionValue is a CustomizeDiff function that sets the value of the top-level `region`
// attribute to the provider's configured Region if it is not set.
func defaultRegionValue(ctx context.Context, d *schema.ResourceDiff, meta any) error {
	if _, ok := d.GetOk(names.AttrRegion); !ok {
		return d.SetNew(names.AttrRegion, meta.(*conns.AWSClient).AwsConfig(ctx).Region)
	}

	return nil
}
