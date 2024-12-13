// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package imagebuilder

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/imagebuilder"
	"github.com/aws/aws-sdk-go-v2/service/imagebuilder/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
	awsv2.Register("aws_imagebuilder_component", sweepComponents)

	resource.AddTestSweepers("aws_imagebuilder_distribution_configuration", &resource.Sweeper{
		Name: "aws_imagebuilder_distribution_configuration",
		F:    sweepDistributionConfigurations,
	})

	resource.AddTestSweepers("aws_imagebuilder_image_pipeline", &resource.Sweeper{
		Name: "aws_imagebuilder_image_pipeline",
		F:    sweepImagePipelines,
	})

	resource.AddTestSweepers("aws_imagebuilder_image_recipe", &resource.Sweeper{
		Name: "aws_imagebuilder_image_recipe",
		F:    sweepImageRecipes,
	})

	resource.AddTestSweepers("aws_imagebuilder_container_recipe", &resource.Sweeper{
		Name: "aws_imagebuilder_container_recipe",
		F:    sweepContainerRecipes,
	})

	resource.AddTestSweepers("aws_imagebuilder_image", &resource.Sweeper{
		Name: "aws_imagebuilder_image",
		F:    sweepImages,
	})

	resource.AddTestSweepers("aws_imagebuilder_infrastructure_configuration", &resource.Sweeper{
		Name: "aws_imagebuilder_infrastructure_configuration",
		F:    sweepInfrastructureConfigurations,
	})

	resource.AddTestSweepers("aws_imagebuilder_lifecycle_policy", &resource.Sweeper{
		Name: "aws_imagebuilder_lifecycle_policy",
		F:    sweepLifecyclePolicies,
	})
}

func sweepComponents(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.ImageBuilderClient(ctx)

	var sweepResources []sweep.Sweepable

	r := resourceComponent()
	input := imagebuilder.ListComponentsInput{
		Owner: types.OwnershipSelf,
	}
	pages := imagebuilder.NewListComponentsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, v := range page.ComponentVersionList {
			// The Delete operation needs the Component Build Version ARN, not just the Component Version ARN
			input := imagebuilder.ListComponentBuildVersionsInput{
				ComponentVersionArn: v.Arn,
			}
			pages := imagebuilder.NewListComponentBuildVersionsPaginator(conn, &input)
			for pages.HasMorePages() {
				page, err := pages.NextPage(ctx)
				if err != nil {
					return nil, err
				}

				for _, v := range page.ComponentSummaryList {
					d := r.Data(nil)
					d.SetId(aws.ToString(v.Arn))

					sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
				}
			}
		}
	}

	return sweepResources, nil
}

func sweepDistributionConfigurations(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.ImageBuilderClient(ctx)
	input := &imagebuilder.ListDistributionConfigurationsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := imagebuilder.NewListDistributionConfigurationsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Image Builder Distribution Configuration sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing Image Builder Distribution Configurations (%s): %w", region, err)
		}

		for _, v := range page.DistributionConfigurationSummaryList {
			r := resourceDistributionConfiguration()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.Arn))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Image Builder Distribution Configuration Summary (%s): %w", region, err)
	}

	return nil
}

func sweepImagePipelines(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.ImageBuilderClient(ctx)
	input := &imagebuilder.ListImagePipelinesInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := imagebuilder.NewListImagePipelinesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Image Builder Image Pipeline sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing Image Builder Image Pipelines (%s): %w", region, err)
		}

		for _, v := range page.ImagePipelineList {
			r := resourceImagePipeline()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.Arn))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Image Builder Image Pipelines (%s): %w", region, err)
	}

	return nil
}

func sweepImageRecipes(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.ImageBuilderClient(ctx)
	input := &imagebuilder.ListImageRecipesInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := imagebuilder.NewListImageRecipesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Image Builder Image Recipe sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing Image Builder Image Recipes (%s): %w", region, err)
		}

		for _, v := range page.ImageRecipeSummaryList {
			r := resourceImageRecipe()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.Arn))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Image Builder Image Recipes (%s): %w", region, err)
	}

	return nil
}

func sweepContainerRecipes(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.ImageBuilderClient(ctx)
	input := &imagebuilder.ListContainerRecipesInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := imagebuilder.NewListContainerRecipesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Image Builder Container Recipe sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing Image Builder Container Recipes (%s): %w", region, err)
		}

		for _, v := range page.ContainerRecipeSummaryList {
			r := resourceContainerRecipe()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.Arn))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Image Builder Container Recipes (%s): %w", region, err)
	}

	return nil
}

func sweepImages(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.ImageBuilderClient(ctx)
	input := &imagebuilder.ListImagesInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := imagebuilder.NewListImagesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Image Builder Image sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing Image Builder Images (%s): %w", region, err)
		}

		for _, v := range page.ImageVersionList {
			r := resourceImage()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.Arn))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Image Builder Images (%s): %w", region, err)
	}

	return nil
}

func sweepInfrastructureConfigurations(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.ImageBuilderClient(ctx)
	input := &imagebuilder.ListInfrastructureConfigurationsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := imagebuilder.NewListInfrastructureConfigurationsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Image Builder Infrastructure Configuration sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing Image Builder Infrastructure Configurations (%s): %w", region, err)
		}

		for _, v := range page.InfrastructureConfigurationSummaryList {
			r := resourceInfrastructureConfiguration()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.Arn))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Image Builder Infrastructure Configurations (%s): %w", region, err)
	}

	return nil
}

func sweepLifecyclePolicies(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.ImageBuilderClient(ctx)
	input := &imagebuilder.ListLifecyclePoliciesInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := imagebuilder.NewListLifecyclePoliciesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Image Builder Lifecycle Policy sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing Image Builder Lifecycle Policies (%s): %w", region, err)
		}

		for _, v := range page.LifecyclePolicySummaryList {
			sweepResources = append(sweepResources, framework.NewSweepResource(newLifecyclePolicyResource, client, framework.NewAttribute(names.AttrARN, v.Arn)))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Image Builder Lifecycle Policies (%s): %w", region, err)
	}

	return nil
}
