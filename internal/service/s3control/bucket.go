// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3control

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/s3control"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	// Maximum amount of time to wait for s3control Bucket state to propagate
	bucketStatePropagationTimeout = 5 * time.Minute
)

// @SDKResource("aws_s3control_bucket", name="Bucket")
// @Tags
func resourceBucket() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceBucketCreate,
		ReadWithoutTimeout:   resourceBucketRead,
		UpdateWithoutTimeout: resourceBucketUpdate,
		DeleteWithoutTimeout: resourceBucketDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"bucket": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(3, 63),
					validation.StringMatch(regexp.MustCompile(`^[a-z0-9.-]+$`), "must contain only lowercase letters, numbers, periods, and hyphens"),
					validation.StringMatch(regexp.MustCompile(`^[a-z0-9]`), "must begin with lowercase letter or number"),
					validation.StringMatch(regexp.MustCompile(`[a-z0-9]$`), "must end with lowercase letter or number"),
				),
			},
			"creation_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"outpost_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 64),
			},
			"public_access_block_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceBucketCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).S3ControlConn(ctx)

	bucket := d.Get("bucket").(string)
	input := &s3control.CreateBucketInput{
		Bucket:    aws.String(bucket),
		OutpostId: aws.String(d.Get("outpost_id").(string)),
	}

	output, err := conn.CreateBucketWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("creating S3 Control Bucket (%s): %s", bucket, err)
	}

	d.SetId(aws.StringValue(output.BucketArn))

	if tags := KeyValueTags(ctx, getTagsIn(ctx)); len(tags) > 0 {
		if err := bucketUpdateTags(ctx, conn, d.Id(), nil, tags); err != nil {
			return diag.Errorf("adding S3 Control Bucket (%s) tags: %s", d.Id(), err)
		}
	}

	return resourceBucketRead(ctx, d, meta)
}

func resourceBucketRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).S3ControlConn(ctx)

	parsedArn, err := arn.Parse(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	// ARN resource format: outpost/<outpost-id>/bucket/<my-bucket-name>
	arnResourceParts := strings.Split(parsedArn.Resource, "/")

	if parsedArn.AccountID == "" || len(arnResourceParts) != 4 {
		return diag.Errorf("parsing S3 Control Bucket ARN (%s): unknown format", d.Id())
	}

	output, err := FindBucketByTwoPartKey(ctx, conn, parsedArn.AccountID, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] S3 Control Bucket (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading S3 Control Bucket (%s): %s", d.Id(), err)
	}

	d.Set("arn", d.Id())
	d.Set("bucket", output.Bucket)
	if output.CreationDate != nil {
		d.Set("creation_date", aws.TimeValue(output.CreationDate).Format(time.RFC3339))
	}
	d.Set("outpost_id", arnResourceParts[1])
	d.Set("public_access_block_enabled", output.PublicAccessBlockEnabled)

	tags, err := bucketListTags(ctx, conn, d.Id())

	if err != nil {
		return diag.Errorf("listing tags for S3 Control Bucket (%s): %s", d.Id(), err)
	}

	setTagsOut(ctx, Tags(tags))

	return nil
}

func resourceBucketUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).S3ControlConn(ctx)

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := bucketUpdateTags(ctx, conn, d.Id(), o, n); err != nil {
			return diag.Errorf("updating S3 Control Bucket (%s) tags: %s", d.Id(), err)
		}
	}

	return resourceBucketRead(ctx, d, meta)
}

func resourceBucketDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).S3ControlConn(ctx)

	parsedArn, err := arn.Parse(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	input := &s3control.DeleteBucketInput{
		AccountId: aws.String(parsedArn.AccountID),
		Bucket:    aws.String(d.Id()),
	}

	// S3 Control Bucket have a backend state which cannot be checked so this error
	// can occur on deletion:
	//   InvalidBucketState: Bucket is in an invalid state
	log.Printf("[DEBUG] Deleting S3 Control Bucket: %s", d.Id())
	_, err = tfresource.RetryWhenAWSErrCodeEquals(ctx, bucketStatePropagationTimeout, func() (interface{}, error) {
		return conn.DeleteBucketWithContext(ctx, input)
	}, errCodeInvalidBucketState)

	if tfawserr.ErrCodeEquals(err, errCodeNoSuchBucket, errCodeNoSuchOutpost) {
		return nil
	}

	if err != nil {
		return diag.Errorf("deleting S3 Control Bucket (%s): %s", d.Id(), err)
	}

	return nil
}

func FindBucketByTwoPartKey(ctx context.Context, conn *s3control.S3Control, accountID, bucket string) (*s3control.GetBucketOutput, error) {
	input := &s3control.GetBucketInput{
		AccountId: aws.String(accountID),
		Bucket:    aws.String(bucket),
	}

	output, err := conn.GetBucketWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, errCodeNoSuchBucket, errCodeNoSuchOutpost) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

// Custom S3control tagging functions using similar formatting as other service generated code.

// bucketListTags lists S3control bucket tags.
// The identifier is the bucket ARN.
func bucketListTags(ctx context.Context, conn *s3control.S3Control, identifier string) (tftags.KeyValueTags, error) {
	parsedArn, err := arn.Parse(identifier)

	if err != nil {
		return tftags.New(ctx, nil), err
	}

	input := &s3control.GetBucketTaggingInput{
		AccountId: aws.String(parsedArn.AccountID),
		Bucket:    aws.String(identifier),
	}

	output, err := conn.GetBucketTaggingWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, errCodeNoSuchTagSet) {
		return tftags.New(ctx, nil), nil
	}

	if err != nil {
		return tftags.New(ctx, nil), err
	}

	return KeyValueTags(ctx, output.TagSet), nil
}

// bucketUpdateTags updates S3control bucket tags.
// The identifier is the bucket ARN.
func bucketUpdateTags(ctx context.Context, conn *s3control.S3Control, identifier string, oldTagsMap, newTagsMap any) error {
	parsedArn, err := arn.Parse(identifier)

	if err != nil {
		return err
	}

	oldTags := tftags.New(ctx, oldTagsMap)
	newTags := tftags.New(ctx, newTagsMap)

	// We need to also consider any existing ignored tags.
	allTags, err := bucketListTags(ctx, conn, identifier)

	if err != nil {
		return fmt.Errorf("listing resource tags (%s): %w", identifier, err)
	}

	ignoredTags := allTags.Ignore(oldTags).Ignore(newTags)

	if len(newTags)+len(ignoredTags) > 0 {
		input := &s3control.PutBucketTaggingInput{
			AccountId: aws.String(parsedArn.AccountID),
			Bucket:    aws.String(identifier),
			Tagging: &s3control.Tagging{
				TagSet: Tags(newTags.Merge(ignoredTags)),
			},
		}

		_, err := conn.PutBucketTaggingWithContext(ctx, input)

		if err != nil {
			return fmt.Errorf("setting resource tags (%s): %s", identifier, err)
		}
	} else if len(oldTags) > 0 && len(ignoredTags) == 0 {
		input := &s3control.DeleteBucketTaggingInput{
			AccountId: aws.String(parsedArn.AccountID),
			Bucket:    aws.String(identifier),
		}

		_, err := conn.DeleteBucketTaggingWithContext(ctx, input)

		if err != nil {
			return fmt.Errorf("deleting resource tags (%s): %s", identifier, err)
		}
	}

	return nil
}
