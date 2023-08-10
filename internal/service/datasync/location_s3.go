// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package datasync

import (
	"context"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/datasync"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_datasync_location_s3", name="Location S3")
// @Tags(identifierAttribute="id")
func ResourceLocationS3() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceLocationS3Create,
		ReadWithoutTimeout:   resourceLocationS3Read,
		UpdateWithoutTimeout: resourceLocationS3Update,
		DeleteWithoutTimeout: resourceLocationS3Delete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"agent_arns": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: verify.ValidARN,
				},
			},
			"s3_bucket_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"s3_config": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"bucket_access_role_arn": {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: verify.ValidARN,
						},
					},
				},
			},
			"s3_storage_class": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(datasync.S3StorageClass_Values(), false),
			},
			"subdirectory": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				// Ignore missing trailing slash
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					if new == "/" {
						return false
					}
					if strings.TrimSuffix(old, "/") == strings.TrimSuffix(new, "/") {
						return true
					}
					return false
				},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"uri": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceLocationS3Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DataSyncConn(ctx)

	input := &datasync.CreateLocationS3Input{
		S3BucketArn:  aws.String(d.Get("s3_bucket_arn").(string)),
		S3Config:     expandS3Config(d.Get("s3_config").([]interface{})),
		Subdirectory: aws.String(d.Get("subdirectory").(string)),
		Tags:         getTagsIn(ctx),
	}

	if v, ok := d.GetOk("agent_arns"); ok {
		input.AgentArns = flex.ExpandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("s3_storage_class"); ok {
		input.S3StorageClass = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating DataSync Location S3: %s", input)

	var output *datasync.CreateLocationS3Output
	err := retry.RetryContext(ctx, propagationTimeout, func() *retry.RetryError {
		var err error
		output, err = conn.CreateLocationS3WithContext(ctx, input)

		// Retry for IAM eventual consistency on error:
		// InvalidRequestException: Unable to assume role. Reason: Access denied when calling sts:AssumeRole
		if tfawserr.ErrMessageContains(err, datasync.ErrCodeInvalidRequestException, "Unable to assume role") {
			return retry.RetryableError(err)
		}

		// Retry for IAM eventual consistency on error:
		// InvalidRequestException: DataSync location access test failed: could not perform s3:ListObjectsV2 on bucket
		if tfawserr.ErrMessageContains(err, datasync.ErrCodeInvalidRequestException, "access test failed") {
			return retry.RetryableError(err)
		}

		if err != nil {
			return retry.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		output, err = conn.CreateLocationS3WithContext(ctx, input)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating DataSync Location S3: %s", err)
	}

	d.SetId(aws.StringValue(output.LocationArn))

	return append(diags, resourceLocationS3Read(ctx, d, meta)...)
}

func resourceLocationS3Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DataSyncConn(ctx)

	input := &datasync.DescribeLocationS3Input{
		LocationArn: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Reading DataSync Location S3: %s", input)
	output, err := conn.DescribeLocationS3WithContext(ctx, input)

	if tfawserr.ErrMessageContains(err, "InvalidRequestException", "not found") {
		log.Printf("[WARN] DataSync Location S3 %q not found - removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading DataSync Location S3 (%s): %s", d.Id(), err)
	}

	subdirectory, err := subdirectoryFromLocationURI(aws.StringValue(output.LocationUri))

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading DataSync Location S3 (%s): %s", d.Id(), err)
	}

	d.Set("agent_arns", flex.FlattenStringSet(output.AgentArns))
	d.Set("arn", output.LocationArn)
	if err := d.Set("s3_config", flattenS3Config(output.S3Config)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting s3_config: %s", err)
	}
	d.Set("s3_storage_class", output.S3StorageClass)
	d.Set("subdirectory", subdirectory)
	d.Set("uri", output.LocationUri)

	return diags
}

func resourceLocationS3Update(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// Tags only.

	return append(diags, resourceLocationS3Read(ctx, d, meta)...)
}

func resourceLocationS3Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DataSyncConn(ctx)

	input := &datasync.DeleteLocationInput{
		LocationArn: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Deleting DataSync Location S3: %s", input)
	_, err := conn.DeleteLocationWithContext(ctx, input)

	if tfawserr.ErrMessageContains(err, "InvalidRequestException", "not found") {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting DataSync Location S3 (%s): %s", d.Id(), err)
	}

	return diags
}

func flattenS3Config(s3Config *datasync.S3Config) []interface{} {
	if s3Config == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"bucket_access_role_arn": aws.StringValue(s3Config.BucketAccessRoleArn),
	}

	return []interface{}{m}
}

func expandS3Config(l []interface{}) *datasync.S3Config {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	s3Config := &datasync.S3Config{
		BucketAccessRoleArn: aws.String(m["bucket_access_role_arn"].(string)),
	}

	return s3Config
}
