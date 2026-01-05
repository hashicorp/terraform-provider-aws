// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package sagemaker

import (
	"context"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sagemaker"
	awstypes "github.com/aws/aws-sdk-go-v2/service/sagemaker/types"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
)

func RegisterSweepers() {
	awsv2.Register("aws_sagemaker_app_image_config", sweepAppImagesConfig)
	awsv2.Register("aws_sagemaker_app", sweepApps)
	awsv2.Register("aws_sagemaker_code_repository", sweepCodeRepositories)
	awsv2.Register("aws_sagemaker_domain", sweepDomains,
		"aws_efs_mount_target",
		"aws_efs_file_system",
		"aws_sagemaker_user_profile",
		"aws_sagemaker_space",
	)
	awsv2.Register("aws_sagemaker_endpoint_configuration", sweepEndpointConfigurations, "aws_sagemaker_model")
	awsv2.Register("aws_sagemaker_endpoint", sweepEndpoints,
		"aws_sagemaker_model",
		"aws_sagemaker_endpoint_configuration",
	)
	awsv2.Register("aws_sagemaker_feature_group", sweepFeatureGroups)
	awsv2.Register("aws_sagemaker_flow_definition", sweepFlowDefinitions)
	awsv2.Register("aws_sagemaker_human_task_ui", sweepHumanTaskUIs)
	awsv2.Register("aws_sagemaker_image", sweepImages)
	awsv2.Register("aws_sagemaker_mlflow_tracking_server", sweepMlflowTrackingServers)
	awsv2.Register("aws_sagemaker_model_package_group", sweepModelPackageGroups)
	awsv2.Register("aws_sagemaker_model", sweepModels)
	awsv2.Register("aws_sagemaker_notebook_instance_lifecycle_configuration", sweepNotebookInstanceLifecycleConfiguration, "aws_sagemaker_notebook_instance")
	awsv2.Register("aws_sagemaker_notebook_instance", sweepNotebookInstances)
	awsv2.Register("aws_sagemaker_studio_lifecycle_config", sweepStudioLifecyclesConfig, "aws_sagemaker_domain")
	awsv2.Register("aws_sagemaker_project", sweepProjects)
	awsv2.Register("aws_sagemaker_space", sweepSpaces, "aws_sagemaker_app")
	awsv2.Register("aws_sagemaker_user_profile", sweepUserProfiles, "aws_sagemaker_app")
	awsv2.Register("aws_sagemaker_workforce", sweepWorkforces, "aws_sagemaker_workteam")
	awsv2.Register("aws_sagemaker_workteam", sweepWorkteams)
	awsv2.Register("aws_sagemaker_pipeline", sweepPipelines)
	awsv2.Register("aws_sagemaker_hub", sweepHubs)
}

func sweepAppImagesConfig(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.SageMakerClient(ctx)
	var input sagemaker.ListAppImageConfigsInput
	sweepResources := make([]sweep.Sweepable, 0)

	pages := sagemaker.NewListAppImageConfigsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.AppImageConfigs {
			r := resourceAppImageConfig()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.AppImageConfigName))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, nil
}

func sweepSpaces(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.SageMakerClient(ctx)
	var input sagemaker.ListSpacesInput
	sweepResources := make([]sweep.Sweepable, 0)

	pages := sagemaker.NewListSpacesPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.Spaces {
			r := resourceSpace()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.SpaceName))
			d.Set("domain_id", v.DomainId)
			d.Set("space_name", v.SpaceName)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, nil
}

func sweepApps(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.SageMakerClient(ctx)
	var input sagemaker.ListAppsInput
	sweepResources := make([]sweep.Sweepable, 0)

	pages := sagemaker.NewListAppsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.Apps {
			if v.Status == awstypes.AppStatusDeleted {
				continue
			}

			r := resourceApp()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.AppName))
			d.Set("app_name", v.AppName)
			d.Set("app_type", v.AppType)
			d.Set("domain_id", v.DomainId)
			d.Set("user_profile_name", v.UserProfileName)
			d.Set("space_name", v.SpaceName)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, nil
}

func sweepCodeRepositories(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.SageMakerClient(ctx)
	var input sagemaker.ListCodeRepositoriesInput
	sweepResources := make([]sweep.Sweepable, 0)

	pages := sagemaker.NewListCodeRepositoriesPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.CodeRepositorySummaryList {
			r := resourceCodeRepository()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.CodeRepositoryName))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, nil
}

func sweepDomains(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.SageMakerClient(ctx)
	var input sagemaker.ListDomainsInput
	sweepResources := make([]sweep.Sweepable, 0)

	pages := sagemaker.NewListDomainsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.Domains {
			r := resourceDomain()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.DomainId))
			d.Set("retention_policy.0.home_efs_file_system", "Delete")

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, nil
}

func sweepEndpointConfigurations(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.SageMakerClient(ctx)
	input := sagemaker.ListEndpointConfigsInput{
		NameContains: aws.String(sweep.ResourcePrefix),
	}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := sagemaker.NewListEndpointConfigsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.EndpointConfigs {
			r := resourceEndpointConfiguration()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.EndpointConfigName))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, nil
}

func sweepEndpoints(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.SageMakerClient(ctx)
	input := sagemaker.ListEndpointsInput{
		NameContains: aws.String(sweep.ResourcePrefix),
	}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := sagemaker.NewListEndpointsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.Endpoints {
			r := resourceEndpoint()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.EndpointName))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, nil
}

func sweepFeatureGroups(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.SageMakerClient(ctx)
	var input sagemaker.ListFeatureGroupsInput
	sweepResources := make([]sweep.Sweepable, 0)

	pages := sagemaker.NewListFeatureGroupsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.FeatureGroupSummaries {
			r := resourceFeatureGroup()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.FeatureGroupName))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, nil
}

func sweepFlowDefinitions(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.SageMakerClient(ctx)
	var input sagemaker.ListFlowDefinitionsInput
	sweepResources := make([]sweep.Sweepable, 0)

	pages := sagemaker.NewListFlowDefinitionsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.FlowDefinitionSummaries {
			r := resourceFlowDefinition()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.FlowDefinitionName))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, nil
}

func sweepHumanTaskUIs(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.SageMakerClient(ctx)
	var input sagemaker.ListHumanTaskUisInput
	sweepResources := make([]sweep.Sweepable, 0)

	pages := sagemaker.NewListHumanTaskUisPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.HumanTaskUiSummaries {
			r := resourceHumanTaskUI()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.HumanTaskUiName))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, nil
}

func sweepImages(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.SageMakerClient(ctx)
	var input sagemaker.ListImagesInput
	sweepResources := make([]sweep.Sweepable, 0)

	pages := sagemaker.NewListImagesPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.Images {
			r := resourceImage()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.ImageName))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, nil
}

func sweepModelPackageGroups(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.SageMakerClient(ctx)
	var input sagemaker.ListModelPackageGroupsInput
	sweepResources := make([]sweep.Sweepable, 0)

	pages := sagemaker.NewListModelPackageGroupsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.ModelPackageGroupSummaryList {
			r := resourceModelPackageGroup()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.ModelPackageGroupName))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, nil
}

func sweepModels(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.SageMakerClient(ctx)
	var input sagemaker.ListModelsInput
	sweepResources := make([]sweep.Sweepable, 0)

	pages := sagemaker.NewListModelsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.Models {
			r := resourceModel()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.ModelName))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, nil
}

func sweepNotebookInstanceLifecycleConfiguration(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.SageMakerClient(ctx)
	var input sagemaker.ListNotebookInstanceLifecycleConfigsInput
	sweepResources := make([]sweep.Sweepable, 0)

	pages := sagemaker.NewListNotebookInstanceLifecycleConfigsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.NotebookInstanceLifecycleConfigs {
			name := aws.ToString(v.NotebookInstanceLifecycleConfigName)

			if !strings.HasPrefix(name, sweep.ResourcePrefix) {
				log.Printf("[INFO] Skipping SageMaker AI Notebook Instance Lifecycle Configuration %s", name)
				continue
			}

			r := resourceNotebookInstanceLifeCycleConfiguration()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.NotebookInstanceLifecycleConfigName))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, nil
}

func sweepNotebookInstances(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.SageMakerClient(ctx)
	var input sagemaker.ListNotebookInstancesInput
	sweepResources := make([]sweep.Sweepable, 0)

	pages := sagemaker.NewListNotebookInstancesPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.NotebookInstances {
			name := aws.ToString(v.NotebookInstanceName)

			if status := v.NotebookInstanceStatus; status == awstypes.NotebookInstanceStatusDeleting {
				log.Printf("[INFO] Skipping SageMaker AI Notebook Instance %s: NotebookInstanceStatus=%s", name, status)
				continue
			}

			r := resourceNotebookInstance()
			d := r.Data(nil)
			d.SetId(name)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, nil
}

func sweepStudioLifecyclesConfig(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.SageMakerClient(ctx)
	var input sagemaker.ListStudioLifecycleConfigsInput
	sweepResources := make([]sweep.Sweepable, 0)

	pages := sagemaker.NewListStudioLifecycleConfigsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.StudioLifecycleConfigs {
			r := resourceStudioLifecycleConfig()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.StudioLifecycleConfigName))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, nil
}

func sweepUserProfiles(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.SageMakerClient(ctx)

	var sweepResources []sweep.Sweepable
	r := resourceUserProfile()

	var input sagemaker.ListUserProfilesInput
	pages := sagemaker.NewListUserProfilesPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.UserProfiles {
			describeInput := sagemaker.DescribeUserProfileInput{
				DomainId:        v.DomainId,
				UserProfileName: v.UserProfileName,
			}
			userProfile, err := conn.DescribeUserProfile(ctx, &describeInput)
			if err != nil {
				return nil, err
			}

			d := r.Data(nil)
			d.SetId(aws.ToString(userProfile.UserProfileArn))
			d.Set("user_profile_name", userProfile.UserProfileName)
			d.Set("domain_id", userProfile.DomainId)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, nil
}

func sweepWorkforces(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.SageMakerClient(ctx)
	var input sagemaker.ListWorkforcesInput
	sweepResources := make([]sweep.Sweepable, 0)

	pages := sagemaker.NewListWorkforcesPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.Workforces {
			r := resourceWorkforce()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.WorkforceName))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, nil
}

func sweepWorkteams(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.SageMakerClient(ctx)
	var input sagemaker.ListWorkteamsInput
	sweepResources := make([]sweep.Sweepable, 0)

	pages := sagemaker.NewListWorkteamsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.Workteams {
			r := resourceWorkteam()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.WorkteamName))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, nil
}

func sweepProjects(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.SageMakerClient(ctx)
	var input sagemaker.ListProjectsInput
	sweepResources := make([]sweep.Sweepable, 0)

	pages := sagemaker.NewListProjectsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.ProjectSummaryList {
			name := aws.ToString(v.ProjectName)

			if status := v.ProjectStatus; status == awstypes.ProjectStatusDeleteCompleted {
				log.Printf("[INFO] Skipping SageMaker AI Project %s: ProjectStatus=%s", name, status)
				continue
			}

			r := resourceProject()
			d := r.Data(nil)
			d.SetId(name)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, nil
}

func sweepPipelines(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.SageMakerClient(ctx)
	var input sagemaker.ListPipelinesInput
	sweepResources := make([]sweep.Sweepable, 0)

	pages := sagemaker.NewListPipelinesPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.PipelineSummaries {
			r := resourcePipeline()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.PipelineName))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, nil
}

func sweepMlflowTrackingServers(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.SageMakerClient(ctx)
	var input sagemaker.ListMlflowTrackingServersInput
	sweepResources := make([]sweep.Sweepable, 0)

	pages := sagemaker.NewListMlflowTrackingServersPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.TrackingServerSummaries {
			r := resourceMlflowTrackingServer()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.TrackingServerName))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, nil
}

func sweepHubs(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.SageMakerClient(ctx)
	var input sagemaker.ListHubsInput
	sweepResources := make([]sweep.Sweepable, 0)

	err := listHubsPages(ctx, conn, &input, func(page *sagemaker.ListHubsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.HubSummaries {
			name := aws.ToString(v.HubName)

			if !strings.HasPrefix(name, sweep.ResourcePrefix) {
				log.Printf("[INFO] Skipping SageMaker AI Hub %s", name)
				continue
			}

			r := resourceHub()
			d := r.Data(nil)
			d.SetId(name)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	return sweepResources, nil
}
