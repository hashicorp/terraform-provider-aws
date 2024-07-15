// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package serverlessrepo

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	serverlessrepo "github.com/aws/aws-sdk-go-v2/service/serverlessapplicationrepository"
	awstypes "github.com/aws/aws-sdk-go-v2/service/serverlessapplicationrepository/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
)

func findApplication(ctx context.Context, conn *serverlessrepo.Client, applicationID, version string) (*serverlessrepo.GetApplicationOutput, error) {
	input := &serverlessrepo.GetApplicationInput{
		ApplicationId: aws.String(applicationID),
	}
	if version != "" {
		input.SemanticVersion = aws.String(version)
	}

	resp, err := conn.GetApplication(ctx, input)
	if errs.IsA[*awstypes.NotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:    err,
			LastRequest:  input,
			LastResponse: resp,
		}
	}
	if err != nil {
		return nil, err
	}

	if resp == nil {
		return nil, &retry.NotFoundError{
			LastRequest:  input,
			LastResponse: resp,
			Message:      "returned empty response",
		}
	}

	return resp, nil
}
