// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3control"
	"github.com/aws/aws-sdk-go-v2/service/s3control/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKResource("aws_s3control_bucket_versioning")
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
			"account_id": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 64),
			},
			"bucket": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 255),
					validation.StringMatch(regexp.MustCompile(`^arn:aws:s3-outposts:[a-z0-9-]+:[0-9]{12}:outpost/[^/]+/bucket/[^/]+$`), "Invalid ARN format for S3 Outposts bucket"),
				),
			},
			"versioning_configuration": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"status": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice([]string{"Enabled", "Suspended"}, false),
						},
					},
				},
			},
		},
	}
}

func resourceBucketVersioningCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3ControlClient(ctx)

	accountID := meta.(*conns.AWSClient).AccountID(ctx)
	if v, ok := d.GetOk("account_id"); ok {
		accountID = v.(string)
	}

	bucket := d.Get("bucket").(string)
	if strings.HasPrefix(bucket, "arn:") {
		parts := strings.Split(bucket, "/")
		if len(parts) < 6 {
			return diag.Errorf("Invalid ARN format for S3 Outposts bucket")
		}
		bucket = parts[5]
	}

	input := &s3control.PutBucketVersioningInput{
		AccountId:               aws.String(accountID),
		Bucket:                  aws.String(bucket),
		VersioningConfiguration: expandVersioningConfiguration(d.Get("versioning_configuration").([]interface{})),
	}

	_, err := conn.PutBucketVersioning(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating S3 Control Bucket Versioning (%s): %s", bucket, err)
	}

	d.SetId(fmt.Sprintf("%s:%s", accountID, bucket))

	return append(diags, resourceBucketVersioningRead(ctx, d, meta)...)
}

func resourceBucketVersioningRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3ControlClient(ctx)

	accountID, bucket, err := ResourceBucketVersioningParseID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	output, err := findBucketVersioningByTwoPartKey(ctx, conn, accountID, bucket)
	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] S3 Control Bucket Versioning (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading S3 Control Bucket Versioning (%s): %s", d.Id(), err)
	}

	d.Set("account_id", accountID)
	d.Set("bucket", bucket)
	if err := d.Set("versioning_configuration", flattenVersioningConfiguration(output)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting versioning_configuration: %s", err)
	}

	return diags
}

func resourceBucketVersioningUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3ControlClient(ctx)

	accountID, bucket, err := ResourceBucketVersioningParseID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	input := &s3control.PutBucketVersioningInput{
		AccountId:               aws.String(accountID),
		Bucket:                  aws.String(bucket),
		VersioningConfiguration: expandVersioningConfiguration(d.Get("versioning_configuration").([]interface{})),
	}

	_, err = conn.PutBucketVersioning(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating S3 Control Bucket Versioning (%s): %s", d.Id(), err)
	}

	return append(diags, resourceBucketVersioningRead(ctx, d, meta)...)
}

func resourceBucketVersioningDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3ControlClient(ctx)

	accountID, bucket, err := ResourceBucketVersioningParseID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	input := &s3control.PutBucketVersioningInput{
		AccountId: aws.String(accountID),
		Bucket:    aws.String(bucket),
		VersioningConfiguration: &types.VersioningConfiguration{
			Status: types.BucketVersioningStatusSuspended,
		},
	}

	_, err = conn.PutBucketVersioning(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting S3 Control Bucket Versioning (%s): %s", d.Id(), err)
	}

	return diags
}

func ResourceBucketVersioningParseID(id string) (string, string, error) {
	parts := strings.Split(id, ":")

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format of ID (%s), expected account-id:bucket", id)
	}

	return parts[0], parts[1], nil
}

func findBucketVersioningByTwoPartKey(ctx context.Context, conn *s3control.Client, accountID, bucket string) (*s3control.GetBucketVersioningOutput, error) {
	input := &s3control.GetBucketVersioningInput{
		AccountId: aws.String(accountID),
		Bucket:    aws.String(bucket),
	}

	output, err := conn.GetBucketVersioning(ctx, input)
	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func expandVersioningConfiguration(l []interface{}) *types.VersioningConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})
	if !ok {
		return nil
	}

	result := &types.VersioningConfiguration{}

	if v, ok := tfMap["status"].(string); ok && v != "" {
		result.Status = types.BucketVersioningStatus(v)
	}

	return result
}

func flattenVersioningConfiguration(output *s3control.GetBucketVersioningOutput) []interface{} {
	if output == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"status": output.Status,
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
