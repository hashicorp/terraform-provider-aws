// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package arczonalshift

import (
	"context"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/arczonalshift"
	awstypes "github.com/aws/aws-sdk-go-v2/service/arczonalshift/types"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	sweepfw "github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
	awsv2.Register("aws_arczonalshift_zonal_autoshift_configuration", sweepZonalAutoshiftConfigurations)
}

func sweepZonalAutoshiftConfigurations(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	input := arczonalshift.ListManagedResourcesInput{}
	conn := client.ARCZonalShiftClient(ctx)
	var sweepResources []sweep.Sweepable

	pages := arczonalshift.NewListManagedResourcesPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, smarterr.NewError(err)
		}

		for _, v := range page.Items {
			if v.ZonalAutoshiftStatus == awstypes.ZonalAutoshiftStatusEnabled || v.PracticeRunStatus != "" {
				sweepResources = append(sweepResources, sweepfw.NewSweepResource(newResourceZonalAutoshiftConfiguration, client,
					sweepfw.NewAttribute(names.AttrResourceARN, aws.ToString(v.Arn))),
				)
			}
		}
	}

	return sweepResources, nil
}
