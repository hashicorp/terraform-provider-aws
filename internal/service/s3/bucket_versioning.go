// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_s3_bucket_versioning", name="Bucket Versioning")
func resourceBucketVersioning() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceBucketVersioningCreate,
		ReadWithoutTimeout:   resourceBucketVersioningRead,
		UpdateWithoutTimeout: resourceBucketVersioningUpdate,
		DeleteWithoutTimeout: resourceBucketVersioningDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrBucket: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 63),
			},
			names.AttrExpectedBucketOwner: {
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
							Type:             schema.TypeString,
							Optional:         true,
							Computed:         true,
							ValidateDiagFunc: enum.Validate[types.MFADelete](),
						},
						names.AttrStatus: {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(bucketVersioningStatus_Values(), false),
						},
					},
				},
			},
		},

		CustomizeDiff: customdiff.Sequence(
			func(_ context.Context, diff *schema.ResourceDiff, v interface{}) error {
				// This CustomizeDiff acts as a plan-time validation to prevent MalformedXML errors
				// when updating bucket versioning to "Disabled" on existing resources
				// as it's not supported by the AWS S3 API.
				if diff.Id() == "" || !diff.HasChange("versioning_configuration.0.status") {
					return nil
				}

				oldStatusRaw, newStatusRaw := diff.GetChange("versioning_configuration.0.status")
				oldStatus, newStatus := oldStatusRaw.(string), newStatusRaw.(string)

				if newStatus == bucketVersioningStatusDisabled && (oldStatus == string(types.BucketVersioningStatusEnabled) || oldStatus == string(types.BucketVersioningStatusSuspended)) {
					return fmt.Errorf("versioning_configuration.status cannot be updated from '%s' to '%s'", oldStatus, newStatus)
				}

				return nil
			},
		),
	}
}

func resourceBucketVersioningCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3Client(ctx)

	bucket := d.Get(names.AttrBucket).(string)
	expectedBucketOwner := d.Get(names.AttrExpectedBucketOwner).(string)

	versioningConfiguration := expandBucketVersioningConfiguration(d.Get("versioning_configuration").([]interface{}))

	// To support migration from v3 to v4 of the provider, we need to support
	// versioning resources that represent unversioned S3 buckets as was previously
	// supported within the aws_s3_bucket resource of the 3.x version of the provider.
	// Thus, we essentially bring existing bucket versioning into adoption.
	if string(versioningConfiguration.Status) != bucketVersioningStatusDisabled {
		input := &s3.PutBucketVersioningInput{
			Bucket:                  aws.String(bucket),
			VersioningConfiguration: versioningConfiguration,
		}
		if expectedBucketOwner != "" {
			input.ExpectedBucketOwner = aws.String(expectedBucketOwner)
		}

		if v, ok := d.GetOk("mfa"); ok {
			input.MFA = aws.String(v.(string))
		}

		_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, bucketPropagationTimeout, func() (interface{}, error) {
			return conn.PutBucketVersioning(ctx, input)
		}, errCodeNoSuchBucket)

		if tfawserr.ErrMessageContains(err, errCodeInvalidArgument, "VersioningConfiguration is not valid, expected CreateBucketConfiguration") {
			err = errDirectoryBucket(err)
		}

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "creating S3 Bucket (%s) Versioning: %s", bucket, err)
		}
	} else {
		log.Printf("[DEBUG] Creating S3 bucket versioning for unversioned bucket: %s", bucket)
	}

	d.SetId(CreateResourceID(bucket, expectedBucketOwner))

	// Waiting for the versioning configuration to appear is done in resource Read.

	return append(diags, resourceBucketVersioningRead(ctx, d, meta)...)
}

func resourceBucketVersioningRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3Client(ctx)

	bucket, expectedBucketOwner, err := ParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	output, err := waitForBucketVersioningStatus(ctx, conn, bucket, expectedBucketOwner)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] S3 Bucket Versioning (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading S3 Bucket Versioning (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrBucket, bucket)
	d.Set(names.AttrExpectedBucketOwner, expectedBucketOwner)
	if err := d.Set("versioning_configuration", flattenVersioning(output)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting versioning_configuration: %s", err)
	}

	return diags
}

func resourceBucketVersioningUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3Client(ctx)

	bucket, expectedBucketOwner, err := ParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
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

	_, err = conn.PutBucketVersioning(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating S3 Bucket Versioning (%s): %s", d.Id(), err)
	}

	return append(diags, resourceBucketVersioningRead(ctx, d, meta)...)
}

func resourceBucketVersioningDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3Client(ctx)

	if v := expandBucketVersioningConfiguration(d.Get("versioning_configuration").([]interface{})); v != nil && string(v.Status) == bucketVersioningStatusDisabled {
		log.Printf("[DEBUG] Removing S3 bucket versioning for unversioned bucket (%s) from state", d.Id())
		return diags
	}

	bucket, expectedBucketOwner, err := ParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	input := &s3.PutBucketVersioningInput{
		Bucket: aws.String(bucket),
		VersioningConfiguration: &types.VersioningConfiguration{
			// Status must be provided thus to "remove" this resource,
			// we suspend versioning
			Status: types.BucketVersioningStatusSuspended,
		},
	}
	if expectedBucketOwner != "" {
		input.ExpectedBucketOwner = aws.String(expectedBucketOwner)
	}

	_, err = conn.PutBucketVersioning(ctx, input)

	if tfawserr.ErrCodeEquals(err, errCodeNoSuchBucket) {
		return diags
	}

	if tfawserr.ErrMessageContains(err, errCodeInvalidBucketState, "An Object Lock configuration is present on this bucket, so the versioning state cannot be changed") {
		log.Printf("[WARN] S3 bucket versioning cannot be suspended with Object Lock Configuration present on bucket (%s), removing from state", bucket)
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting S3 Bucket Versioning (%s): %s", d.Id(), err)
	}

	// Don't wait for the versioning configuration to disappear as it still exists after suspension.

	return diags
}

func expandBucketVersioningConfiguration(l []interface{}) *types.VersioningConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})
	if !ok {
		return nil
	}

	result := &types.VersioningConfiguration{}

	if v, ok := tfMap["mfa_delete"].(string); ok && v != "" {
		result.MFADelete = types.MFADelete(v)
	}

	if v, ok := tfMap[names.AttrStatus].(string); ok && v != "" {
		result.Status = types.BucketVersioningStatus(v)
	}

	return result
}

func flattenVersioning(config *s3.GetBucketVersioningOutput) []interface{} {
	if config == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"mfa_delete": config.MFADelete,
	}

	if config.Status != "" {
		m[names.AttrStatus] = config.Status
	} else {
		// Bucket Versioning by default is disabled but not set in the config struct's Status field
		m[names.AttrStatus] = bucketVersioningStatusDisabled
	}

	return []interface{}{m}
}

func findBucketVersioning(ctx context.Context, conn *s3.Client, bucket, expectedBucketOwner string) (*s3.GetBucketVersioningOutput, error) {
	input := &s3.GetBucketVersioningInput{
		Bucket: aws.String(bucket),
	}
	if expectedBucketOwner != "" {
		input.ExpectedBucketOwner = aws.String(expectedBucketOwner)
	}

	output, err := conn.GetBucketVersioning(ctx, input)

	if tfawserr.ErrCodeEquals(err, errCodeNoSuchBucket) {
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

func statusBucketVersioning(ctx context.Context, conn *s3.Client, bucket, expectedBucketOwner string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findBucketVersioning(ctx, conn, bucket, expectedBucketOwner)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		if output.Status == "" {
			return output, bucketVersioningStatusDisabled, nil
		}

		return output, string(output.Status), nil
	}
}

func waitForBucketVersioningStatus(ctx context.Context, conn *s3.Client, bucket, expectedBucketOwner string) (*s3.GetBucketVersioningOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{""},
		Target:                    bucketVersioningStatus_Values(),
		Refresh:                   statusBucketVersioning(ctx, conn, bucket, expectedBucketOwner),
		Timeout:                   bucketPropagationTimeout,
		ContinuousTargetOccurence: 3,
		NotFoundChecks:            3,
		Delay:                     1 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*s3.GetBucketVersioningOutput); ok {
		return output, err
	}

	return nil, err
}

const (
	bucketVersioningStatusDisabled = "Disabled"
)

func bucketVersioningStatus_Values() []string {
	return tfslices.AppendUnique(enum.Values[types.BucketVersioningStatus](), bucketVersioningStatusDisabled)
}
