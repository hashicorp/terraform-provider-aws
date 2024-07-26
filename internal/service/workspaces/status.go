// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package workspaces

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/workspaces"
	"github.com/aws/aws-sdk-go-v2/service/workspaces/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func StatusDirectoryState(ctx context.Context, conn *workspaces.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindDirectoryByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.State), nil
	}
}

// nosemgrep:ci.workspaces-in-func-name
func StatusWorkspaceState(ctx context.Context, conn *workspaces.Client, workspaceID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := conn.DescribeWorkspaces(ctx, &workspaces.DescribeWorkspacesInput{
			WorkspaceIds: []string{workspaceID},
		})
		if err != nil {
			return nil, string(types.WorkspaceStateError), err
		}

		if len(output.Workspaces) == 0 {
			return output, string(types.WorkspaceStateTerminated), nil
		}

		workspace := output.Workspaces[0]

		// https://docs.aws.amazon.com/workspaces/latest/api/API_TerminateWorkspaces.html
		// State TERMINATED is overridden with TERMINATING to catch up directory metadata clean up.
		if workspace.State == types.WorkspaceStateTerminated {
			return workspace, string(types.WorkspaceStateTerminating), nil
		}

		return workspace, string(workspace.State), nil
	}
}
