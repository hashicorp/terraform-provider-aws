// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3_test

import (
	"flag"
	"testing"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfs3 "github.com/hashicorp/terraform-provider-aws/internal/service/s3"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// AWS_REGION=us-west-2 go test -v ./internal/service/s3 -run=TestEmptyBucket -b ewbankkit-test-empty-bucket-001 -f

var bucket = flag.String("b", "", names.AttrBucket)
var force = flag.Bool("f", false, "force")

func TestEmptyBucket(t *testing.T) {
	t.Parallel()

	ctx := acctest.Context(t)

	if *bucket == "" {
		t.Skip("bucket not specified")
	}

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		t.Fatalf("error loading default SDK config: %s", err)
	}

	client := s3.NewFromConfig(cfg)
	n, err := tfs3.EmptyBucket(ctx, client, *bucket, *force)

	if err != nil {
		t.Fatalf("error emptying S3 bucket (%s): %s", *bucket, err)
	}

	t.Logf("%d S3 objects deleted", n)
}

func TestDeleteAllObjectVersions(t *testing.T) {
	t.Parallel()

	ctx := acctest.Context(t)

	if *bucket == "" {
		t.Skip("bucket not specified")
	}

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		t.Fatalf("error loading default SDK config: %s", err)
	}

	client := s3.NewFromConfig(cfg)
	n, err := tfs3.DeleteAllObjectVersions(ctx, client, *bucket, "", *force, false)

	if err != nil {
		t.Fatalf("error emptying S3 bucket (%s): %s", *bucket, err)
	}

	t.Logf("%d S3 objects deleted", n)
}
