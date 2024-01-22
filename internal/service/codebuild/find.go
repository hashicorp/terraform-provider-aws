// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package codebuild

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codebuild"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func FindProjectByARN(ctx context.Context, conn *codebuild.CodeBuild, arn string) (*codebuild.Project, error) {
	input := &codebuild.BatchGetProjectsInput{
		Names: []*string{aws.String(arn)},
	}

	output, err := conn.BatchGetProjectsWithContext(ctx, input)
	if err != nil {
		return nil, err
	}

	if output == nil || len(output.Projects) == 0 || output.Projects[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output.Projects); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output.Projects[0], nil
}
