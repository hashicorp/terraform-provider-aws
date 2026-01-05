// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package tagpolicy

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/resourcegroupstaggingapi"
	"github.com/aws/aws-sdk-go-v2/service/resourcegroupstaggingapi/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

func GetRequiredTags(ctx context.Context, awsConfig aws.Config) (map[string]tftags.KeyValueTags, error) {
	client := resourcegroupstaggingapi.NewFromConfig(awsConfig)
	paginator := resourcegroupstaggingapi.NewListRequiredTagsPaginator(client, &resourcegroupstaggingapi.ListRequiredTagsInput{})

	var reqTags []types.RequiredTag
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		reqTags = append(reqTags, page.RequiredTags...)
	}

	return convert(ctx, reqTags), nil
}

// convert translates the ListRequiredTags API response into a map of required
// tags per Terraform resource type
func convert(ctx context.Context, reqTags []types.RequiredTag) map[string]tftags.KeyValueTags {
	m := make(map[string]tftags.KeyValueTags)
	for _, t := range reqTags {
		tfTypes, ok := Lookup[aws.ToString(t.ResourceType)]
		if !ok {
			continue
		}

		newTags := tftags.New(ctx, t.ReportingTagKeys)
		for _, tfType := range tfTypes {
			if v, ok := m[tfType]; ok {
				m[tfType] = v.Merge(newTags)
			} else {
				m[tfType] = newTags
			}
		}
	}
	return m
}
