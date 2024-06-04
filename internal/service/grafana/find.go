// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package grafana

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/grafana"
	awstypes "github.com/aws/aws-sdk-go-v2/service/grafana/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func FindLicensedWorkspaceByID(ctx context.Context, conn *grafana.Client, id string) (*awstypes.WorkspaceDescription, error) {
	output, err := FindWorkspaceByID(ctx, conn, id)

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindWorkspaceByID(ctx context.Context, conn *grafana.Client, id string) (*awstypes.WorkspaceDescription, error) {
	input := &grafana.DescribeWorkspaceInput{
		WorkspaceId: aws.String(id),
	}

	output, err := conn.DescribeWorkspace(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Workspace == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Workspace, nil
}

func FindSamlConfigurationByID(ctx context.Context, conn *grafana.Client, id string) (*awstypes.SamlAuthentication, error) {
	input := &grafana.DescribeWorkspaceAuthenticationInput{
		WorkspaceId: aws.String(id),
	}

	output, err := conn.DescribeWorkspaceAuthentication(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Authentication == nil || output.Authentication.Saml == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if status := output.Authentication.Saml.Status; status == awstypes.SamlConfigurationStatusNotConfigured {
		return nil, &retry.NotFoundError{
			Message:     string(status),
			LastRequest: input,
		}
	}

	return output.Authentication.Saml, nil
}

func FindRoleAssociationsByRoleAndWorkspaceID(ctx context.Context, conn *grafana.Client, role string, workspaceID string) (map[string][]string, error) {
	input := &grafana.ListPermissionsInput{
		WorkspaceId: aws.String(workspaceID),
	}
	output := make(map[string][]string, 0)

	pages := grafana.NewListPermissionsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return output, err
		}

		for _, v := range page.Permissions {
			if string(v.Role) == role {
				userType := string(v.User.Type)
				output[userType] = append(output[userType], aws.ToString(v.User.Id))
			}
		}
	}

	if len(output) == 0 {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}
