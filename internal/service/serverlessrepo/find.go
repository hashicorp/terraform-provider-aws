// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package serverlessrepo

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	serverlessrepo "github.com/aws/aws-sdk-go/service/serverlessapplicationrepository"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

func findApplication(ctx context.Context, conn *serverlessrepo.ServerlessApplicationRepository, applicationID, version string) (*serverlessrepo.GetApplicationOutput, error) {
	input := &serverlessrepo.GetApplicationInput{
		ApplicationId: aws.String(applicationID),
	}
	if version != "" {
		input.SemanticVersion = aws.String(version)
	}

	log.Printf("[DEBUG] Getting Serverless findApplication Repository Application: %s", input)
	resp, err := conn.GetApplicationWithContext(ctx, input)
	if tfawserr.ErrCodeEquals(err, serverlessrepo.ErrCodeNotFoundException) {
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
