// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package fis

import (
	"context"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/fis"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	sweepfw "github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
	awsv2.Register("aws_fis_experiment_template", sweepExperimentTemplates)
	awsv2.Register("aws_fis_target_account_configuration", sweepTargetAccountConfigurations)
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

func sweepTargetAccountConfigurations(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.FISClient(ctx)
	var sweepResources []sweep.Sweepable

	experimentsInput := &fis.ListExperimentTemplatesInput{}
	experimentPages := fis.NewListExperimentTemplatesPaginator(conn, experimentsInput)

	for experimentPages.HasMorePages() {
		experimentPage, err := experimentPages.NextPage(ctx)
		if err != nil {
			return nil, smarterr.NewError(err)
		}

		for _, experiment := range experimentPage.ExperimentTemplates {
			input := &fis.ListTargetAccountConfigurationsInput{
				ExperimentTemplateId: experiment.Id,
			}

			pages := fis.NewListTargetAccountConfigurationsPaginator(conn, input)
			for pages.HasMorePages() {
				page, err := pages.NextPage(ctx)
				if err != nil {
					return nil, smarterr.NewError(err)
				}

				for _, v := range page.TargetAccountConfigurations {
					sweepResources = append(sweepResources, sweepfw.NewSweepResource(
						newResourceTargetAccountConfiguration,
						client,
						sweepfw.NewAttribute(names.AttrAccountID, aws.ToString(v.AccountId)),
						sweepfw.NewAttribute("experiment_template_id", aws.ToString(experiment.Id)),
					))
				}
			}
		}
	}

	return sweepResources, nil
}
