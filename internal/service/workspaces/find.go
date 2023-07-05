// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package workspaces

import (
	"context"
	"reflect"

	"github.com/aws/aws-sdk-go-v2/service/workspaces"
	"github.com/aws/aws-sdk-go-v2/service/workspaces/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

func FindDirectoryByID(ctx context.Context, conn *workspaces.Client, id string) (*types.WorkspaceDirectory, error) {
	input := &workspaces.DescribeWorkspaceDirectoriesInput{
		DirectoryIds: []string{id},
	}

	output, err := conn.DescribeWorkspaceDirectories(ctx, input)

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.Directories) == 0 || reflect.DeepEqual(output.Directories[0], (types.WorkspaceDirectory{})) {
		return nil, &retry.NotFoundError{
			Message:     "Empty result",
			LastRequest: input,
		}
	}

	// TODO Check for multiple results.
	// TODO https://github.com/hashicorp/terraform-provider-aws/pull/17613.

	directory := output.Directories[0]

	if state := string(directory.State); state == string(types.WorkspaceDirectoryStateDeregistered) {
		return nil, &retry.NotFoundError{
			Message:     state,
			LastRequest: input,
		}
	}

	return &directory, nil
}
