// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package s3control

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/s3control"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	// Maximum amount of time to wait for s3control Bucket state to propagate
	bucketStatePropagationTimeout = 5 * time.Minute
)

// @SDKResource("aws_s3control_bucket", name="Bucket")
// @Tags
// @ArnIdentity
// @Testing(preIdentityVersion="v6.14.1")
// @Testing(preCheck="acctest.PreCheckOutpostsOutposts")
func resourceBucket() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceBucketCreate,
		ReadWithoutTimeout:   resourceBucketRead,
		UpdateWithoutTimeout: resourceBucketUpdate,
		DeleteWithoutTimeout: resourceBucketDelete,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrBucket: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(3, 63),
					validation.StringMatch(regexache.MustCompile(`^[0-9a-z.-]+$`), "must contain only lowercase letters, numbers, periods, and hyphens"),
					validation.StringMatch(regexache.MustCompile(`^[0-9a-z]`), "must begin with lowercase letter or number"),
					validation.StringMatch(regexache.MustCompile(`[0-9a-z]$`), "must end with lowercase letter or number"),
				),
			},
			names.AttrCreationDate: {
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
	}
}

func resourceBucketCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3ControlClient(ctx)

	bucket := d.Get(names.AttrBucket).(string)
	input := &s3control.CreateBucketInput{
		Bucket:    aws.String(bucket),
		OutpostId: aws.String(d.Get("outpost_id").(string)),
	}

	output, err := conn.CreateBucket(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating S3 Control Bucket (%s): %s", bucket, err)
	}

	d.SetId(aws.ToString(output.BucketArn))

	if err := bucketCreateTags(ctx, conn, d.Id(), getS3TagsIn(ctx)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting S3 Control Bucket (%s) tags: %s", d.Id(), err)
	}

	return append(diags, resourceBucketRead(ctx, d, meta)...)
}

func resourceBucketRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3ControlClient(ctx)

	parsedArn, err := arn.Parse(d.Id())

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	// ARN resource format: outpost/<outpost-id>/bucket/<my-bucket-name>
	arnResourceParts := strings.Split(parsedArn.Resource, "/")

	if parsedArn.AccountID == "" || len(arnResourceParts) != 4 {
		return sdkdiag.AppendErrorf(diags, "parsing S3 Control Bucket ARN (%s): unknown format", d.Id())
	}

	output, err := findBucketByTwoPartKey(ctx, conn, parsedArn.AccountID, d.Id())

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] S3 Control Bucket (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading S3 Control Bucket (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, d.Id())
	d.Set(names.AttrBucket, output.Bucket)
	if output.CreationDate != nil {
		d.Set(names.AttrCreationDate, aws.ToTime(output.CreationDate).Format(time.RFC3339))
	}
	d.Set("outpost_id", arnResourceParts[1])
	d.Set("public_access_block_enabled", output.PublicAccessBlockEnabled)

	tags, err := bucketListTags(ctx, conn, d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing tags for S3 Control Bucket (%s): %s", d.Id(), err)
	}

	setS3TagsOut(ctx, svcS3Tags(tags))

	return diags
}

func resourceBucketUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3ControlClient(ctx)

	if d.HasChange(names.AttrTagsAll) {
		o, n := d.GetChange(names.AttrTagsAll)

		if err := bucketUpdateTags(ctx, conn, d.Id(), o, n); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating S3 Control Bucket (%s) tags: %s", d.Id(), err)
		}
	}

	return append(diags, resourceBucketRead(ctx, d, meta)...)
}

func resourceBucketDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3ControlClient(ctx)

	parsedArn, err := arn.Parse(d.Id())

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	input := &s3control.DeleteBucketInput{
		AccountId: aws.String(parsedArn.AccountID),
		Bucket:    aws.String(d.Id()),
	}

	// S3 Control Bucket have a backend state which cannot be checked so this error
	// can occur on deletion:
	//   InvalidBucketState: Bucket is in an invalid state
	log.Printf("[DEBUG] Deleting S3 Control Bucket: %s", d.Id())
	_, err = tfresource.RetryWhenAWSErrCodeEquals(ctx, bucketStatePropagationTimeout, func(ctx context.Context) (any, error) {
		return conn.DeleteBucket(ctx, input)
	}, errCodeInvalidBucketState)

	if tfawserr.ErrCodeEquals(err, errCodeNoSuchBucket, errCodeNoSuchOutpost) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting S3 Control Bucket (%s): %s", d.Id(), err)
	}

	return diags
}

func findBucketByTwoPartKey(ctx context.Context, conn *s3control.Client, accountID, bucket string) (*s3control.GetBucketOutput, error) {
	input := &s3control.GetBucketInput{
		AccountId: aws.String(accountID),
		Bucket:    aws.String(bucket),
	}

	output, err := conn.GetBucket(ctx, input)

	if tfawserr.ErrCodeEquals(err, errCodeNoSuchBucket, errCodeNoSuchOutpost) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output, nil
}
