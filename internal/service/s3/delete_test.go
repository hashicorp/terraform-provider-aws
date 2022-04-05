package s3_test

import (
	"context"
	"flag"
	"testing"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	tfs3 "github.com/hashicorp/terraform-provider-aws/internal/service/s3"
)

var bucket = flag.String("b", "", "bucket")
var force = flag.Bool("f", false, "force")

func TestEmptyBucket(t *testing.T) {
	if *bucket == "" {
		t.Skip("bucket not specified")
	}

	sess := session.Must(session.NewSession())
	svc := s3.New(sess)
	ctx := context.Background()

	n, err := tfs3.EmptyBucket(ctx, svc, *bucket, *force)

	if err != nil {
		t.Fatalf("error emptying S3 bucket (%s): %s", *bucket, err)
	}

	t.Logf("%d S3 objects deleted", n)
}
