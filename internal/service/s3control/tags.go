// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3control

import (
	"context"
	"strings"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/service/s3control"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/types/option"
)

// isDirectoryBucketARN returns true if the ARN represents an S3 Directory Bucket (S3 Express One Zone).
// Directory Bucket ARNs contain "s3express" in the service portion.
func isDirectoryBucketARN(arn string) bool {
	return strings.Contains(arn, ":s3express:")
}

// ListTags lists s3control service tags and set them in Context.
// This overrides the generated function to handle Directory Buckets correctly.
func (p *servicePackage) ListTags(ctx context.Context, meta any, identifier string) error {
	c := meta.(*conns.AWSClient)

	var conn *s3control.Client
	// Directory Buckets require the S3 Express Control API endpoint
	if isDirectoryBucketARN(identifier) {
		conn = c.S3ExpressControlClient(ctx)
	} else {
		conn = c.S3ControlClient(ctx)
	}

	tags, err := listTagsImpl(ctx, conn, identifier, c.AccountID(ctx))

	if err != nil {
		return smarterr.NewError(err)
	}

	if inContext, ok := tftags.FromContext(ctx); ok {
		inContext.TagsOut = option.Some(tags)
	}

	return nil
}

// UpdateTags updates s3control service tags.
// This overrides the generated function to handle Directory Buckets correctly.
func (p *servicePackage) UpdateTags(ctx context.Context, meta any, identifier string, oldTags, newTags any) error {
	c := meta.(*conns.AWSClient)

	var conn *s3control.Client
	// Directory Buckets require the S3 Express Control API endpoint
	if isDirectoryBucketARN(identifier) {
		conn = c.S3ExpressControlClient(ctx)
	} else {
		conn = c.S3ControlClient(ctx)
	}

	return updateTagsImpl(ctx, conn, identifier, c.AccountID(ctx), oldTags, newTags)
}
