// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3_test

import (
	"flag"
	"testing"

	config_sdkv2 "github.com/aws/aws-sdk-go-v2/config"
	s3_sdkv2 "github.com/aws/aws-sdk-go-v2/service/s3"
	session_sdkv1 "github.com/aws/aws-sdk-go/aws/session"
	s3_sdkv1 "github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfs3 "github.com/hashicorp/terraform-provider-aws/internal/service/s3"
)

// AWS_REGION=us-west-2 go test -v ./internal/service/s3 -run=TestEmptyBucket -b ewbankkit-test-empty-bucket-001 -f

var bucket = flag.String("b", "", "bucket")
var force = flag.Bool("f", false, "force")

func TestEmptyBucket(t *testing.T) {
	t.Parallel()

	ctx := acctest.Context(t)

	if *bucket == "" {
		t.Skip("bucket not specified")
	}

	sess := session_sdkv1.Must(session_sdkv1.NewSession())
	svc := s3_sdkv1.New(sess)

	n, err := tfs3.EmptyBucket(ctx, svc, *bucket, *force)

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

	cfg, err := config_sdkv2.LoadDefaultConfig(ctx)
	if err != nil {
		t.Fatalf("error loading default SDK config: %s", err)
	}

	client := s3_sdkv2.NewFromConfig(cfg)
	n, err := tfs3.DeleteAllObjectVersions(ctx, client, *bucket, "", *force, false)

	if err != nil {
		t.Fatalf("error emptying S3 bucket (%s): %s", *bucket, err)
	}

	t.Logf("%d S3 objects deleted", n)
}
