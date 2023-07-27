// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build sweep
// +build sweep

package imagebuilder

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/imagebuilder"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
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
	conn := client.ImageBuilderConn(ctx)

	sweepResources := make([]sweep.Sweepable, 0)
	var sweeperErrs *multierror.Error

	input := &imagebuilder.ListComponentsInput{
		Owner: aws.String(imagebuilder.OwnershipSelf),
	}
	err = conn.ListComponentsPagesWithContext(ctx, input, func(page *imagebuilder.ListComponentsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, componentVersion := range page.ComponentVersionList {
			if componentVersion == nil {
				continue
			}

			arn := aws.StringValue(componentVersion.Arn)
			input := &imagebuilder.ListComponentBuildVersionsInput{
				ComponentVersionArn: componentVersion.Arn,
			}

			err := conn.ListComponentBuildVersionsPagesWithContext(ctx, input, func(page *imagebuilder.ListComponentBuildVersionsOutput, lastPage bool) bool {
				if page == nil {
					return !lastPage
				}

				for _, componentSummary := range page.ComponentSummaryList {
					if componentSummary == nil {
						continue
					}

					arn := aws.StringValue(componentSummary.Arn)

					r := ResourceComponent()
					d := r.Data(nil)
					d.SetId(arn)

					sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
				}

				return !lastPage
			})

			if err != nil {
				sweeperErr := fmt.Errorf("error listing Image Builder Component (%s) versions: %w", arn, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Image Builder Component sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}
	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing Image Builder Components: %w", err))
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error sweeping Image Builder Components: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepDistributionConfigurations(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.ImageBuilderConn(ctx)

	sweepResources := make([]sweep.Sweepable, 0)
	var sweeperErrs *multierror.Error

	input := &imagebuilder.ListDistributionConfigurationsInput{}
	err = conn.ListDistributionConfigurationsPagesWithContext(ctx, input, func(page *imagebuilder.ListDistributionConfigurationsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, distributionConfigurationSummary := range page.DistributionConfigurationSummaryList {
			if distributionConfigurationSummary == nil {
				continue
			}

			arn := aws.StringValue(distributionConfigurationSummary.Arn)

			r := ResourceDistributionConfiguration()
			d := r.Data(nil)
			d.SetId(arn)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Image Builder Distribution Configuration sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}
	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing Image Builder Distribution Configurations: %w", err))
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error sweeping Image Builder Distribution Configurations: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepImagePipelines(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.ImageBuilderConn(ctx)

	sweepResources := make([]sweep.Sweepable, 0)
	var sweeperErrs *multierror.Error

	input := &imagebuilder.ListImagePipelinesInput{}
	err = conn.ListImagePipelinesPagesWithContext(ctx, input, func(page *imagebuilder.ListImagePipelinesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, imagePipeline := range page.ImagePipelineList {
			if imagePipeline == nil {
				continue
			}

			arn := aws.StringValue(imagePipeline.Arn)

			r := ResourceImagePipeline()
			d := r.Data(nil)
			d.SetId(arn)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Image Builder Image Pipeline sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}
	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing Image Builder Image Pipelines: %w", err))
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error sweeping Image Builder Image Pipelines: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepImageRecipes(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.ImageBuilderConn(ctx)

	sweepResources := make([]sweep.Sweepable, 0)
	var sweeperErrs *multierror.Error

	input := &imagebuilder.ListImageRecipesInput{
		Owner: aws.String(imagebuilder.OwnershipSelf),
	}
	err = conn.ListImageRecipesPagesWithContext(ctx, input, func(page *imagebuilder.ListImageRecipesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, imageRecipeSummary := range page.ImageRecipeSummaryList {
			if imageRecipeSummary == nil {
				continue
			}

			arn := aws.StringValue(imageRecipeSummary.Arn)

			r := ResourceImageRecipe()
			d := r.Data(nil)
			d.SetId(arn)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Image Builder Image Recipe sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}
	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing Image Builder Image Recipes: %w", err))
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error sweeping Image Builder Image Recipes: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepContainerRecipes(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.ImageBuilderConn(ctx)

	sweepResources := make([]sweep.Sweepable, 0)
	var sweeperErrs *multierror.Error

	input := &imagebuilder.ListContainerRecipesInput{
		Owner: aws.String(imagebuilder.OwnershipSelf),
	}
	err = conn.ListContainerRecipesPagesWithContext(ctx, input, func(page *imagebuilder.ListContainerRecipesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, containerRecipeSummary := range page.ContainerRecipeSummaryList {
			if containerRecipeSummary == nil {
				continue
			}

			arn := aws.StringValue(containerRecipeSummary.Arn)

			r := ResourceContainerRecipe()
			d := r.Data(nil)
			d.SetId(arn)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Image Builder Container Recipe sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}
	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing Image Builder Container Recipes: %w", err))
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error sweeping Image Builder Container Recipes: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepImages(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.ImageBuilderConn(ctx)
	sweepResources := make([]sweep.Sweepable, 0)
	var errs *multierror.Error

	input := &imagebuilder.ListImagesInput{
		Owner: aws.String(imagebuilder.OwnershipSelf),
	}

	err = conn.ListImagesPagesWithContext(ctx, input, func(page *imagebuilder.ListImagesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, imageVersion := range page.ImageVersionList {
			if imageVersion == nil {
				continue
			}

			// Retrieve the Image's Build Version ARNs required as input
			// to the ResourceImage()'s Delete operation
			// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/19851
			imageVersionArn := aws.StringValue(imageVersion.Arn)

			input := &imagebuilder.ListImageBuildVersionsInput{
				ImageVersionArn: imageVersion.Arn,
			}

			err := conn.ListImageBuildVersionsPagesWithContext(ctx, input, func(page *imagebuilder.ListImageBuildVersionsOutput, lastPage bool) bool {
				if page == nil {
					return !lastPage
				}

				for _, imageSummary := range page.ImageSummaryList {
					if imageSummary == nil {
						continue
					}

					imageBuildVersionArn := aws.StringValue(imageSummary.Arn)

					r := ResourceImage()
					d := r.Data(nil)
					d.SetId(imageBuildVersionArn)

					sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
				}

				return !lastPage
			})

			if err != nil {
				errs = multierror.Append(errs, fmt.Errorf("error listing Image Builder Image Build Versions for image (%s): %w", imageVersionArn, err))
			}
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error listing Image Builder Images for %s: %w", region, err))
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping Image Builder Images for %s: %w", region, err))
	}

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Image Builder Image sweep for %s: %s", region, err)
		return nil
	}

	return errs.ErrorOrNil()
}

func sweepInfrastructureConfigurations(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.ImageBuilderConn(ctx)

	sweepResources := make([]sweep.Sweepable, 0)
	var sweeperErrs *multierror.Error

	input := &imagebuilder.ListInfrastructureConfigurationsInput{}

	err = conn.ListInfrastructureConfigurationsPagesWithContext(ctx, input, func(page *imagebuilder.ListInfrastructureConfigurationsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, infrastructureConfigurationSummary := range page.InfrastructureConfigurationSummaryList {
			if infrastructureConfigurationSummary == nil {
				continue
			}

			arn := aws.StringValue(infrastructureConfigurationSummary.Arn)

			r := ResourceInfrastructureConfiguration()
			d := r.Data(nil)
			d.SetId(arn)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Image Builder Infrastructure Configuration sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}
	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing Image Builder Infrastructure Configurations: %w", err))
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error sweeping Image Builder Infrastructure Configurations: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}
