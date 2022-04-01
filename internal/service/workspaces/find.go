package workspaces

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/workspaces"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func FindDirectoryByID(conn *workspaces.WorkSpaces, id string) (*workspaces.WorkspaceDirectory, error) {
	input := &workspaces.DescribeWorkspaceDirectoriesInput{
		DirectoryIds: aws.StringSlice([]string{id}),
	}

	output, err := conn.DescribeWorkspaceDirectories(input)

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.Directories) == 0 || output.Directories[0] == nil {
		return nil, &resource.NotFoundError{
			Message:     "Empty result",
			LastRequest: input,
		}
	}

	// TODO Check for multiple results.
	// TODO https://github.com/hashicorp/terraform-provider-aws/pull/17613.

	directory := output.Directories[0]

	if state := aws.StringValue(directory.State); state == workspaces.WorkspaceDirectoryStateDeregistered {
		return nil, &resource.NotFoundError{
			Message:     state,
			LastRequest: input,
		}
	}

	return directory, nil
}
