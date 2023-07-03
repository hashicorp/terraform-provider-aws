// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package workspaces

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/workspaces"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

func FindDirectoryByID(ctx context.Context, conn *workspaces.WorkSpaces, id string) (*workspaces.WorkspaceDirectory, error) {
	input := &workspaces.DescribeWorkspaceDirectoriesInput{
		DirectoryIds: aws.StringSlice([]string{id}),
	}

	output, err := conn.DescribeWorkspaceDirectoriesWithContext(ctx, input)

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.Directories) == 0 || output.Directories[0] == nil {
		return nil, &retry.NotFoundError{
			Message:     "Empty result",
			LastRequest: input,
		}
	}

	// TODO Check for multiple results.
	// TODO https://github.com/hashicorp/terraform-provider-aws/pull/17613.

	directory := output.Directories[0]

	if state := aws.StringValue(directory.State); state == workspaces.WorkspaceDirectoryStateDeregistered {
		return nil, &retry.NotFoundError{
			Message:     state,
			LastRequest: input,
		}
	}

	return directory, nil
}
