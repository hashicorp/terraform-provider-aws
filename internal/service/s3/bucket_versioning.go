package s3

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceBucketVersioning() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceBucketVersioningCreate,
		ReadContext:   resourceBucketVersioningRead,
		UpdateContext: resourceBucketVersioningUpdate,
		DeleteContext: resourceBucketVersioningDelete,
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
			"mfa": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"versioning_configuration": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"mfa_delete": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.StringInSlice(s3.MFADelete_Values(), false),
						},
						"status": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(s3.BucketVersioningStatus_Values(), false),
						},
					},
				},
			},
		},
	}
}

func resourceBucketVersioningCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).S3Conn

	bucket := d.Get("bucket").(string)
	expectedBucketOwner := d.Get("expected_bucket_owner").(string)

	input := &s3.PutBucketVersioningInput{
		Bucket:                  aws.String(bucket),
		VersioningConfiguration: expandBucketVersioningConfiguration(d.Get("versioning_configuration").([]interface{})),
	}

	if expectedBucketOwner != "" {
		input.ExpectedBucketOwner = aws.String(expectedBucketOwner)
	}

	if v, ok := d.GetOk("mfa"); ok {
		input.MFA = aws.String(v.(string))
	}

	_, err := verify.RetryOnAWSCode(s3.ErrCodeNoSuchBucket, func() (interface{}, error) {
		return conn.PutBucketVersioningWithContext(ctx, input)
	})

	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating S3 bucket versioning for %s: %w", bucket, err))
	}

	d.SetId(CreateResourceID(bucket, expectedBucketOwner))

	return resourceBucketVersioningRead(ctx, d, meta)
}

func resourceBucketVersioningRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).S3Conn

	bucket, expectedBucketOwner, err := ParseResourceID(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	var output *s3.GetBucketVersioningOutput

	if output, err = waitForBucketVersioningStatus(ctx, conn, bucket, expectedBucketOwner); err != nil {
		return diag.Errorf("error waiting for S3 Bucket Versioning status for bucket (%s): %s", d.Id(), err)
	}

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] S3 Bucket Versioning (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("error getting S3 bucket versioning (%s): %s", d.Id(), err)
	}

	d.Set("bucket", bucket)
	d.Set("expected_bucket_owner", expectedBucketOwner)
	if err := d.Set("versioning_configuration", flattenBucketVersioningConfiguration(output)); err != nil {
		return diag.FromErr(fmt.Errorf("error setting versioning_configuration: %w", err))
	}

	return nil
}

func resourceBucketVersioningUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).S3Conn

	bucket, expectedBucketOwner, err := ParseResourceID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	input := &s3.PutBucketVersioningInput{
		Bucket:                  aws.String(bucket),
		VersioningConfiguration: expandBucketVersioningConfiguration(d.Get("versioning_configuration").([]interface{})),
	}

	if expectedBucketOwner != "" {
		input.ExpectedBucketOwner = aws.String(expectedBucketOwner)
	}

	if v, ok := d.GetOk("mfa"); ok {
		input.MFA = aws.String(v.(string))
	}

	_, err = conn.PutBucketVersioningWithContext(ctx, input)

	if err != nil {
		return diag.FromErr(fmt.Errorf("error updating S3 bucket versioning (%s): %w", d.Id(), err))
	}

	return resourceBucketVersioningRead(ctx, d, meta)
}

func resourceBucketVersioningDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).S3Conn

	bucket, expectedBucketOwner, err := ParseResourceID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	input := &s3.PutBucketVersioningInput{
		Bucket: aws.String(bucket),
		VersioningConfiguration: &s3.VersioningConfiguration{
			// Status must be provided thus to "remove" this resource,
			// we suspend versioning
			Status: aws.String(s3.BucketVersioningStatusSuspended),
		},
	}

	if expectedBucketOwner != "" {
		input.ExpectedBucketOwner = aws.String(expectedBucketOwner)
	}

	_, err = conn.PutBucketVersioningWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, s3.ErrCodeNoSuchBucket) {
		return nil
	}

	if tfawserr.ErrMessageContains(err, ErrCodeInvalidBucketState, "An Object Lock configuration is present on this bucket, so the versioning state cannot be changed") {
		log.Printf("[WARN] S3 bucket versioning cannot be suspended with Object Lock Configuration present on bucket (%s), removing from state", bucket)
		return nil
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error deleting S3 bucket versioning (%s): %w", d.Id(), err))
	}

	return nil
}

func expandBucketVersioningConfiguration(l []interface{}) *s3.VersioningConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})
	if !ok {
		return nil
	}

	result := &s3.VersioningConfiguration{}

	if v, ok := tfMap["mfa_delete"].(string); ok && v != "" {
		result.MFADelete = aws.String(v)
	}

	if v, ok := tfMap["status"].(string); ok && v != "" {
		result.Status = aws.String(v)
	}

	return result
}

func flattenBucketVersioningConfiguration(config *s3.GetBucketVersioningOutput) []interface{} {
	if config == nil {
		return []interface{}{}
	}

	m := make(map[string]interface{})

	if config.MFADelete != nil {
		m["mfa_delete"] = aws.StringValue(config.MFADelete)
	}

	if config.Status != nil {
		m["status"] = aws.StringValue(config.Status)
	}

	return []interface{}{m}
}
