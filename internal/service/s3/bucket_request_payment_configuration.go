package s3

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceBucketRequestPaymentConfiguration() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceBucketRequestPaymentConfigurationCreate,
		ReadContext:   resourceBucketRequestPaymentConfigurationRead,
		UpdateContext: resourceBucketRequestPaymentConfigurationUpdate,
		DeleteContext: resourceBucketRequestPaymentConfigurationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"bucket": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 63),
			},
			"expected_bucket_owner": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidAccountID,
			},
			"payer": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(s3.Payer_Values(), false),
			},
		},
	}
}

func resourceBucketRequestPaymentConfigurationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).S3Conn

	bucket := d.Get("bucket").(string)
	expectedBucketOwner := d.Get("expected_bucket_owner").(string)

	input := &s3.PutBucketRequestPaymentInput{
		Bucket: aws.String(bucket),
		RequestPaymentConfiguration: &s3.RequestPaymentConfiguration{
			Payer: aws.String(d.Get("payer").(string)),
		},
	}

	if expectedBucketOwner != "" {
		input.ExpectedBucketOwner = aws.String(expectedBucketOwner)
	}

	_, err := verify.RetryOnAWSCode(s3.ErrCodeNoSuchBucket, func() (interface{}, error) {
		return conn.PutBucketRequestPaymentWithContext(ctx, input)
	})

	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating S3 bucket (%s) request payment configuration: %w", bucket, err))
	}

	d.SetId(resourceBucketRequestPaymentConfigurationCreateResourceID(bucket, expectedBucketOwner))

	return resourceBucketRequestPaymentConfigurationRead(ctx, d, meta)
}

func resourceBucketRequestPaymentConfigurationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).S3Conn

	bucket, expectedBucketOwner, err := resourceBucketRequestPaymentConfigurationParseResourceID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	input := &s3.GetBucketRequestPaymentInput{
		Bucket: aws.String(bucket),
	}

	if expectedBucketOwner != "" {
		input.ExpectedBucketOwner = aws.String(expectedBucketOwner)
	}

	output, err := conn.GetBucketRequestPaymentWithContext(ctx, input)

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, s3.ErrCodeNoSuchBucket) {
		log.Printf("[WARN] S3 Bucket Request Payment Configuration (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if output == nil {
		return diag.FromErr(fmt.Errorf("error reading S3 bucket request payment configuration (%s): empty output", d.Id()))
	}

	d.Set("bucket", bucket)
	d.Set("expected_bucket_owner", expectedBucketOwner)
	d.Set("payer", output.Payer)

	return nil
}

func resourceBucketRequestPaymentConfigurationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).S3Conn

	bucket, expectedBucketOwner, err := resourceBucketRequestPaymentConfigurationParseResourceID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	input := &s3.PutBucketRequestPaymentInput{
		Bucket: aws.String(bucket),
		RequestPaymentConfiguration: &s3.RequestPaymentConfiguration{
			Payer: aws.String(d.Get("payer").(string)),
		},
	}

	if expectedBucketOwner != "" {
		input.ExpectedBucketOwner = aws.String(expectedBucketOwner)
	}

	_, err = conn.PutBucketRequestPaymentWithContext(ctx, input)

	if err != nil {
		return diag.FromErr(fmt.Errorf("error updating S3 bucket request payment configuration (%s): %w", d.Id(), err))
	}

	return resourceBucketRequestPaymentConfigurationRead(ctx, d, meta)
}

func resourceBucketRequestPaymentConfigurationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).S3Conn

	bucket, expectedBucketOwner, err := resourceBucketRequestPaymentConfigurationParseResourceID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	input := &s3.PutBucketRequestPaymentInput{
		Bucket: aws.String(bucket),
		RequestPaymentConfiguration: &s3.RequestPaymentConfiguration{
			// To remove a configuration, it is equivalent to disabling
			// "Requester Pays" in the console; thus, we reset "Payer" back to "BucketOwner"
			Payer: aws.String(s3.PayerBucketOwner),
		},
	}

	if expectedBucketOwner != "" {
		input.ExpectedBucketOwner = aws.String(expectedBucketOwner)
	}

	_, err = conn.PutBucketRequestPaymentWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, s3.ErrCodeNoSuchBucket) {
		return nil
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error deleting S3 bucket request payment configuration (%s): %w", d.Id(), err))
	}

	return nil
}

func resourceBucketRequestPaymentConfigurationCreateResourceID(bucket, expectedBucketOwner string) string {
	if bucket == "" {
		return expectedBucketOwner
	}

	if expectedBucketOwner == "" {
		return bucket
	}

	parts := []string{bucket, expectedBucketOwner}
	id := strings.Join(parts, ",")

	return id
}

func resourceBucketRequestPaymentConfigurationParseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, ",")

	if len(parts) == 1 && parts[0] != "" {
		return parts[0], "", nil
	}

	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected BUCKET or BUCKET,EXPECTED_BUCKET_OWNER", id)
}
