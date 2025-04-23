// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fis

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/fis"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
)

func RegisterSweepers() {
	awsv2.Register("aws_fis_experiment_template", sweepExperimentTemplates)
}

func sweepExperimentTemplates(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.FISClient(ctx)
	var input fis.ListExperimentTemplatesInput
	sweepResources := make([]sweep.Sweepable, 0)

	pages := fis.NewListExperimentTemplatesPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.ExperimentTemplates {
			r := resourceExperimentTemplate()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.Id))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, nil
}
