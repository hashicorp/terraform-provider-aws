package sdkresource

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

// @SDKResource("aws_test_thing", name="Test Thing")
func resourceTestThing() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceTestThingCreate,
		ReadWithoutTimeout:   resourceTestThingRead,
		UpdateWithoutTimeout: resourceTestThingUpdate,
		DeleteWithoutTimeout: resourceTestThingDelete,
	}
}

func resourceTestThingCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).S3Client(ctx)
	conn.CreateBucket(ctx, &s3.CreateBucketInput{})
	conn.PutBucketTagging(ctx, &s3.PutBucketTaggingInput{})
	return nil
}

func resourceTestThingRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).S3Client(ctx)
	findTestThing(ctx, conn, "id")
	return nil
}

func resourceTestThingUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).S3Client(ctx)
	conn.PutBucketTagging(ctx, &s3.PutBucketTaggingInput{})
	return nil
}

func resourceTestThingDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).S3Client(ctx)
	conn.DeleteBucket(ctx, &s3.DeleteBucketInput{})
	return nil
}

func findTestThing(ctx context.Context, conn *s3.Client, id string) (*s3.GetBucketLocationOutput, error) {
	return conn.GetBucketLocation(ctx, &s3.GetBucketLocationInput{})
}
