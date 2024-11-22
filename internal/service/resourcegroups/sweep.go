// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package resourcegroups

import (
	"context"
	"slices"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/resourcegroups"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
	awsv2.Register("aws_resourcegroups_group", sweepGroups,
		"aws_servicecatalogappregistry_application",
	)
}

func sweepGroups(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.ResourceGroupsClient(ctx)

	var sweepResources []sweep.Sweepable

	r := resourceGroup()
	pages := resourcegroups.NewListGroupsPaginator(conn, &resourcegroups.ListGroupsInput{})
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, v := range page.GroupIdentifiers {
			tags, err := listTags(ctx, conn, aws.ToString(v.GroupArn))
			if err != nil {
				tflog.Warn(ctx, "Skipping resource", map[string]any{
					"error": err.Error(),
				})
				continue
			}

			if slices.Contains(tags.Keys(), "aws:servicecatalog:applicationId") {
				tflog.Warn(ctx, "Skipping resource", map[string]any{
					"skip_reason":       "managed by AppRegistry",
					names.AttrGroupName: aws.ToString(v.GroupName),
				})
				continue
			}

			d := r.Data(nil)
			d.SetId(aws.ToString(v.GroupName))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, nil
}
