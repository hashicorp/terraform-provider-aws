//go:build sweep
// +build sweep

package sagemaker

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sagemaker"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
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

	// resource.AddTestSweepers("aws_sagemaker_device", &resource.Sweeper{
	// 	Name: "aws_sagemaker_device",
	// 	F:    sweepDevices,
	// })

	resource.AddTestSweepers("aws_sagemaker_domain", &resource.Sweeper{
		Name: "aws_sagemaker_domain",
		F:    sweepDomains,
		Dependencies: []string{
			"aws_efs_mount_target",
			"aws_efs_file_system",
			"aws_sagemaker_user_profile",
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

	resource.AddTestSweepers("aws_sagemaker_project", &resource.Sweeper{
		Name: "aws_sagemaker_project",
		F:    sweepProjects,
	})
}

func sweepAppImagesConfig(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("getting client: %w", err)
	}

	conn := client.(*conns.AWSClient).SageMakerConn
	input := &sagemaker.ListAppImageConfigsInput{}
	var sweeperErrs *multierror.Error

	for {
		output, err := conn.ListAppImageConfigs(input)

		if sweep.SkipSweepError(err) {
			log.Printf("[WARN] Skipping SageMaker App Image Config for %s: %s", region, err)
			return sweeperErrs.ErrorOrNil()
		}

		if err != nil {
			sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("retrieving Example Thing: %w", err))
			return sweeperErrs
		}

		for _, config := range output.AppImageConfigs {

			name := aws.StringValue(config.AppImageConfigName)
			r := ResourceAppImageConfig()
			d := r.Data(nil)
			d.SetId(name)
			err = r.Delete(d, client)
			if err != nil {
				sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("deleting SageMaker App Image Config (%s): %w", name, err))
				continue
			}
		}

		if aws.StringValue(output.NextToken) == "" {
			break
		}

		input.NextToken = output.NextToken
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepApps(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("getting client: %w", err)
	}
	conn := client.(*conns.AWSClient).SageMakerConn
	var sweeperErrs *multierror.Error

	err = conn.ListAppsPages(&sagemaker.ListAppsInput{}, func(page *sagemaker.ListAppsOutput, lastPage bool) bool {
		for _, app := range page.Apps {

			if aws.StringValue(app.Status) == sagemaker.AppStatusDeleted {
				continue
			}

			r := ResourceApp()
			d := r.Data(nil)
			d.SetId(aws.StringValue(app.AppName))
			d.Set("app_name", app.AppName)
			d.Set("app_type", app.AppType)
			d.Set("domain_id", app.DomainId)
			d.Set("user_profile_name", app.UserProfileName)

			err := r.Delete(d, client)
			if err != nil {
				sweeperErrs = multierror.Append(sweeperErrs, err)
				continue
			}
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping SageMaker domain sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil()
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("retrieving SageMaker Apps: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepCodeRepositories(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).SageMakerConn
	var sweeperErrs *multierror.Error

	err = conn.ListCodeRepositoriesPages(&sagemaker.ListCodeRepositoriesInput{}, func(page *sagemaker.ListCodeRepositoriesOutput, lastPage bool) bool {
		for _, instance := range page.CodeRepositorySummaryList {
			r := ResourceCodeRepository()
			d := r.Data(nil)
			d.SetId(aws.StringValue(instance.CodeRepositoryName))

			err := r.Delete(d, client)
			if err != nil {
				sweeperErrs = multierror.Append(sweeperErrs, err)
				continue
			}
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping SageMaker Code Repository sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil()
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("retrieving SageMaker Code Repositories: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepDeviceFleets(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).SageMakerConn
	var sweeperErrs *multierror.Error

	err = conn.ListDeviceFleetsPages(&sagemaker.ListDeviceFleetsInput{}, func(page *sagemaker.ListDeviceFleetsOutput, lastPage bool) bool {
		for _, deviceFleet := range page.DeviceFleetSummaries {
			name := aws.StringValue(deviceFleet.DeviceFleetName)

			r := ResourceDeviceFleet()
			d := r.Data(nil)
			d.SetId(name)

			err := r.Delete(d, client)
			if err != nil {
				sweeperErrs = multierror.Append(sweeperErrs, err)
				continue
			}
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping SageMaker Device Fleet sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil()
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("retrieving SageMaker Device Fleets: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

// func sweepDevices(region string) error {
// 	client, err := sweep.SharedRegionalSweepClient(region)
// 	if err != nil {
// 		return fmt.Errorf("getting client: %s", err)
// 	}
// 	conn := client.(*conns.AWSClient).SageMakerConn
// 	var sweeperErrs *multierror.Error

// 	err = conn.ListDevicesPages(&sagemaker.ListDevicesInput{}, func(page *sagemaker.ListDevicesOutput, lastPage bool) bool {
// 		for _, deviceFleet := range page.DeviceFleetSummaries {
// 			name := aws.StringValue(deviceFleet.DeviceFleetName)

// 			r := ResourceDeviceFleet()
// 			d := r.Data(nil)
// 			d.SetId(name)

// 			err := r.Delete(d, client)
// 			if err != nil {
// 				sweeperErrs = multierror.Append(sweeperErrs, err)
// 				continue
// 			}
// 		}

// 		return !lastPage
// 	})

// 	if sweep.SkipSweepError(err) {
// 		log.Printf("[WARN] Skipping SageMaker Device Fleet sweep for %s: %s", region, err)
// 		return sweeperErrs.ErrorOrNil()
// 	}

// 	if err != nil {
// 		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("retrieving SageMaker Device Fleets: %w", err))
// 	}

// 	return sweeperErrs.ErrorOrNil()
// }

func sweepDomains(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).SageMakerConn
	var sweeperErrs *multierror.Error

	err = conn.ListDomainsPages(&sagemaker.ListDomainsInput{}, func(page *sagemaker.ListDomainsOutput, lastPage bool) bool {
		for _, domain := range page.Domains {

			r := ResourceDomain()
			d := r.Data(nil)
			d.SetId(aws.StringValue(domain.DomainId))
			d.Set("retention_policy.0.home_efs_file_system", "Delete")

			err = r.Delete(d, client)
			if err != nil {
				sweeperErrs = multierror.Append(sweeperErrs, err)
				continue
			}
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping SageMaker domain sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil()
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("retrieving SageMaker Domains: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepEndpointConfigurations(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("getting client: %w", err)
	}
	conn := client.(*conns.AWSClient).SageMakerConn
	var sweeperErrs *multierror.Error

	req := &sagemaker.ListEndpointConfigsInput{
		NameContains: aws.String(sweep.ResourcePrefix),
	}
	err = conn.ListEndpointConfigsPages(req, func(page *sagemaker.ListEndpointConfigsOutput, lastPage bool) bool {
		for _, endpointConfig := range page.EndpointConfigs {
			r := ResourceEndpointConfiguration()
			d := r.Data(nil)
			d.SetId(aws.StringValue(endpointConfig.EndpointConfigName))

			err := r.Delete(d, client)
			if err != nil {
				sweeperErrs = multierror.Append(sweeperErrs, err)
				continue
			}
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping SageMaker Endpoint Config sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil()
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("retrieving SageMaker Endpoint Configs: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepEndpoints(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).SageMakerConn

	req := &sagemaker.ListEndpointsInput{
		NameContains: aws.String(sweep.ResourcePrefix),
	}
	resp, err := conn.ListEndpoints(req)
	if err != nil {
		return fmt.Errorf("listing endpoints: %s", err)
	}

	if len(resp.Endpoints) == 0 {
		log.Print("[DEBUG] No SageMaker Endpoint to sweep")
		return nil
	}

	for _, endpoint := range resp.Endpoints {
		_, err := conn.DeleteEndpoint(&sagemaker.DeleteEndpointInput{
			EndpointName: endpoint.EndpointName,
		})
		if err != nil {
			return fmt.Errorf("deleting SageMaker Endpoint (%s): %s", aws.StringValue(endpoint.EndpointName), err)
		}
	}

	return nil
}

func sweepFeatureGroups(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).SageMakerConn
	var sweeperErrs *multierror.Error

	err = conn.ListFeatureGroupsPages(&sagemaker.ListFeatureGroupsInput{}, func(page *sagemaker.ListFeatureGroupsOutput, lastPage bool) bool {
		for _, group := range page.FeatureGroupSummaries {
			r := ResourceFeatureGroup()
			d := r.Data(nil)
			d.SetId(aws.StringValue(group.FeatureGroupName))

			err := r.Delete(d, client)
			if err != nil {
				sweeperErrs = multierror.Append(sweeperErrs, err)
				continue
			}
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping SageMaker Feature Group sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil()
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("retrieving SageMaker Feature Groups: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepFlowDefinitions(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("getting client: %w", err)
	}
	conn := client.(*conns.AWSClient).SageMakerConn
	var sweeperErrs *multierror.Error

	err = conn.ListFlowDefinitionsPages(&sagemaker.ListFlowDefinitionsInput{}, func(page *sagemaker.ListFlowDefinitionsOutput, lastPage bool) bool {
		for _, flowDefinition := range page.FlowDefinitionSummaries {

			r := ResourceFlowDefinition()
			d := r.Data(nil)
			d.SetId(aws.StringValue(flowDefinition.FlowDefinitionName))

			err := r.Delete(d, client)
			if err != nil {
				sweeperErrs = multierror.Append(sweeperErrs, err)
				continue
			}
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping SageMaker Flow Definition sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil()
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("retrieving SageMaker Flow Definitions: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepHumanTaskUIs(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("getting client: %w", err)
	}
	conn := client.(*conns.AWSClient).SageMakerConn
	var sweeperErrs *multierror.Error

	err = conn.ListHumanTaskUisPages(&sagemaker.ListHumanTaskUisInput{}, func(page *sagemaker.ListHumanTaskUisOutput, lastPage bool) bool {
		for _, humanTaskUi := range page.HumanTaskUiSummaries {

			r := ResourceHumanTaskUI()
			d := r.Data(nil)
			d.SetId(aws.StringValue(humanTaskUi.HumanTaskUiName))

			err := r.Delete(d, client)
			if err != nil {
				sweeperErrs = multierror.Append(sweeperErrs, err)
				continue
			}
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping SageMaker humanTaskUi sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil()
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("retrieving SageMaker HumanTaskUis: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepImages(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).SageMakerConn
	var sweeperErrs *multierror.Error

	err = conn.ListImagesPages(&sagemaker.ListImagesInput{}, func(page *sagemaker.ListImagesOutput, lastPage bool) bool {
		for _, image := range page.Images {
			r := ResourceImage()
			d := r.Data(nil)
			d.SetId(aws.StringValue(image.ImageName))

			err := r.Delete(d, client)
			if err != nil {
				sweeperErrs = multierror.Append(sweeperErrs, err)
				continue
			}
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping SageMaker Image sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil()
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("retrieving SageMaker Images: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepModelPackageGroups(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).SageMakerConn
	var sweeperErrs *multierror.Error

	err = conn.ListModelPackageGroupsPages(&sagemaker.ListModelPackageGroupsInput{}, func(page *sagemaker.ListModelPackageGroupsOutput, lastPage bool) bool {
		for _, modelPackageGroup := range page.ModelPackageGroupSummaryList {
			r := ResourceModelPackageGroup()
			d := r.Data(nil)
			d.SetId(aws.StringValue(modelPackageGroup.ModelPackageGroupName))

			err := r.Delete(d, client)
			if err != nil {
				sweeperErrs = multierror.Append(sweeperErrs, err)
				continue
			}
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping SageMaker Model Package Group sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil()
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("retrieving SageMaker Model Package Groups: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepModels(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("getting client: %w", err)
	}
	conn := client.(*conns.AWSClient).SageMakerConn
	var sweeperErrs *multierror.Error

	err = conn.ListModelsPages(&sagemaker.ListModelsInput{}, func(page *sagemaker.ListModelsOutput, lastPage bool) bool {
		for _, model := range page.Models {

			r := ResourceModel()
			d := r.Data(nil)
			d.SetId(aws.StringValue(model.ModelName))

			err = r.Delete(d, client)
			if err != nil {
				sweeperErrs = multierror.Append(sweeperErrs, err)
				continue
			}
		}

		return !lastPage
	})
	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping SageMaker Model sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil()
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("retrieving SageMaker Models: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepNotebookInstanceLifecycleConfiguration(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).SageMakerConn
	var sweeperErrs *multierror.Error

	input := &sagemaker.ListNotebookInstanceLifecycleConfigsInput{}
	err = conn.ListNotebookInstanceLifecycleConfigsPages(input, func(page *sagemaker.ListNotebookInstanceLifecycleConfigsOutput, lastPage bool) bool {
		for _, lifecycleConfig := range page.NotebookInstanceLifecycleConfigs {
			name := aws.StringValue(lifecycleConfig.NotebookInstanceLifecycleConfigName)
			if !strings.HasPrefix(name, sweep.ResourcePrefix) {
				log.Printf("[INFO] Skipping SageMaker Notebook Instance Lifecycle Configuration (%s): not in allow list", name)
				continue
			}

			r := ResourceNotebookInstanceLifeCycleConfiguration()
			d := r.Data(nil)
			d.SetId(aws.StringValue(lifecycleConfig.NotebookInstanceLifecycleConfigName))

			err := r.Delete(d, client)
			if err != nil {
				sweeperErrs = multierror.Append(sweeperErrs, err)
				continue
			}
		}
		return !lastPage
	})
	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping SageMaker Notebook Instance Lifecycle Configuration sweep for %s: %s", region, err)
		return nil
	}
	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("retrieving SageMaker Notebook Instance Lifecycle Configuration: %s", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepNotebookInstances(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).SageMakerConn
	var sweeperErrs *multierror.Error

	err = conn.ListNotebookInstancesPages(&sagemaker.ListNotebookInstancesInput{}, func(page *sagemaker.ListNotebookInstancesOutput, lastPage bool) bool {
		for _, instance := range page.NotebookInstances {
			name := aws.StringValue(instance.NotebookInstanceName)

			r := ResourceNotebookInstance()
			d := r.Data(nil)
			d.SetId(name)

			err = r.Delete(d, client)
			if err != nil {
				sweeperErrs = multierror.Append(sweeperErrs, err)
				continue
			}
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping SageMaker Notebook Instance sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil()
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("retrieving SageMaker Notbook Instances: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepStudioLifecyclesConfig(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("getting client: %w", err)
	}
	conn := client.(*conns.AWSClient).SageMakerConn
	var sweeperErrs *multierror.Error

	err = conn.ListStudioLifecycleConfigsPages(&sagemaker.ListStudioLifecycleConfigsInput{}, func(page *sagemaker.ListStudioLifecycleConfigsOutput, lastPage bool) bool {
		for _, config := range page.StudioLifecycleConfigs {

			r := ResourceStudioLifecycleConfig()
			d := r.Data(nil)
			d.SetId(aws.StringValue(config.StudioLifecycleConfigName))

			err := r.Delete(d, client)
			if err != nil {
				sweeperErrs = multierror.Append(sweeperErrs, err)
				continue
			}
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping SageMaker Studio Lifecycle Config sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil()
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("retrieving SageMaker Studio Lifecycle Configs: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepUserProfiles(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).SageMakerConn
	var sweeperErrs *multierror.Error

	err = conn.ListUserProfilesPages(&sagemaker.ListUserProfilesInput{}, func(page *sagemaker.ListUserProfilesOutput, lastPage bool) bool {
		for _, userProfile := range page.UserProfiles {

			r := ResourceUserProfile()
			d := r.Data(nil)
			d.SetId(aws.StringValue(userProfile.UserProfileName))
			d.Set("user_profile_name", userProfile.UserProfileName)
			d.Set("domain_id", userProfile.DomainId)

			err := r.Delete(d, client)
			if err != nil {
				sweeperErrs = multierror.Append(sweeperErrs, err)
				continue
			}
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping SageMaker domain sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil()
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("retrieving SageMaker User Profiles: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepWorkforces(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("getting client: %w", err)
	}
	conn := client.(*conns.AWSClient).SageMakerConn
	var sweeperErrs *multierror.Error

	err = conn.ListWorkforcesPages(&sagemaker.ListWorkforcesInput{}, func(page *sagemaker.ListWorkforcesOutput, lastPage bool) bool {
		for _, workforce := range page.Workforces {

			r := ResourceWorkforce()
			d := r.Data(nil)
			d.SetId(aws.StringValue(workforce.WorkforceName))

			err := r.Delete(d, client)
			if err != nil {
				sweeperErrs = multierror.Append(sweeperErrs, err)
				continue
			}
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping SageMaker workforce sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil()
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("retrieving SageMaker Workforces: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepWorkteams(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("getting client: %w", err)
	}
	conn := client.(*conns.AWSClient).SageMakerConn
	var sweeperErrs *multierror.Error

	err = conn.ListWorkteamsPages(&sagemaker.ListWorkteamsInput{}, func(page *sagemaker.ListWorkteamsOutput, lastPage bool) bool {
		for _, workteam := range page.Workteams {

			r := ResourceWorkteam()
			d := r.Data(nil)
			d.SetId(aws.StringValue(workteam.WorkteamName))

			err := r.Delete(d, client)
			if err != nil {
				sweeperErrs = multierror.Append(sweeperErrs, err)
				continue
			}
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping SageMaker workteam sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil()
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("retrieving SageMaker Workteams: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepProjects(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).SageMakerConn
	var sweeperErrs *multierror.Error

	err = conn.ListProjectsPages(&sagemaker.ListProjectsInput{}, func(page *sagemaker.ListProjectsOutput, lastPage bool) bool {
		for _, project := range page.ProjectSummaryList {
			name := aws.StringValue(project.ProjectName)

			r := ResourceProject()
			d := r.Data(nil)
			d.SetId(name)

			err := r.Delete(d, client)
			if err != nil {
				sweeperErrs = multierror.Append(sweeperErrs, err)
				continue
			}
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping SageMaker Project sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil()
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("retrieving SageMaker Projects: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}
