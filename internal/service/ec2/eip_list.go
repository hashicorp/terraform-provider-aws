// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
)

// @SDKListResource("aws_eip")
func newEIPResourceAsListResource() inttypes.ListResourceForSDK {
	return framework.NewSDKListResource(resourceEIP(), framework.SDKListSpec[awstypes.Address, framework.RegionQuery]{
		Items: func(ctx context.Context, c *conns.AWSClient, _ framework.RegionQuery) (iter.Seq2[awstypes.Address, error], diag.Diagnostics) {
			var input ec2.DescribeAddressesInput

			return listEIPs(ctx, c.EC2Client(ctx), &input), nil
		},
		ID: func(a awstypes.Address, _ framework.RegionQuery) string {
			return aws.ToString(a.AllocationId)
		},
		Prepare: func(ctx context.Context, a awstypes.Address, _ framework.RegionQuery, _ *schema.ResourceData) {
			setTagsOut(ctx, a.Tags)
		},
		Flatten: resourceEIPFlatten,
	})
}

func listEIPs(ctx context.Context, conn *ec2.Client, input *ec2.DescribeAddressesInput) iter.Seq2[awstypes.Address, error] {
	return func(yield func(awstypes.Address, error) bool) {
		items, err := findEIPs(ctx, conn, input)
		if err != nil {
			yield(awstypes.Address{}, fmt.Errorf("listing EC2 EIPs: %w", err))
			return
		}

		for _, item := range items {
			if !yield(item, nil) {
				return
			}
		}
	}
}
