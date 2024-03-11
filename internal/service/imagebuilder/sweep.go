// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package imagebuilder

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/imagebuilder"
	"github.com/aws/aws-sdk-go-v2/service/imagebuilder/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func RegisterSweepers() {
	resource.AddTestSweepers("aws_imagebuilder_component", &resource.Sweeper{
		Name: "aws_imagebuilder_component",
		F:    sweepComponents,
	})

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
}

func sweepComponents(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.ImageBuilderClient(ctx)

	input := &imagebuilder.ListComponentsInput{
		Owner: types.OwnershipSelf,
	}

	sweepResources := make([]sweep.Sweepable, 0)

	pages := imagebuilder.NewListComponentsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping ImageBuilder Components sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing ImageBuilder Components (%s): %w", region, err)
		}

		for _, v := range page.ComponentVersionList {
			arn := aws.ToString(v.Arn)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return fmt.Errorf("error reading ImageBuilder Components (%s): %w", arn, err)
			}

			r := ResourceComponent()
			d := r.Data(nil)
			d.SetId(arn)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping ImageBuilder Components (%s): %w", region, err)
	}

	return nil
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
			log.Printf("[WARN] Skipping ImageBuilder Distribution Configuration Summary sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing ImageBuilder Distribution Configuration Summary (%s): %w", region, err)
		}

		for _, v := range page.DistributionConfigurationSummaryList {
			arn := aws.ToString(v.Arn)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return fmt.Errorf("error reading ImageBuilder Distribution Configuration Summary (%s): %w", arn, err)
			}

			r := ResourceDistributionConfiguration()
			d := r.Data(nil)
			d.SetId(arn)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping ImageBuilder Distribution Configuration Summary (%s): %w", region, err)
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
			log.Printf("[WARN] Skipping ImageBuilder Image Pipelines sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing ImageBuilder Image Pipelines (%s): %w", region, err)
		}

		for _, v := range page.ImagePipelineList {
			arn := aws.ToString(v.Arn)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return fmt.Errorf("error reading ImageBuilder Image Pipelines (%s): %w", arn, err)
			}

			r := ResourceImagePipeline()
			d := r.Data(nil)
			d.SetId(arn)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping ImageBuilder Image Pipelines (%s): %w", region, err)
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
			log.Printf("[WARN] Skipping ImageBuilder Image Recipes sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing ImageBuilder Image Recipes (%s): %w", region, err)
		}

		for _, v := range page.ImageRecipeSummaryList {
			arn := aws.ToString(v.Arn)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return fmt.Errorf("error reading ImageBuilder Image Recipes (%s): %w", arn, err)
			}

			r := ResourceImageRecipe()
			d := r.Data(nil)
			d.SetId(arn)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping ImageBuilder Image Recipes (%s): %w", region, err)
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
			log.Printf("[WARN] Skipping ImageBuilder Container Recipes sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing ImageBuilder Container Recipes (%s): %w", region, err)
		}

		for _, v := range page.ContainerRecipeSummaryList {
			arn := aws.ToString(v.Arn)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return fmt.Errorf("error reading ImageBuilder Container Recipes (%s): %w", arn, err)
			}

			r := ResourceContainerRecipe()
			d := r.Data(nil)
			d.SetId(arn)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping ImageBuilder Container Recipes (%s): %w", region, err)
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
			log.Printf("[WARN] Skipping ImageBuilder Images sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing ImageBuilder Images (%s): %w", region, err)
		}

		for _, v := range page.ImageVersionList {
			arn := aws.ToString(v.Arn)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return fmt.Errorf("error reading ImageBuilder Images (%s): %w", arn, err)
			}

			r := ResourceImage()
			d := r.Data(nil)
			d.SetId(arn)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping ImageBuilder Images (%s): %w", region, err)
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
			log.Printf("[WARN] Skipping ImageBuilder Infrastructure Configurations sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing ImageBuilder Infrastructure Configurations (%s): %w", region, err)
		}

		for _, v := range page.InfrastructureConfigurationSummaryList {
			arn := aws.ToString(v.Arn)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return fmt.Errorf("error reading ImageBuilder Infrastructure Configurations (%s): %w", arn, err)
			}

			r := ResourceInfrastructureConfiguration()
			d := r.Data(nil)
			d.SetId(arn)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping ImageBuilder Infrastructure Configurations (%s): %w", region, err)
	}

	return nil
}
