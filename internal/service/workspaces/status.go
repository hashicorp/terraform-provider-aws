package workspaces

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/workspaces"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func StatusDirectoryState(conn *workspaces.WorkSpaces, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindDirectoryByID(conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.State), nil
	}
}

// nosemgrep: workspaces-in-func-name
func StatusWorkspaceState(conn *workspaces.WorkSpaces, workspaceID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := conn.DescribeWorkspaces(&workspaces.DescribeWorkspacesInput{
			WorkspaceIds: aws.StringSlice([]string{workspaceID}),
		})
		if err != nil {
			return nil, workspaces.WorkspaceStateError, err
		}

		if len(output.Workspaces) == 0 {
			return output, workspaces.WorkspaceStateTerminated, nil
		}

		workspace := output.Workspaces[0]

		// https://docs.aws.amazon.com/workspaces/latest/api/API_TerminateWorkspaces.html
		// State TERMINATED is overridden with TERMINATING to catch up directory metadata clean up.
		if aws.StringValue(workspace.State) == workspaces.WorkspaceStateTerminated {
			return workspace, workspaces.WorkspaceStateTerminating, nil
		}

		return workspace, aws.StringValue(workspace.State), nil
	}
}
