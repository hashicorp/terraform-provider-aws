// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package rekognition

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rekognition"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
	awsv2.Register("aws_rekognition_collection", sweepCollections)
	awsv2.Register("aws_rekognition_project", sweepProjects)
	awsv2.Register("aws_rekognition_stream_processor", sweepStreamProcessors)
}

func sweepCollections(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.RekognitionClient(ctx)
	var sweepResources []sweep.Sweepable
	var input rekognition.ListCollectionsInput

	pages := rekognition.NewListCollectionsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, v := range page.CollectionIds {
			sweepResources = append(sweepResources, framework.NewSweepResource(newCollectionResource, client,
				framework.NewAttribute(names.AttrID, v),
			))
		}
	}

	return sweepResources, nil
}

func sweepProjects(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.RekognitionClient(ctx)
	var sweepResources []sweep.Sweepable
	var input rekognition.DescribeProjectsInput

	pages := rekognition.NewDescribeProjectsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, v := range page.ProjectDescriptions {
			sweepResources = append(sweepResources, framework.NewSweepResource(newProjectResource, client,
				framework.NewAttribute(names.AttrID, v.ProjectArn),
			))
		}
	}

	return sweepResources, nil
}

func sweepStreamProcessors(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.RekognitionClient(ctx)
	var sweepResources []sweep.Sweepable
	var input rekognition.ListStreamProcessorsInput

	pages := rekognition.NewListStreamProcessorsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, v := range page.StreamProcessors {
			sweepResources = append(sweepResources, framework.NewSweepResource(newStreamProcessorResource, client,
				framework.NewAttribute(names.AttrName, aws.ToString(v.Name)),
			))
		}
	}

	return sweepResources, nil
}
