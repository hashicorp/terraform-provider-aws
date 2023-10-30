// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package codebuild

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codebuild"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

const (
	reportGroupStatusUnknown  = "Unknown"
	reportGroupStatusNotFound = "NotFound"
)

// statusReportGroup fetches the Report Group and its Status
func statusReportGroup(ctx context.Context, conn *codebuild.CodeBuild, arn string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindReportGroupByARN(ctx, conn, arn)
		if err != nil {
			return nil, reportGroupStatusUnknown, err
		}

		if output == nil {
			return nil, reportGroupStatusNotFound, nil
		}

		return output, aws.StringValue(output.Status), nil
	}
}
