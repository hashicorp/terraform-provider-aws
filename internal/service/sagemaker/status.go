// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sagemaker

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/sagemaker"
	awstypes "github.com/aws/aws-sdk-go-v2/service/sagemaker/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func statusNotebookInstance(ctx context.Context, conn *sagemaker.Client, notebookName string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findNotebookInstanceByName(ctx, conn, notebookName)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.NotebookInstanceStatus), nil
	}
}

func statusModelPackageGroup(ctx context.Context, conn *sagemaker.Client, name string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findModelPackageGroupByName(ctx, conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.ModelPackageGroupStatus), nil
	}
}

func statusImage(ctx context.Context, conn *sagemaker.Client, name string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findImageByName(ctx, conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.ImageStatus), nil
	}
}

func statusImageVersion(ctx context.Context, conn *sagemaker.Client, name string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findImageVersionByName(ctx, conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.ImageVersionStatus), nil
	}
}

func statusDomain(ctx context.Context, conn *sagemaker.Client, domainID string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findDomainByName(ctx, conn, domainID)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func statusFeatureGroup(ctx context.Context, conn *sagemaker.Client, name string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findFeatureGroupByName(ctx, conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.FeatureGroupStatus), nil
	}
}

func statusFeatureGroupUpdate(ctx context.Context, conn *sagemaker.Client, name string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findFeatureGroupByName(ctx, conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		if output.LastUpdateStatus == nil {
			return output, string(awstypes.LastUpdateStatusValueSuccessful), nil
		}

		return output, string(output.LastUpdateStatus.Status), nil
	}
}

func statusFlowDefinition(ctx context.Context, conn *sagemaker.Client, name string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findFlowDefinitionByName(ctx, conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.FlowDefinitionStatus), nil
	}
}

func statusApp(ctx context.Context, conn *sagemaker.Client, domainID, userProfileOrSpaceName, appType, appName string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findAppByName(ctx, conn, domainID, userProfileOrSpaceName, appType, appName)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func statusProject(ctx context.Context, conn *sagemaker.Client, name string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findProjectByName(ctx, conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.ProjectStatus), nil
	}
}

func statusWorkforce(ctx context.Context, conn *sagemaker.Client, name string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findWorkforceByName(ctx, conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func statusSpace(ctx context.Context, conn *sagemaker.Client, domainId, name string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findSpaceByName(ctx, conn, domainId, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func statusMonitoringSchedule(ctx context.Context, conn *sagemaker.Client, name string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findMonitoringScheduleByName(ctx, conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.MonitoringScheduleStatus), nil
	}
}

func statusMlflowTrackingServer(ctx context.Context, conn *sagemaker.Client, name string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findMlflowTrackingServerByName(ctx, conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.TrackingServerStatus), nil
	}
}

func statusHub(ctx context.Context, conn *sagemaker.Client, name string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findHubByName(ctx, conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.HubStatus), nil
	}
}
