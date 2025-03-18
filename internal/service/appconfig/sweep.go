// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appconfig

import (
	"context"
	"fmt"
	"log"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/appconfig"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
	awsv2.Register("aws_appconfig_application", sweepApplications, "aws_appconfig_configuration_profile", "aws_appconfig_environment", "aws_appconfig_extension_association")
	awsv2.Register("aws_appconfig_configuration_profile", sweepConfigurationProfiles, "aws_appconfig_extension_association", "aws_appconfig_hosted_configuration_version")
	awsv2.Register("aws_appconfig_deployment_strategy", sweepDeploymentStrategies)
	awsv2.Register("aws_appconfig_extension", sweepExtensions, "aws_appconfig_extension_association")

	resource.AddTestSweepers("aws_appconfig_environment", &resource.Sweeper{
		Name: "aws_appconfig_environment",
		F:    sweepEnvironments,
		Dependencies: []string{
			"aws_appconfig_extension_association",
		},
	})

	resource.AddTestSweepers("aws_appconfig_hosted_configuration_version", &resource.Sweeper{
		Name: "aws_appconfig_hosted_configuration_version",
		F:    sweepHostedConfigurationVersions,
	})

	resource.AddTestSweepers("aws_appconfig_extension_association", &resource.Sweeper{
		Name: "aws_appconfig_extension_association",
		F:    sweepExtensionAssociations,
	})
}

func sweepApplications(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.AppConfigClient(ctx)
	var input appconfig.ListApplicationsInput
	sweepResources := make([]sweep.Sweepable, 0)

	pages := appconfig.NewListApplicationsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.Items {
			r := resourceApplication()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.Id))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, nil
}

func sweepConfigurationProfiles(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.AppConfigClient(ctx)
	var input appconfig.ListApplicationsInput
	sweepResources := make([]sweep.Sweepable, 0)

	pages := appconfig.NewListApplicationsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.Items {
			applicationID := aws.ToString(v.Id)
			input := appconfig.ListConfigurationProfilesInput{
				ApplicationId: aws.String(applicationID),
			}

			pages := appconfig.NewListConfigurationProfilesPaginator(conn, &input)
			for pages.HasMorePages() {
				page, err := pages.NextPage(ctx)

				if err != nil {
					return nil, err
				}

				for _, v := range page.Items {
					r := resourceConfigurationProfile()
					d := r.Data(nil)
					d.SetId(configurationProfileCreateResourceID(aws.ToString(v.Id), applicationID))

					sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
				}
			}
		}
	}

	return sweepResources, nil
}

func sweepDeploymentStrategies(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.AppConfigClient(ctx)
	var input appconfig.ListDeploymentStrategiesInput
	sweepResources := make([]sweep.Sweepable, 0)

	pages := appconfig.NewListDeploymentStrategiesPaginator(conn, &input)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.Items {
			id := aws.ToString(v.Id)

			// Deleting AppConfig Predefined Strategies is not supported; returns BadRequestException
			if regexache.MustCompile(`^AppConfig\.[0-9A-Za-z]{9,40}$`).MatchString(id) {
				log.Printf("[DEBUG] Skipping AppConfig Deployment Strategy (%s): predefined strategy cannot be deleted", id)
				continue
			}

			r := resourceDeploymentStrategy()
			d := r.Data(nil)
			d.SetId(id)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, nil
}

func sweepEnvironments(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.AppConfigClient(ctx)
	input := &appconfig.ListApplicationsInput{}
	sweepResources := make([]sweep.Sweepable, 0)
	var sweeperErrs *multierror.Error

	pages := appconfig.NewListApplicationsPaginator(conn, input)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping AppConfig Environment sweep for %s: %s", region, err)
			return sweeperErrs.ErrorOrNil()
		}

		if err != nil {
			sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing AppConfig Applications (%s): %w", region, err))
		}

		for _, v := range page.Items {
			appID := aws.ToString(v.Id)
			input := &appconfig.ListEnvironmentsInput{
				ApplicationId: aws.String(appID),
			}

			pages := appconfig.NewListEnvironmentsPaginator(conn, input)

			for pages.HasMorePages() {
				page, err := pages.NextPage(ctx)

				if awsv2.SkipSweepError(err) {
					continue
				}

				if err != nil {
					sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing AppConfig Environments (%s): %w", region, err))
				}

				for _, v := range page.Items {
					sweepResources = append(sweepResources, framework.NewSweepResource(newResourceEnvironment, client,
						framework.NewAttribute(names.AttrApplicationID, aws.ToString(v.ApplicationId)),
						framework.NewAttribute("environment_id", aws.ToString(v.Id)),
					))
				}
			}
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error sweeping AppConfig Environments (%s): %w", region, err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepExtensions(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.AppConfigClient(ctx)
	var input appconfig.ListExtensionsInput
	sweepResources := make([]sweep.Sweepable, 0)

	pages := appconfig.NewListExtensionsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.Items {
			r := resourceExtension()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.Id))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, nil
}

func sweepHostedConfigurationVersions(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.AppConfigClient(ctx)
	input := &appconfig.ListApplicationsInput{}
	sweepResources := make([]sweep.Sweepable, 0)
	var sweeperErrs *multierror.Error

	pages := appconfig.NewListApplicationsPaginator(conn, input)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping AppConfig Hosted Configuration Version sweep for %s: %s", region, err)
			return sweeperErrs.ErrorOrNil()
		}

		if err != nil {
			sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing AppConfig Applications (%s): %w", region, err))
		}

		for _, v := range page.Items {
			appID := aws.ToString(v.Id)
			input := &appconfig.ListConfigurationProfilesInput{
				ApplicationId: aws.String(appID),
			}

			pages := appconfig.NewListConfigurationProfilesPaginator(conn, input)

			for pages.HasMorePages() {
				page, err := pages.NextPage(ctx)

				if awsv2.SkipSweepError(err) {
					continue
				}

				if err != nil {
					sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing AppConfig Hosted Configuration Versions (%s): %w", region, err))
				}

				for _, v := range page.Items {
					profileID := aws.ToString(v.Id)
					input := &appconfig.ListHostedConfigurationVersionsInput{
						ApplicationId:          aws.String(appID),
						ConfigurationProfileId: aws.String(profileID),
					}

					pages := appconfig.NewListHostedConfigurationVersionsPaginator(conn, input)

					for pages.HasMorePages() {
						page, err := pages.NextPage(ctx)

						if awsv2.SkipSweepError(err) {
							continue
						}

						if err != nil {
							sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing AppConfig Configuration Versions (%s): %w", region, err))
						}

						for _, v := range page.Items {
							r := ResourceHostedConfigurationVersion()
							d := r.Data(nil)
							d.SetId(fmt.Sprintf("%s/%s/%d", appID, profileID, v.VersionNumber))

							sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
						}
					}
				}
			}
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error sweeping AppConfig Hosted Configuration Versions (%s): %w", region, err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepExtensionAssociations(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.AppConfigClient(ctx)
	input := &appconfig.ListExtensionAssociationsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := appconfig.NewListExtensionAssociationsPaginator(conn, input)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping AppConfig Extension Association sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing AppConfig Extension Associations (%s): %w", region, err)
		}

		for _, v := range page.Items {
			r := ResourceExtensionAssociation()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.Id))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping AppConfig Extension Associations (%s): %w", region, err)
	}

	return nil
}
