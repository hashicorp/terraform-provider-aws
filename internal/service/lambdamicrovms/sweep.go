// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package lambdamicrovms

import (
	"context"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambdamicrovms"
	awstypes "github.com/aws/aws-sdk-go-v2/service/lambdamicrovms/types"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	sweepfw "github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
	awsv2.Register("aws_lambdamicrovms_image", sweepImages)
	awsv2.Register("aws_lambdamicrovms_microvm", sweepMicrovms)
}

func sweepImages(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	input := lambdamicrovms.ListMicrovmImagesInput{}
	conn := client.LambdaMicroVMsClient(ctx)
	var sweepResources []sweep.Sweepable

	pages := lambdamicrovms.NewListMicrovmImagesPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, smarterr.NewError(err)
		}

		for _, v := range page.Items {
			sweepResources = append(sweepResources, sweepfw.NewSweepResource(newImageResource, client,
				sweepfw.NewAttribute(names.AttrARN, aws.ToString(v.ImageArn))),
			)
		}
	}

	return sweepResources, nil
}

func sweepMicrovms(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	input := lambdamicrovms.ListMicrovmsInput{}
	conn := client.LambdaMicroVMsClient(ctx)
	var sweepResources []sweep.Sweepable

	pages := lambdamicrovms.NewListMicrovmsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, smarterr.NewError(err)
		}

		for _, v := range page.Items {
			// Terminated MicroVMs remain listable until they age out; skip them
			// so the sweeper does not churn on resources that are already gone.
			if v.State == awstypes.MicrovmStateTerminated {
				continue
			}

			sweepResources = append(sweepResources, sweepfw.NewSweepResource(newMicrovmResource, client,
				sweepfw.NewAttribute("microvm_id", aws.ToString(v.MicrovmId))),
			)
		}
	}

	return sweepResources, nil
}
