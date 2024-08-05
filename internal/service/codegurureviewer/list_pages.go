// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package codegurureviewer

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/codegurureviewer"
)

// Custom CodeGuruReviewer service lister functions using the same format as generated code.

func listRepositoryAssociationsPages(ctx context.Context, conn *codegurureviewer.Client, input *codegurureviewer.ListRepositoryAssociationsInput, fn func(*codegurureviewer.ListRepositoryAssociationsOutput, bool) bool) error {
	for {
		output, err := conn.ListRepositoryAssociations(ctx, input)
		if err != nil {
			return err
		}

		lastPage := aws.ToString(output.NextToken) == ""
		if !fn(output, lastPage) || lastPage {
			break
		}

		input.NextToken = output.NextToken
	}
	return nil
}
