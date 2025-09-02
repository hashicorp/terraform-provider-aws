// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssmquicksetup

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssmquicksetup"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
)

func RegisterSweepers() {
	awsv2.Register("aws_ssmquicksetup_configuration_manager", sweepConfigurationManagers)
}

func sweepConfigurationManagers(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.SSMQuickSetupClient(ctx)
	var input ssmquicksetup.ListConfigurationManagersInput
	sweepResources := make([]sweep.Sweepable, 0)

	pages := ssmquicksetup.NewListConfigurationManagersPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.ConfigurationManagersList {
			sweepResources = append(sweepResources, framework.NewSweepResource(newConfigurationManagerResource, client,
				framework.NewAttribute("manager_arn", aws.ToString(v.ManagerArn))))
		}
	}

	return sweepResources, nil
}
