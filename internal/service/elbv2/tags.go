// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package elbv2

import (
	"context"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
)

type tagDescriber interface {
	DescribeTags(ctx context.Context, params *elasticloadbalancingv2.DescribeTagsInput, optFns ...func(*elasticloadbalancingv2.Options)) (*elasticloadbalancingv2.DescribeTagsOutput, error)
}

func batchListTags(ctx context.Context, conn tagDescriber, identifiers []string, optFns ...func(*elasticloadbalancingv2.Options)) (map[string][]awstypes.Tag, error) {
	if len(identifiers) == 0 {
		return make(map[string][]awstypes.Tag, 0), nil
	}

	input := elasticloadbalancingv2.DescribeTagsInput{
		ResourceArns: identifiers,
	}

	output, err := conn.DescribeTags(ctx, &input, optFns...)

	if err != nil {
		return make(map[string][]awstypes.Tag, 0), smarterr.NewError(err)
	}

	tags := make(map[string][]awstypes.Tag, len(output.TagDescriptions))
	for _, v := range output.TagDescriptions {
		tags[aws.ToString(v.ResourceArn)] = v.Tags
	}

	return tags, nil
}
