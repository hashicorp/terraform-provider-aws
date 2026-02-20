// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package sagemaker

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/sagemaker"
	awstypes "github.com/aws/aws-sdk-go-v2/service/sagemaker/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
)

func statusNotebookInstance(conn *sagemaker.Client, notebookName string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findNotebookInstanceByName(ctx, conn, notebookName)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.NotebookInstanceStatus), nil
	}
}

func statusModelPackageGroup(conn *sagemaker.Client, name string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findModelPackageGroupByName(ctx, conn, name)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.ModelPackageGroupStatus), nil
	}
}

func statusImage(conn *sagemaker.Client, name string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findImageByName(ctx, conn, name)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.ImageStatus), nil
	}
}

func statusImageVersionByTwoPartKey(conn *sagemaker.Client, name string, version int32) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findImageVersionByTwoPartKey(ctx, conn, name, version)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.ImageVersionStatus), nil
	}
}

func statusImageVersionByID(conn *sagemaker.Client, name string, version int32) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findImageVersionByTwoPartKey(ctx, conn, name, version)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.ImageVersionStatus), nil
	}
}

func statusDomain(conn *sagemaker.Client, domainID string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findDomainByName(ctx, conn, domainID)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func statusFeatureGroup(conn *sagemaker.Client, name string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findFeatureGroupByName(ctx, conn, name)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.FeatureGroupStatus), nil
	}
}

func statusFeatureGroupUpdate(conn *sagemaker.Client, name string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findFeatureGroupByName(ctx, conn, name)

		if retry.NotFound(err) {
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

func statusFlowDefinition(conn *sagemaker.Client, name string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findFlowDefinitionByName(ctx, conn, name)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.FlowDefinitionStatus), nil
	}
}

func statusApp(conn *sagemaker.Client, domainID, userProfileOrSpaceName, appType, appName string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findAppByName(ctx, conn, domainID, userProfileOrSpaceName, appType, appName)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func statusProject(conn *sagemaker.Client, name string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findProjectByName(ctx, conn, name)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.ProjectStatus), nil
	}
}

func statusWorkforce(conn *sagemaker.Client, name string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findWorkforceByName(ctx, conn, name)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func statusSpace(conn *sagemaker.Client, domainId, name string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findSpaceByName(ctx, conn, domainId, name)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func statusMlflowTrackingServer(conn *sagemaker.Client, name string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findMlflowTrackingServerByName(ctx, conn, name)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.TrackingServerStatus), nil
	}
}

func statusHub(conn *sagemaker.Client, name string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findHubByName(ctx, conn, name)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.HubStatus), nil
	}
}
