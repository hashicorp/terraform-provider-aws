// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appconfig

import (
	"context"

	aws_sdkv1 "github.com/aws/aws-sdk-go/aws"
	request_sdkv1 "github.com/aws/aws-sdk-go/aws/request"
	appconfig_sdkv1 "github.com/aws/aws-sdk-go/service/appconfig"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
)

// CustomizeConn customizes a new AWS SDK for Go v1 client for this service package's AWS API.
func (p *servicePackage) CustomizeConn(ctx context.Context, conn *appconfig_sdkv1.AppConfig) (*appconfig_sdkv1.AppConfig, error) {
	// StartDeployment operations can return a ConflictException
	// if ongoing deployments are in-progress, thus we handle them
	// here for the service client.
	conn.Handlers.Retry.PushBack(func(r *request_sdkv1.Request) {
		if r.Operation.Name == "StartDeployment" {
			if tfawserr.ErrCodeEquals(r.Error, appconfig_sdkv1.ErrCodeConflictException) {
				r.Retryable = aws_sdkv1.Bool(true)
			}
		}
	})

	return conn, nil
}
