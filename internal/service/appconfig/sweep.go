// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build sweep
// +build sweep

package appconfig

import (
	"fmt"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appconfig"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
)

func init() {
	resource.AddTestSweepers("aws_appconfig_application", &resource.Sweeper{
		Name: "aws_appconfig_application",
		F:    sweepApplications,
		Dependencies: []string{
			"aws_appconfig_configuration_profile",
			"aws_appconfig_environment",
		},
	})

	resource.AddTestSweepers("aws_appconfig_configuration_profile", &resource.Sweeper{
		Name: "aws_appconfig_configuration_profile",
		F:    sweepConfigurationProfiles,
		Dependencies: []string{
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
	})

	resource.AddTestSweepers("aws_appconfig_hosted_configuration_version", &resource.Sweeper{
		Name: "aws_appconfig_hosted_configuration_version",
		F:    sweepHostedConfigurationVersions,
	})
}

func sweepApplications(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.AppConfigConn(ctx)
	sweepResources := make([]sweep.Sweepable, 0)
	var errs *multierror.Error

	input := &appconfig.ListApplicationsInput{}

	err = conn.ListApplicationsPagesWithContext(ctx, input, func(page *appconfig.ListApplicationsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, item := range page.Items {
			if item == nil {
				continue
			}

			id := aws.StringValue(item.Id)

			log.Printf("[INFO] Deleting AppConfig Application (%s)", id)
			r := ResourceApplication()
			d := r.Data(nil)
			d.SetId(id)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error listing AppConfig Applications: %w", err))
	}

	if err = sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping AppConfig Applications for %s: %w", region, err))
	}

	if sweep.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping AppConfig Applications sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func sweepConfigurationProfiles(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.AppConfigConn(ctx)
	sweepResources := make([]sweep.Sweepable, 0)
	var errs *multierror.Error

	input := &appconfig.ListApplicationsInput{}

	err = conn.ListApplicationsPagesWithContext(ctx, input, func(page *appconfig.ListApplicationsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, item := range page.Items {
			if item == nil {
				continue
			}

			appId := aws.StringValue(item.Id)

			profilesInput := &appconfig.ListConfigurationProfilesInput{
				ApplicationId: item.Id,
			}

			err := conn.ListConfigurationProfilesPagesWithContext(ctx, profilesInput, func(page *appconfig.ListConfigurationProfilesOutput, lastPage bool) bool {
				if page == nil {
					return !lastPage
				}

				for _, item := range page.Items {
					if item == nil {
						continue
					}

					id := fmt.Sprintf("%s:%s", aws.StringValue(item.Id), appId)

					log.Printf("[INFO] Deleting AppConfig Configuration Profile (%s)", id)
					r := ResourceConfigurationProfile()
					d := r.Data(nil)
					d.SetId(id)

					sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
				}

				return !lastPage
			})

			if err != nil {
				errs = multierror.Append(errs, fmt.Errorf("error listing AppConfig Configuration Profiles for Application (%s): %w", appId, err))
			}
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error listing AppConfig Applications: %w", err))
	}

	if err = sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping AppConfig Configuration Profiles for %s: %w", region, err))
	}

	if sweep.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping AppConfig Configuration Profiles sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func sweepDeploymentStrategies(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.AppConfigConn(ctx)
	sweepResources := make([]sweep.Sweepable, 0)
	var errs *multierror.Error

	input := &appconfig.ListDeploymentStrategiesInput{}

	err = conn.ListDeploymentStrategiesPagesWithContext(ctx, input, func(page *appconfig.ListDeploymentStrategiesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, item := range page.Items {
			if item == nil {
				continue
			}

			id := aws.StringValue(item.Id)

			// Deleting AppConfig Predefined Strategies is not supported; returns BadRequestException
			if regexp.MustCompile(`^AppConfig\.[A-Za-z0-9]{9,40}$`).MatchString(id) {
				log.Printf("[DEBUG] Skipping AppConfig Deployment Strategy (%s): predefined strategy cannot be deleted", id)
				continue
			}

			log.Printf("[INFO] Deleting AppConfig Deployment Strategy (%s)", id)
			r := ResourceDeploymentStrategy()
			d := r.Data(nil)
			d.SetId(id)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error listing AppConfig Deployment Strategies: %w", err))
	}

	if err = sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping AppConfig Deployment Strategies for %s: %w", region, err))
	}

	if sweep.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping AppConfig Deployment Strategies sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func sweepEnvironments(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.AppConfigConn(ctx)
	sweepResources := make([]sweep.Sweepable, 0)
	var errs *multierror.Error

	input := &appconfig.ListApplicationsInput{}

	err = conn.ListApplicationsPagesWithContext(ctx, input, func(page *appconfig.ListApplicationsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, item := range page.Items {
			if item == nil {
				continue
			}

			appId := aws.StringValue(item.Id)

			envInput := &appconfig.ListEnvironmentsInput{
				ApplicationId: item.Id,
			}

			err := conn.ListEnvironmentsPagesWithContext(ctx, envInput, func(page *appconfig.ListEnvironmentsOutput, lastPage bool) bool {
				if page == nil {
					return !lastPage
				}

				for _, item := range page.Items {
					if item == nil {
						continue
					}

					sweepResources = append(sweepResources, framework.NewSweepResource(newResourceEnvironment, client,
						framework.NewAttribute("application_id", aws.StringValue(item.ApplicationId)),
						framework.NewAttribute("environment_id", aws.StringValue(item.Id)),
					))
				}

				return !lastPage
			})

			if err != nil {
				errs = multierror.Append(errs, fmt.Errorf("error listing AppConfig Environments for Application (%s): %w", appId, err))
			}
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error listing AppConfig Applications: %w", err))
	}

	if err = sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping AppConfig Environments for %s: %w", region, err))
	}

	if sweep.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping AppConfig Environments sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func sweepHostedConfigurationVersions(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.AppConfigConn(ctx)
	sweepResources := make([]sweep.Sweepable, 0)
	var errs *multierror.Error

	input := &appconfig.ListApplicationsInput{}

	err = conn.ListApplicationsPagesWithContext(ctx, input, func(page *appconfig.ListApplicationsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, item := range page.Items {
			if item == nil {
				continue
			}

			appId := aws.StringValue(item.Id)

			profilesInput := &appconfig.ListConfigurationProfilesInput{
				ApplicationId: item.Id,
			}

			err := conn.ListConfigurationProfilesPagesWithContext(ctx, profilesInput, func(page *appconfig.ListConfigurationProfilesOutput, lastPage bool) bool {
				if page == nil {
					return !lastPage
				}

				for _, item := range page.Items {
					if item == nil {
						continue
					}

					profId := aws.StringValue(item.Id)

					versionInput := &appconfig.ListHostedConfigurationVersionsInput{
						ApplicationId:          aws.String(appId),
						ConfigurationProfileId: aws.String(profId),
					}

					err := conn.ListHostedConfigurationVersionsPagesWithContext(ctx, versionInput, func(page *appconfig.ListHostedConfigurationVersionsOutput, lastPage bool) bool {
						if page == nil {
							return !lastPage
						}

						for _, item := range page.Items {
							if item == nil {
								continue
							}

							id := fmt.Sprintf("%s/%s/%d", appId, profId, aws.Int64Value(item.VersionNumber))

							log.Printf("[INFO] Deleting AppConfig Hosted Configuration Version (%s)", id)
							r := ResourceHostedConfigurationVersion()
							d := r.Data(nil)
							d.SetId(id)

							sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
						}

						return !lastPage
					})

					if err != nil {
						errs = multierror.Append(errs, fmt.Errorf("error listing AppConfig Hosted Configuration Versions for Application (%s) and Configuration Profile (%s): %w", appId, profId, err))
					}
				}

				return !lastPage
			})

			if err != nil {
				errs = multierror.Append(errs, fmt.Errorf("error listing AppConfig Configuration Profiles for Application (%s): %w", appId, err))
			}
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error listing AppConfig Applications: %w", err))
	}

	if err = sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping AppConfig Hosted Configuration Versions for %s: %w", region, err))
	}

	if sweep.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping AppConfig Hosted Configuration Versions sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}
