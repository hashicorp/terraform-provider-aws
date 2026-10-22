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
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
)

// @SDKListResource("aws_key_pair")
func newKeyPairResourceAsListResource() inttypes.ListResourceForSDK {
	return framework.NewSDKListResource(resourceKeyPair(), framework.SDKListSpec[awstypes.KeyPairInfo, framework.RegionQuery]{
		Items: func(ctx context.Context, c *conns.AWSClient, _ framework.RegionQuery) (iter.Seq2[awstypes.KeyPairInfo, error], diag.Diagnostics) {
			var input ec2.DescribeKeyPairsInput

			return listKeyPairs(ctx, c.EC2Client(ctx), &input), nil
		},
		ID: func(kp awstypes.KeyPairInfo, _ framework.RegionQuery) string {
			return aws.ToString(kp.KeyName)
		},
		Flatten: resourceKeyPairFlatten,
	})
}

func listKeyPairs(ctx context.Context, conn *ec2.Client, input *ec2.DescribeKeyPairsInput) iter.Seq2[awstypes.KeyPairInfo, error] {
	return func(yield func(awstypes.KeyPairInfo, error) bool) {
		items, err := findKeyPairs(ctx, conn, input)
		if err != nil {
			yield(awstypes.KeyPairInfo{}, fmt.Errorf("listing EC2 Key Pair resources: %w", err))
			return
		}

		for _, item := range items {
			if !yield(item, nil) {
				return
			}
		}
	}
}
