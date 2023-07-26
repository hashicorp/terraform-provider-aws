// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3

import (
	"context"

	aws_sdkv1 "github.com/aws/aws-sdk-go/aws"
	request_sdkv1 "github.com/aws/aws-sdk-go/aws/request"
	session_sdkv1 "github.com/aws/aws-sdk-go/aws/session"
	s3_sdkv1 "github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
)

// NewConn returns a new AWS SDK for Go v1 client for this service package's AWS API.
func (p *servicePackage) NewConn(ctx context.Context, m map[string]any) (*s3_sdkv1.S3, error) {
	sess := m["session"].(*session_sdkv1.Session)
	config := &aws_sdkv1.Config{
		Endpoint:         aws_sdkv1.String(m["endpoint"].(string)),
		S3ForcePathStyle: aws_sdkv1.Bool(m["s3_use_path_style"].(bool)),
	}

	return s3_sdkv1.New(sess.Copy(config)), nil
}

// CustomizeConn customizes a new AWS SDK for Go v1 client for this service package's AWS API.
func (p *servicePackage) CustomizeConn(ctx context.Context, conn *s3_sdkv1.S3) (*s3_sdkv1.S3, error) {
	conn.Handlers.Retry.PushBack(func(r *request_sdkv1.Request) {
		if tfawserr.ErrMessageContains(r.Error, errCodeOperationAborted, "A conflicting conditional operation is currently in progress against this resource. Please try again.") {
			r.Retryable = aws_sdkv1.Bool(true)
		}
	})

	return conn, nil
}
