// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package lambdamicrovms

import (
	"context"
	"log"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambdamicrovms"
	awstypes "github.com/aws/aws-sdk-go-v2/service/lambdamicrovms/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/endpoints"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	sweepfw "github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
	awsv2.Register("aws_lambdamicrovms_image", sweepImages, "aws_lambdamicrovms_microvm")
	awsv2.Register("aws_lambdamicrovms_microvm", sweepMicroVMs)
}

func sweepImages(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.LambdaMicroVMsClient(ctx)
	var input lambdamicrovms.ListMicrovmImagesInput
	var sweepResources []sweep.Sweepable

	if region := client.Region(ctx); region == endpoints.UsWest1RegionID {
		log.Printf("[WARN] Skipping Lambda MicroVMs Micro VM sweep for region: %s", region)
		return sweepResources, nil
	}

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

func sweepMicroVMs(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.LambdaMicroVMsClient(ctx)
	var input lambdamicrovms.ListMicrovmsInput
	var sweepResources []sweep.Sweepable

	if region := client.Region(ctx); region == endpoints.UsWest1RegionID {
		log.Printf("[WARN] Skipping Lambda MicroVMs Micro VM sweep for region: %s", region)
		return sweepResources, nil
	}

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

			sweepResources = append(sweepResources, sweepfw.NewSweepResource(newMicroVMResource, client,
				sweepfw.NewAttribute("microvm_id", aws.ToString(v.MicrovmId))),
			)
		}
	}

	return sweepResources, nil
}
