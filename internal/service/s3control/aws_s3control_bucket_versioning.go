// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3control

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/s3control"
	"github.com/aws/aws-sdk-go-v2/service/s3control/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_s3control_bucket_versioning")
func ResourceBucketVersioning() *schema.Resource {
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
				ValidateFunc: validateS3ControlBucketName,
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

	bucket := d.Get(names.AttrBucket).(string)

	accountID := meta.(*conns.AWSClient).AccountID(ctx)
	if strings.HasPrefix(bucket, "arn:") {
		parsedARN, err := arn.Parse(bucket)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "parsing S3 Control Bucket ARN (%s): %s", bucket, err)
		}
		accountID = parsedARN.AccountID
	}

	input := &s3control.PutBucketVersioningInput{
		AccountId: aws.String(accountID),
		Bucket:    aws.String(bucket),
		VersioningConfiguration: &types.VersioningConfiguration{
			Status: types.BucketVersioningStatus(expandVersioningStatus(d.Get("versioning_configuration").([]interface{}))),
		},
	}

	_, err := conn.PutBucketVersioning(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating S3 Control Bucket Versioning (%s): %s", bucket, err)
	}

	d.SetId(bucket)

	return append(diags, resourceBucketVersioningRead(ctx, d, meta)...)
}

func resourceBucketVersioningRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3ControlClient(ctx)

	bucket := d.Id()
	accountID := meta.(*conns.AWSClient).AccountID(ctx)

	if strings.HasPrefix(bucket, "arn:") {
		parsedARN, err := arn.Parse(bucket)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "parsing S3 Control Bucket ARN (%s): %s", bucket, err)
		}
		accountID = parsedARN.AccountID
	}

	input := &s3control.GetBucketVersioningInput{
		AccountId: aws.String(accountID),
		Bucket:    aws.String(bucket),
	}

	output, err := conn.GetBucketVersioning(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading S3 Control Bucket Versioning (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrBucket, bucket)
	if err := d.Set("versioning_configuration", flattenVersioningConfiguration(output)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting versioning_configuration: %s", err)
	}

	return diags
}

func resourceBucketVersioningUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3ControlClient(ctx)

	bucket := d.Id()
	accountID := meta.(*conns.AWSClient).AccountID(ctx)
	if strings.HasPrefix(bucket, "arn:") {
		parsedARN, err := arn.Parse(bucket)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "parsing S3 Control Bucket ARN (%s): %s", bucket, err)
		}
		accountID = parsedARN.AccountID
	}

	input := &s3control.PutBucketVersioningInput{
		AccountId: aws.String(accountID),
		Bucket:    aws.String(bucket),
		VersioningConfiguration: &types.VersioningConfiguration{
			Status: types.BucketVersioningStatus(expandVersioningStatus(d.Get("versioning_configuration").([]interface{}))),
		},
	}

	_, err := conn.PutBucketVersioning(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating S3 Control Bucket Versioning (%s): %s", d.Id(), err)
	}

	return append(diags, resourceBucketVersioningRead(ctx, d, meta)...)
}

func resourceBucketVersioningDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).S3ControlClient(ctx)

	bucket := d.Id()
	accountID := meta.(*conns.AWSClient).AccountID(ctx)

	if strings.HasPrefix(bucket, "arn:") {
		parsedARN, err := arn.Parse(bucket)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "parsing S3 Control Bucket ARN (%s): %s", bucket, err)
		}
		accountID = parsedARN.AccountID
	}

	input := &s3control.PutBucketVersioningInput{
		AccountId: aws.String(accountID),
		Bucket:    aws.String(bucket),
		VersioningConfiguration: &types.VersioningConfiguration{
			Status: types.BucketVersioningStatusSuspended,
		},
	}

	_, err := conn.PutBucketVersioning(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "suspending S3 Control Bucket Versioning (%s): %s", d.Id(), err)
	}

	return diags
}
func isValidOutpostBucketArn(value string) bool {
	prefix := "arn:aws:s3-outposts:"
	return strings.HasPrefix(value, prefix)
}

func validateS3ControlBucketName(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)

	if strings.HasPrefix(value, "arn:") {
		if !isValidOutpostBucketArn(value) {
			errors = append(errors, fmt.Errorf("%q must be a valid S3 Outposts bucket ARN", k))
		}
	} else {
		if len(value) < 1 || len(value) > 63 {
			errors = append(errors, fmt.Errorf("%q must be between 1 and 63 characters", k))
		}
	}

	return
}

func expandVersioningStatus(l []interface{}) string {
	if len(l) == 0 || l[0] == nil {
		return ""
	}

	m := l[0].(map[string]interface{})
	return m["status"].(string)
}

func flattenVersioningConfiguration(output *s3control.GetBucketVersioningOutput) []interface{} {
	if output == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"status": aws.String(string(output.Status)),
	}

	return []interface{}{m}
}
