// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fms

import (
	"context"

	aws_sdkv1 "github.com/aws/aws-sdk-go/aws"
	request_sdkv1 "github.com/aws/aws-sdk-go/aws/request"
	fms_sdkv1 "github.com/aws/aws-sdk-go/service/fms"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
)

// CustomizeConn customizes a new AWS SDK for Go v1 client for this service package's AWS API.
func (p *servicePackage) CustomizeConn(ctx context.Context, conn *fms_sdkv1.FMS) (*fms_sdkv1.FMS, error) {
	conn.Handlers.Retry.PushBack(func(r *request_sdkv1.Request) {
		// Acceptance testing creates and deletes resources in quick succession.
		// The FMS onboarding process into Organizations is opaque to consumers.
		// Since we cannot reasonably check this status before receiving the error,
		// set the operation as retryable.
		switch r.Operation.Name {
		case "AssociateAdminAccount":
			if tfawserr.ErrMessageContains(r.Error, fms_sdkv1.ErrCodeInvalidOperationException, "Your AWS Organization is currently offboarding with AWS Firewall Manager. Please submit onboard request after offboarded.") {
				r.Retryable = aws_sdkv1.Bool(true)
			}
		case "DisassociateAdminAccount":
			if tfawserr.ErrMessageContains(r.Error, fms_sdkv1.ErrCodeInvalidOperationException, "Your AWS Organization is currently onboarding with AWS Firewall Manager and cannot be offboarded.") {
				r.Retryable = aws_sdkv1.Bool(true)
			}
		// System problems can arise during FMS policy updates (maybe also creation),
		// so we set the following operation as retryable.
		// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/23946
		case "PutPolicy":
			if tfawserr.ErrCodeEquals(r.Error, fms_sdkv1.ErrCodeInternalErrorException) {
				r.Retryable = aws_sdkv1.Bool(true)
			}
		}
	})

	return conn, nil
}
