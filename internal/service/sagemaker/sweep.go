// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sagemaker

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sagemaker"
	awstypes "github.com/aws/aws-sdk-go-v2/service/sagemaker/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/sdk"
)

func RegisterSweepers() {
	resource.AddTestSweepers("aws_sagemaker_app_image_config", &resource.Sweeper{
		Name: "aws_sagemaker_app_image_config",
		F:    sweepAppImagesConfig,
	})

	resource.AddTestSweepers("aws_sagemaker_app", &resource.Sweeper{
		Name: "aws_sagemaker_app",
		F:    sweepApps,
	})

	resource.AddTestSweepers("aws_sagemaker_code_repository", &resource.Sweeper{
		Name: "aws_sagemaker_code_repository",
		F:    sweepCodeRepositories,
	})

	resource.AddTestSweepers("aws_sagemaker_device_fleet", &resource.Sweeper{
		Name: "aws_sagemaker_device_fleet",
		F:    sweepDeviceFleets,
	})

	resource.AddTestSweepers("aws_sagemaker_domain", &resource.Sweeper{
		Name: "aws_sagemaker_domain",
		F:    sweepDomains,
		Dependencies: []string{
			"aws_efs_mount_target",
			"aws_efs_file_system",
			"aws_sagemaker_user_profile",
			"aws_sagemaker_space",
		},
	})

	resource.AddTestSweepers("aws_sagemaker_endpoint_configuration", &resource.Sweeper{
		Name: "aws_sagemaker_endpoint_configuration",
		Dependencies: []string{
			"aws_sagemaker_model",
		},
		F: sweepEndpointConfigurations,
	})

	resource.AddTestSweepers("aws_sagemaker_endpoint", &resource.Sweeper{
		Name: "aws_sagemaker_endpoint",
		Dependencies: []string{
			"aws_sagemaker_model",
			"aws_sagemaker_endpoint_configuration",
		},
		F: sweepEndpoints,
	})

	resource.AddTestSweepers("aws_sagemaker_feature_group", &resource.Sweeper{
		Name: "aws_sagemaker_feature_group",
		F:    sweepFeatureGroups,
	})

	resource.AddTestSweepers("aws_sagemaker_flow_definition", &resource.Sweeper{
		Name: "aws_sagemaker_flow_definition",
		F:    sweepFlowDefinitions,
	})

	resource.AddTestSweepers("aws_sagemaker_human_task_ui", &resource.Sweeper{
		Name: "aws_sagemaker_human_task_ui",
		F:    sweepHumanTaskUIs,
	})

	resource.AddTestSweepers("aws_sagemaker_image", &resource.Sweeper{
		Name: "aws_sagemaker_image",
		F:    sweepImages,
	})

	resource.AddTestSweepers("aws_sagemaker_mlflow_tracking_server", &resource.Sweeper{
		Name: "aws_sagemaker_mlflow_tracking_server",
		F:    sweepMlflowTrackingServers,
	})

	resource.AddTestSweepers("aws_sagemaker_model_package_group", &resource.Sweeper{
		Name: "aws_sagemaker_model_package_group",
		F:    sweepModelPackageGroups,
	})

	resource.AddTestSweepers("aws_sagemaker_model", &resource.Sweeper{
		Name: "aws_sagemaker_model",
		F:    sweepModels,
	})

	resource.AddTestSweepers("aws_sagemaker_notebook_instance_lifecycle_configuration", &resource.Sweeper{
		Name: "aws_sagemaker_notebook_instance_lifecycle_configuration",
		F:    sweepNotebookInstanceLifecycleConfiguration,
		Dependencies: []string{
			"aws_sagemaker_notebook_instance",
		},
	})

	resource.AddTestSweepers("aws_sagemaker_notebook_instance", &resource.Sweeper{
		Name: "aws_sagemaker_notebook_instance",
		F:    sweepNotebookInstances,
	})

	resource.AddTestSweepers("aws_sagemaker_studio_lifecycle_config", &resource.Sweeper{
		Name: "aws_sagemaker_studio_lifecycle_config",
		F:    sweepStudioLifecyclesConfig,
		Dependencies: []string{
			"aws_sagemaker_domain",
		},
	})

	resource.AddTestSweepers("aws_sagemaker_project", &resource.Sweeper{
		Name: "aws_sagemaker_project",
		F:    sweepProjects,
	})

	resource.AddTestSweepers("aws_sagemaker_space", &resource.Sweeper{
		Name: "aws_sagemaker_space",
		F:    sweepSpaces,
		Dependencies: []string{
			"aws_sagemaker_app",
		},
	})

	resource.AddTestSweepers("aws_sagemaker_user_profile", &resource.Sweeper{
		Name: "aws_sagemaker_user_profile",
		F:    sweepUserProfiles,
		Dependencies: []string{
			"aws_sagemaker_app",
		},
	})

	resource.AddTestSweepers("aws_sagemaker_workforce", &resource.Sweeper{
		Name: "aws_sagemaker_workforce",
		F:    sweepWorkforces,
		Dependencies: []string{
			"aws_sagemaker_workteam",
		},
	})

	resource.AddTestSweepers("aws_sagemaker_workteam", &resource.Sweeper{
		Name: "aws_sagemaker_workteam",
		F:    sweepWorkteams,
	})

	resource.AddTestSweepers("aws_sagemaker_pipeline", &resource.Sweeper{
		Name: "aws_sagemaker_pipeline",
		F:    sweepPipelines,
	})

	resource.AddTestSweepers("aws_sagemaker_hub", &resource.Sweeper{
		Name: "aws_sagemaker_hub",
		F:    sweepHubs,
	})
}

func sweepAppImagesConfig(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("getting client: %w", err)
	}
	conn := client.SageMakerClient(ctx)

	sweepResources := make([]sweep.Sweepable, 0)
	var sweeperErrs *multierror.Error

	pages := sagemaker.NewListAppImageConfigsPaginator(conn, &sagemaker.ListAppImageConfigsInput{})
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping SageMaker App Image Config for %s: %s", region, err)
			return sweeperErrs.ErrorOrNil()
		}

		if err != nil {
			sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("retrieving Example Thing: %w", err))
			return sweeperErrs
		}

		for _, config := range page.AppImageConfigs {
			name := aws.ToString(config.AppImageConfigName)
			r := resourceAppImageConfig()
			d := r.Data(nil)
			d.SetId(name)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("sweeping SageMaker App Image Configs: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepSpaces(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("getting client: %w", err)
	}
	conn := client.SageMakerClient(ctx)

	sweepResources := make([]sweep.Sweepable, 0)
	var sweeperErrs *multierror.Error

	pages := sagemaker.NewListSpacesPaginator(conn, &sagemaker.ListSpacesInput{})
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping SageMaker Space sweep for %s: %s", region, err)
			return sweeperErrs.ErrorOrNil()
		}
		if err != nil {
			sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("retrieving SageMaker Spaces: %w", err))
		}

		for _, space := range page.Spaces {
			r := resourceSpace()
			d := r.Data(nil)
			d.SetId(aws.ToString(space.SpaceName))
			d.Set("domain_id", space.DomainId)
			d.Set("space_name", space.SpaceName)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("sweeping SageMaker Spaces: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepApps(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("getting client: %w", err)
	}
	conn := client.SageMakerClient(ctx)

	sweepResources := make([]sweep.Sweepable, 0)
	var sweeperErrs *multierror.Error

	pages := sagemaker.NewListAppsPaginator(conn, &sagemaker.ListAppsInput{})
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping SageMaker App sweep for %s: %s", region, err)
			return sweeperErrs.ErrorOrNil()
		}
		if err != nil {
			sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("retrieving SageMaker Apps: %w", err))
		}

		for _, app := range page.Apps {
			if app.Status == awstypes.AppStatusDeleted {
				continue
			}

			r := resourceApp()
			d := r.Data(nil)
			d.SetId(aws.ToString(app.AppName))
			d.Set("app_name", app.AppName)
			d.Set("app_type", app.AppType)
			d.Set("domain_id", app.DomainId)
			d.Set("user_profile_name", app.UserProfileName)
			d.Set("space_name", app.SpaceName)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("sweeping SageMaker Apps: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepCodeRepositories(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("getting client: %s", err)
	}
	conn := client.SageMakerClient(ctx)

	sweepResources := make([]sweep.Sweepable, 0)
	var sweeperErrs *multierror.Error

	pages := sagemaker.NewListCodeRepositoriesPaginator(conn, &sagemaker.ListCodeRepositoriesInput{})
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping SageMaker Code Repository sweep for %s: %s", region, err)
			return sweeperErrs.ErrorOrNil()
		}

		if err != nil {
			sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("retrieving SageMaker Code Repositories: %w", err))
		}

		for _, instance := range page.CodeRepositorySummaryList {
			r := resourceCodeRepository()
			d := r.Data(nil)
			d.SetId(aws.ToString(instance.CodeRepositoryName))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("sweeping SageMaker Code Repositories: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepDeviceFleets(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("getting client: %s", err)
	}
	conn := client.SageMakerClient(ctx)

	sweepResources := make([]sweep.Sweepable, 0)
	var sweeperErrs *multierror.Error

	pages := sagemaker.NewListDeviceFleetsPaginator(conn, &sagemaker.ListDeviceFleetsInput{})
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping SageMaker Device Fleet sweep for %s: %s", region, err)
			return sweeperErrs.ErrorOrNil()
		}

		if err != nil {
			sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("retrieving SageMaker Device Fleets: %w", err))
		}

		for _, deviceFleet := range page.DeviceFleetSummaries {
			name := aws.ToString(deviceFleet.DeviceFleetName)

			r := resourceDeviceFleet()
			d := r.Data(nil)
			d.SetId(name)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("sweeping SageMaker Device Fleets: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepDomains(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("getting client: %s", err)
	}
	conn := client.SageMakerClient(ctx)

	sweepResources := make([]sweep.Sweepable, 0)
	var sweeperErrs *multierror.Error

	pages := sagemaker.NewListDomainsPaginator(conn, &sagemaker.ListDomainsInput{})
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping SageMaker domain sweep for %s: %s", region, err)
			return sweeperErrs.ErrorOrNil()
		}

		if err != nil {
			sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("retrieving SageMaker Domains: %w", err))
		}

		for _, domain := range page.Domains {
			r := resourceDomain()
			d := r.Data(nil)
			d.SetId(aws.ToString(domain.DomainId))
			d.Set("retention_policy.0.home_efs_file_system", "Delete")

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("sweeping API Gateway VPC Links: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepEndpointConfigurations(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("getting client: %w", err)
	}
	conn := client.SageMakerClient(ctx)

	sweepResources := make([]sweep.Sweepable, 0)
	var sweeperErrs *multierror.Error

	input := &sagemaker.ListEndpointConfigsInput{
		NameContains: aws.String(sweep.ResourcePrefix),
	}

	pages := sagemaker.NewListEndpointConfigsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping SageMaker Endpoint Config sweep for %s: %s", region, err)
			return sweeperErrs.ErrorOrNil()
		}
		if err != nil {
			sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("retrieving SageMaker Endpoint Configs: %w", err))
		}

		for _, endpointConfig := range page.EndpointConfigs {
			r := resourceEndpointConfiguration()
			d := r.Data(nil)
			d.SetId(aws.ToString(endpointConfig.EndpointConfigName))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("sweeping SageMaker Endpoint Configs: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepEndpoints(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("getting client: %s", err)
	}
	conn := client.SageMakerClient(ctx)

	input := &sagemaker.ListEndpointsInput{
		NameContains: aws.String(sweep.ResourcePrefix),
	}

	pages := sagemaker.NewListEndpointsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return fmt.Errorf("listing endpoints: %s", err)
		}

		if len(page.Endpoints) == 0 {
			log.Print("[DEBUG] No SageMaker Endpoint to sweep")
			return nil
		}

		for _, endpoint := range page.Endpoints {
			_, err := conn.DeleteEndpoint(ctx, &sagemaker.DeleteEndpointInput{
				EndpointName: endpoint.EndpointName,
			})
			if err != nil {
				return fmt.Errorf("deleting SageMaker Endpoint (%s): %s", aws.ToString(endpoint.EndpointName), err)
			}
		}
	}

	return nil
}

func sweepFeatureGroups(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("getting client: %s", err)
	}
	conn := client.SageMakerClient(ctx)

	sweepResources := make([]sweep.Sweepable, 0)
	var sweeperErrs *multierror.Error

	pages := sagemaker.NewListFeatureGroupsPaginator(conn, &sagemaker.ListFeatureGroupsInput{})
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping SageMaker Feature Group sweep for %s: %s", region, err)
			return sweeperErrs.ErrorOrNil()
		}

		if err != nil {
			sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("retrieving SageMaker Feature Groups: %w", err))
		}

		for _, group := range page.FeatureGroupSummaries {
			r := resourceFeatureGroup()
			d := r.Data(nil)
			d.SetId(aws.ToString(group.FeatureGroupName))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("sweeping SageMaker Feature Groups: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepFlowDefinitions(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("getting client: %w", err)
	}
	conn := client.SageMakerClient(ctx)

	sweepResources := make([]sweep.Sweepable, 0)
	var sweeperErrs *multierror.Error

	pages := sagemaker.NewListFlowDefinitionsPaginator(conn, &sagemaker.ListFlowDefinitionsInput{})
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping SageMaker Flow Definition sweep for %s: %s", region, err)
			return sweeperErrs.ErrorOrNil()
		}

		if err != nil {
			sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("retrieving SageMaker Flow Definitions: %w", err))
		}

		for _, flowDefinition := range page.FlowDefinitionSummaries {
			r := resourceFlowDefinition()
			d := r.Data(nil)
			d.SetId(aws.ToString(flowDefinition.FlowDefinitionName))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("sweeping SageMaker Flow Definitions: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepHumanTaskUIs(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("getting client: %w", err)
	}
	conn := client.SageMakerClient(ctx)

	sweepResources := make([]sweep.Sweepable, 0)
	var sweeperErrs *multierror.Error

	pages := sagemaker.NewListHumanTaskUisPaginator(conn, &sagemaker.ListHumanTaskUisInput{})
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping SageMaker HumanTaskUi sweep for %s: %s", region, err)
			return sweeperErrs.ErrorOrNil()
		}

		if err != nil {
			sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("retrieving SageMaker HumanTaskUis: %w", err))
		}

		for _, humanTaskUi := range page.HumanTaskUiSummaries {
			r := resourceHumanTaskUI()
			d := r.Data(nil)
			d.SetId(aws.ToString(humanTaskUi.HumanTaskUiName))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("sweeping SageMaker HumanTaskUis: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepImages(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("getting client: %s", err)
	}
	conn := client.SageMakerClient(ctx)

	sweepResources := make([]sweep.Sweepable, 0)
	var sweeperErrs *multierror.Error

	pages := sagemaker.NewListImagesPaginator(conn, &sagemaker.ListImagesInput{})
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping SageMaker Image sweep for %s: %s", region, err)
			return sweeperErrs.ErrorOrNil()
		}

		if err != nil {
			sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("retrieving SageMaker Images: %w", err))
		}

		for _, image := range page.Images {
			r := resourceImage()
			d := r.Data(nil)
			d.SetId(aws.ToString(image.ImageName))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("sweeping SageMaker Images: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepModelPackageGroups(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("getting client: %s", err)
	}
	conn := client.SageMakerClient(ctx)

	sweepResources := make([]sweep.Sweepable, 0)
	var sweeperErrs *multierror.Error

	pages := sagemaker.NewListModelPackageGroupsPaginator(conn, &sagemaker.ListModelPackageGroupsInput{})
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping SageMaker Model Package Group sweep for %s: %s", region, err)
			return sweeperErrs.ErrorOrNil()
		}

		if err != nil {
			sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("retrieving SageMaker Model Package Groups: %w", err))
		}

		for _, modelPackageGroup := range page.ModelPackageGroupSummaryList {
			r := resourceModelPackageGroup()
			d := r.Data(nil)
			d.SetId(aws.ToString(modelPackageGroup.ModelPackageGroupName))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("sweeping SageMaker Model Package Groups: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepModels(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("getting client: %w", err)
	}
	conn := client.SageMakerClient(ctx)

	sweepResources := make([]sweep.Sweepable, 0)
	var sweeperErrs *multierror.Error

	pages := sagemaker.NewListModelsPaginator(conn, &sagemaker.ListModelsInput{})
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping SageMaker Model sweep for %s: %s", region, err)
			return sweeperErrs.ErrorOrNil()
		}

		if err != nil {
			sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("retrieving SageMaker Models: %w", err))
		}

		for _, model := range page.Models {
			r := resourceModel()
			d := r.Data(nil)
			d.SetId(aws.ToString(model.ModelName))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("sweeping SageMaker Models: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepNotebookInstanceLifecycleConfiguration(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("getting client: %s", err)
	}
	conn := client.SageMakerClient(ctx)

	sweepResources := make([]sweep.Sweepable, 0)
	var sweeperErrs *multierror.Error

	pages := sagemaker.NewListNotebookInstanceLifecycleConfigsPaginator(conn, &sagemaker.ListNotebookInstanceLifecycleConfigsInput{})
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping SageMaker Notebook Instance Lifecycle Configuration sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("retrieving SageMaker Notebook Instance Lifecycle Configurations: %s", err))
		}

		for _, lifecycleConfig := range page.NotebookInstanceLifecycleConfigs {
			name := aws.ToString(lifecycleConfig.NotebookInstanceLifecycleConfigName)
			if !strings.HasPrefix(name, sweep.ResourcePrefix) {
				log.Printf("[INFO] Skipping SageMaker Notebook Instance Lifecycle Configuration (%s): not in allow list", name)
				continue
			}

			r := resourceNotebookInstanceLifeCycleConfiguration()
			d := r.Data(nil)
			d.SetId(aws.ToString(lifecycleConfig.NotebookInstanceLifecycleConfigName))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("sweeping SageMaker Notebook Instance Lifecycle Configurations: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepNotebookInstances(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("getting client: %s", err)
	}
	conn := client.SageMakerClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	pages := sagemaker.NewListNotebookInstancesPaginator(conn, &sagemaker.ListNotebookInstancesInput{})
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping SageMaker Notebook Instance sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing SageMaker Notebook Instances (%s): %w", region, err)
		}

		for _, v := range page.NotebookInstances {
			name := aws.ToString(v.NotebookInstanceName)

			if status := v.NotebookInstanceStatus; status == awstypes.NotebookInstanceStatusDeleting {
				log.Printf("[INFO] Skipping SageMaker Notebook Instance %s: NotebookInstanceStatus=%s", name, status)
				continue
			}

			r := resourceNotebookInstance()
			d := r.Data(nil)
			d.SetId(name)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping SageMaker Notebook Instances (%s): %w", region, err)
	}

	return nil
}

func sweepStudioLifecyclesConfig(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("getting client: %w", err)
	}
	conn := client.SageMakerClient(ctx)

	sweepResources := make([]sweep.Sweepable, 0)
	var sweeperErrs *multierror.Error

	pages := sagemaker.NewListStudioLifecycleConfigsPaginator(conn, &sagemaker.ListStudioLifecycleConfigsInput{})
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping SageMaker Studio Lifecycle Config sweep for %s: %s", region, err)
			return sweeperErrs.ErrorOrNil()
		}

		if err != nil {
			sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("retrieving SageMaker Studio Lifecycle Configs: %w", err))
		}

		for _, config := range page.StudioLifecycleConfigs {
			r := resourceStudioLifecycleConfig()
			d := r.Data(nil)
			d.SetId(aws.ToString(config.StudioLifecycleConfigName))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("sweeping SageMaker Studio Lifecycle Configs: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepUserProfiles(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("getting client: %s", err)
	}
	conn := client.SageMakerClient(ctx)

	sweepResources := make([]sweep.Sweepable, 0)
	var sweeperErrs *multierror.Error

	pages := sagemaker.NewListUserProfilesPaginator(conn, &sagemaker.ListUserProfilesInput{})
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping SageMaker domain sweep for %s: %s", region, err)
			return sweeperErrs.ErrorOrNil()
		}

		if err != nil {
			sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("retrieving SageMaker User Profiles: %w", err))
		}

		for _, userProfile := range page.UserProfiles {
			r := resourceUserProfile()
			d := r.Data(nil)
			d.SetId(aws.ToString(userProfile.UserProfileName))
			d.Set("user_profile_name", userProfile.UserProfileName)
			d.Set("domain_id", userProfile.DomainId)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("sweeping SageMaker User Profiles: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepWorkforces(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("getting client: %w", err)
	}
	conn := client.SageMakerClient(ctx)

	sweepResources := make([]sweep.Sweepable, 0)
	var sweeperErrs *multierror.Error

	pages := sagemaker.NewListWorkforcesPaginator(conn, &sagemaker.ListWorkforcesInput{})
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping SageMaker workforce sweep for %s: %s", region, err)
			return sweeperErrs.ErrorOrNil()
		}

		if err != nil {
			sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("retrieving SageMaker Workforces: %w", err))
		}

		for _, workforce := range page.Workforces {
			r := resourceWorkforce()
			d := r.Data(nil)
			d.SetId(aws.ToString(workforce.WorkforceName))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("sweeping SageMaker Workforces: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepWorkteams(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("getting client: %w", err)
	}
	conn := client.SageMakerClient(ctx)

	sweepResources := make([]sweep.Sweepable, 0)
	var sweeperErrs *multierror.Error

	pages := sagemaker.NewListWorkteamsPaginator(conn, &sagemaker.ListWorkteamsInput{})
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping SageMaker workteam sweep for %s: %s", region, err)
			return sweeperErrs.ErrorOrNil()
		}

		if err != nil {
			sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("retrieving SageMaker Workteams: %w", err))
		}

		for _, workteam := range page.Workteams {
			r := resourceWorkteam()
			d := r.Data(nil)
			d.SetId(aws.ToString(workteam.WorkteamName))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("sweeping SageMaker Workteams: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepProjects(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("getting client: %s", err)
	}
	conn := client.SageMakerClient(ctx)
	var sweeperErrs *multierror.Error

	pages := sagemaker.NewListProjectsPaginator(conn, &sagemaker.ListProjectsInput{})
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping SageMaker Project sweep for %s: %s", region, err)
			return sweeperErrs.ErrorOrNil()
		}

		if err != nil {
			sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing SageMaker Projects (%s): %w", region, err))
		}

		for _, v := range page.ProjectSummaryList {
			name := aws.ToString(v.ProjectName)

			if status := v.ProjectStatus; status == awstypes.ProjectStatusDeleteCompleted {
				log.Printf("[INFO] Skipping SageMaker Project %s: ProjectStatus=%s", name, status)
				continue
			}

			r := resourceProject()
			d := r.Data(nil)
			d.SetId(name)

			if err := sdk.NewSweepResource(r, d, client).Delete(ctx); err != nil {
				sweeperErrs = multierror.Append(sweeperErrs, err)
			}
		}
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepPipelines(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("getting client: %s", err)
	}
	conn := client.SageMakerClient(ctx)

	sweepResources := make([]sweep.Sweepable, 0)
	var sweeperErrs *multierror.Error

	pages := sagemaker.NewListPipelinesPaginator(conn, &sagemaker.ListPipelinesInput{})
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping SageMaker Pipeline sweep for %s: %s", region, err)
			return sweeperErrs.ErrorOrNil()
		}
		if err != nil {
			sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("retrieving SageMaker Pipelines: %w", err))
		}

		for _, project := range page.PipelineSummaries {
			name := aws.ToString(project.PipelineName)

			r := resourcePipeline()
			d := r.Data(nil)
			d.SetId(name)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("sweeping SageMaker Pipelines: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepMlflowTrackingServers(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("getting client: %s", err)
	}
	conn := client.SageMakerClient(ctx)

	sweepResources := make([]sweep.Sweepable, 0)
	var sweeperErrs *multierror.Error

	pages := sagemaker.NewListMlflowTrackingServersPaginator(conn, &sagemaker.ListMlflowTrackingServersInput{})
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping SageMaker Mlflow Tracking Server sweep for %s: %s", region, err)
			return sweeperErrs.ErrorOrNil()
		}
		if err != nil {
			sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("retrieving SageMaker Mlflow Tracking Servers: %w", err))
		}

		for _, project := range page.TrackingServerSummaries {
			name := aws.ToString(project.TrackingServerName)

			r := resourceMlflowTrackingServer()
			d := r.Data(nil)
			d.SetId(name)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("sweeping SageMaker Mlflow Tracking Servers: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepHubs(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("getting client: %s", err)
	}
	conn := client.SageMakerClient(ctx)

	var sweepResources []sweep.Sweepable

	in := sagemaker.ListHubsInput{}
	for {
		out, err := conn.ListHubs(ctx, &in)
		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Sagemaker Hubs sweep for %s: %s", region, err)
			return nil
		}
		// The Sagemaker API returns this in unsupported regions
		if tfawserr.ErrCodeEquals(err, "ThrottlingException") {
			tflog.Warn(ctx, "Skipping sweeper", map[string]any{
				"skip_reason": "Unsupported region",
				"error":       err.Error(),
			})
			return nil
		}
		if err != nil {
			return fmt.Errorf("error retrieving Sagemaker Hubs: %w", err)
		}

		for _, hub := range out.HubSummaries {
			name := aws.ToString(hub.HubName)
			log.Printf("[INFO] Deleting Sagemaker Hubs: %s", name)

			if !strings.HasPrefix(name, sweep.ResourcePrefix) {
				log.Printf("[INFO] Skipping SageMaker Hub (%s): not in allow list", name)
				continue
			}

			r := resourceHub()
			d := r.Data(nil)
			d.SetId(name)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		if aws.ToString(out.NextToken) == "" {
			break
		}
		in.NextToken = out.NextToken
	}

	if err := sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		return fmt.Errorf("error sweeping Sagemaker Hubs for %s: %w", region, err)
	}

	return nil
}
