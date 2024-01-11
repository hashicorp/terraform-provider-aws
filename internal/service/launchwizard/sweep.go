// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package launchwizard

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/launchwizard"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
)

func RegisterSweepers() {
	resource.AddTestSweepers("aws_launchwizard_deployment", &resource.Sweeper{
		Name: "aws_launchwizard_deployment",
		F:    sweepDeployments,
	})
}

func sweepDeployments(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.LaunchWizardClient(ctx)
	input := &launchwizard.ListDeploymentsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := launchwizard.NewListDeploymentsPaginator(conn, input)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Launchwizard Deployment sweep for %s: %s", region, err)
			return nil
		}
		if err != nil {
			return fmt.Errorf("error retrieving Launchwizard Deployment: %w", err)
		}

		for _, deployment := range page.Deployments {
			id := aws.ToString(deployment.Id)

			log.Printf("[INFO] Deleting Launchwizard Deployment: %s", id)
			sweepResources = append(sweepResources, framework.NewSweepResource(newResourceDeployment, client,
				framework.NewAttribute("id", id),
			))
		}
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		return fmt.Errorf("error sweeping Launchwizard Deployment for %s: %w", region, err)
	}

	return nil
}
