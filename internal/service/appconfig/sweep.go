// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appconfig

import (
	"fmt"
	"log"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appconfig"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv1"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
)

func RegisterSweepers() {
	resource.AddTestSweepers("aws_appconfig_application", &resource.Sweeper{
		Name: "aws_appconfig_application",
		F:    sweepApplications,
		Dependencies: []string{
			"aws_appconfig_configuration_profile",
			"aws_appconfig_environment",
			"aws_appconfig_extension_association",
		},
	})

	resource.AddTestSweepers("aws_appconfig_configuration_profile", &resource.Sweeper{
		Name: "aws_appconfig_configuration_profile",
		F:    sweepConfigurationProfiles,
		Dependencies: []string{
			"aws_appconfig_extension_association",
			"aws_appconfig_hosted_configuration_version",
		},
	})

	resource.AddTestSweepers("aws_appconfig_deployment_strategy", &resource.Sweeper{
		Name: "aws_appconfig_deployment_strategy",
		F:    sweepDeploymentStrategies,
	})

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

func sweepApplications(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.AppConfigConn(ctx)
	input := &appconfig.ListApplicationsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = conn.ListApplicationsPagesWithContext(ctx, input, func(page *appconfig.ListApplicationsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Items {
			r := ResourceApplication()
			d := r.Data(nil)
			d.SetId(aws.StringValue(v.Id))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if awsv1.SkipSweepError(err) {
		log.Printf("[WARN] Skipping AppConfig Application sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing AppConfig Applications (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping AppConfig Applications (%s): %w", region, err)
	}

	return nil
}

func sweepConfigurationProfiles(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.AppConfigConn(ctx)
	input := &appconfig.ListApplicationsInput{}
	sweepResources := make([]sweep.Sweepable, 0)
	var sweeperErrs *multierror.Error

	err = conn.ListApplicationsPagesWithContext(ctx, input, func(page *appconfig.ListApplicationsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Items {
			appID := aws.StringValue(v.Id)
			input := &appconfig.ListConfigurationProfilesInput{
				ApplicationId: aws.String(appID),
			}

			err := conn.ListConfigurationProfilesPagesWithContext(ctx, input, func(page *appconfig.ListConfigurationProfilesOutput, lastPage bool) bool {
				for _, v := range page.Items {
					r := ResourceConfigurationProfile()
					d := r.Data(nil)
					d.SetId(fmt.Sprintf("%s:%s", aws.StringValue(v.Id), appID))

					sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
				}

				return !lastPage
			})

			if awsv1.SkipSweepError(err) {
				continue
			}

			if err != nil {
				sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing AppConfig Configuration Profiles (%s): %w", region, err))
			}
		}

		return !lastPage
	})

	if awsv1.SkipSweepError(err) {
		log.Printf("[WARN] Skipping AppConfig Configuration Profile sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil()
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing AppConfig Applications (%s): %w", region, err))
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error sweeping AppConfig Configuration Profiles (%s): %w", region, err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepDeploymentStrategies(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.AppConfigConn(ctx)
	input := &appconfig.ListDeploymentStrategiesInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = conn.ListDeploymentStrategiesPagesWithContext(ctx, input, func(page *appconfig.ListDeploymentStrategiesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Items {
			id := aws.StringValue(v.Id)

			// Deleting AppConfig Predefined Strategies is not supported; returns BadRequestException
			if regexache.MustCompile(`^AppConfig\.[0-9A-Za-z]{9,40}$`).MatchString(id) {
				log.Printf("[DEBUG] Skipping AppConfig Deployment Strategy (%s): predefined strategy cannot be deleted", id)
				continue
			}

			r := ResourceDeploymentStrategy()
			d := r.Data(nil)
			d.SetId(id)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if awsv1.SkipSweepError(err) {
		log.Printf("[WARN] Skipping AppConfig Deployment Strategy sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing AppConfig Deployment Strategies (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping AppConfig Deployment Strategies (%s): %w", region, err)
	}

	return nil
}

func sweepEnvironments(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.AppConfigConn(ctx)
	input := &appconfig.ListApplicationsInput{}
	sweepResources := make([]sweep.Sweepable, 0)
	var sweeperErrs *multierror.Error

	err = conn.ListApplicationsPagesWithContext(ctx, input, func(page *appconfig.ListApplicationsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Items {
			appID := aws.StringValue(v.Id)
			input := &appconfig.ListEnvironmentsInput{
				ApplicationId: aws.String(appID),
			}

			err := conn.ListEnvironmentsPagesWithContext(ctx, input, func(page *appconfig.ListEnvironmentsOutput, lastPage bool) bool {
				if page == nil {
					return !lastPage
				}

				for _, v := range page.Items {
					sweepResources = append(sweepResources, framework.NewSweepResource(newResourceEnvironment, client,
						framework.NewAttribute("application_id", aws.StringValue(v.ApplicationId)),
						framework.NewAttribute("environment_id", aws.StringValue(v.Id)),
					))
				}

				return !lastPage
			})

			if awsv1.SkipSweepError(err) {
				continue
			}

			if err != nil {
				sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing AppConfig Environments (%s): %w", region, err))
			}
		}

		return !lastPage
	})

	if awsv1.SkipSweepError(err) {
		log.Printf("[WARN] Skipping AppConfig Environment sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil()
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing AppConfig Applications (%s): %w", region, err))
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error sweeping AppConfig Environments (%s): %w", region, err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepHostedConfigurationVersions(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.AppConfigConn(ctx)
	input := &appconfig.ListApplicationsInput{}
	sweepResources := make([]sweep.Sweepable, 0)
	var sweeperErrs *multierror.Error

	err = conn.ListApplicationsPagesWithContext(ctx, input, func(page *appconfig.ListApplicationsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Items {
			appID := aws.StringValue(v.Id)
			input := &appconfig.ListConfigurationProfilesInput{
				ApplicationId: aws.String(appID),
			}

			err := conn.ListConfigurationProfilesPagesWithContext(ctx, input, func(page *appconfig.ListConfigurationProfilesOutput, lastPage bool) bool {
				if page == nil {
					return !lastPage
				}

				for _, v := range page.Items {
					profileID := aws.StringValue(v.Id)
					input := &appconfig.ListHostedConfigurationVersionsInput{
						ApplicationId:          aws.String(appID),
						ConfigurationProfileId: aws.String(profileID),
					}

					err := conn.ListHostedConfigurationVersionsPagesWithContext(ctx, input, func(page *appconfig.ListHostedConfigurationVersionsOutput, lastPage bool) bool {
						if page == nil {
							return !lastPage
						}

						for _, v := range page.Items {
							r := ResourceHostedConfigurationVersion()
							d := r.Data(nil)
							d.SetId(fmt.Sprintf("%s/%s/%d", appID, profileID, aws.Int64Value(v.VersionNumber)))

							sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
						}

						return !lastPage
					})

					if awsv1.SkipSweepError(err) {
						continue
					}

					if err != nil {
						sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing AppConfig Hosted Configuration Versions (%s): %w", region, err))
					}
				}

				return !lastPage
			})

			if awsv1.SkipSweepError(err) {
				continue
			}

			if err != nil {
				sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing AppConfig Configuration Profiles (%s): %w", region, err))
			}
		}

		return !lastPage
	})

	if awsv1.SkipSweepError(err) {
		log.Printf("[WARN] Skipping AppConfig Hosted Configuration Version sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil()
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing AppConfig Applications (%s): %w", region, err))
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
	conn := client.AppConfigConn(ctx)
	input := &appconfig.ListExtensionAssociationsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = conn.ListExtensionAssociationsPagesWithContext(ctx, input, func(page *appconfig.ListExtensionAssociationsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Items {
			r := ResourceExtensionAssociation()
			d := r.Data(nil)
			d.SetId(aws.StringValue(v.Id))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if awsv1.SkipSweepError(err) {
		log.Printf("[WARN] Skipping AppConfig Extension Association sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing AppConfig Extension Associations (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping AppConfig Extension Associations (%s): %w", region, err)
	}

	return nil
}
