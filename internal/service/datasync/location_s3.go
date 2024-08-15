// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package datasync

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/datasync"
	awstypes "github.com/aws/aws-sdk-go-v2/service/datasync/types"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_datasync_location_s3", name="Location S3")
// @Tags(identifierAttribute="id")
func resourceLocationS3() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceLocationS3Create,
		ReadWithoutTimeout:   resourceLocationS3Read,
		UpdateWithoutTimeout: resourceLocationS3Update,
		DeleteWithoutTimeout: resourceLocationS3Delete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"agent_arns": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: verify.ValidARN,
				},
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
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
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[awstypes.S3StorageClass](),
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
			names.AttrURI: {
				Type:     schema.TypeString,
				Computed: true,
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceLocationS3Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DataSyncClient(ctx)

	input := &datasync.CreateLocationS3Input{
		S3BucketArn:  aws.String(d.Get("s3_bucket_arn").(string)),
		S3Config:     expandS3Config(d.Get("s3_config").([]interface{})),
		Subdirectory: aws.String(d.Get("subdirectory").(string)),
		Tags:         getTagsIn(ctx),
	}

	if v, ok := d.GetOk("agent_arns"); ok {
		input.AgentArns = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("s3_storage_class"); ok {
		input.S3StorageClass = awstypes.S3StorageClass(v.(string))
	}

	outputRaw, err := tfresource.RetryWhen(ctx, propagationTimeout,
		func() (interface{}, error) {
			return conn.CreateLocationS3(ctx, input)
		},
		func(err error) (bool, error) {
			// Retry for IAM eventual consistency on error:
			// InvalidRequestException: Unable to assume role. Reason: Access denied when calling sts:AssumeRole
			if errs.IsAErrorMessageContains[*awstypes.InvalidRequestException](err, "Unable to assume role") {
				return true, err
			}

			// Retry for IAM eventual consistency on error:
			// InvalidRequestException: DataSync location access test failed: could not perform s3:ListObjectsV2 on bucket
			if errs.IsAErrorMessageContains[*awstypes.InvalidRequestException](err, "access test failed") {
				return true, err
			}

			return false, err
		})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating DataSync Location S3: %s", err)
	}

	d.SetId(aws.ToString(outputRaw.(*datasync.CreateLocationS3Output).LocationArn))

	return append(diags, resourceLocationS3Read(ctx, d, meta)...)
}

func resourceLocationS3Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DataSyncClient(ctx)

	output, err := findLocationS3ByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] DataSync Location S3 (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading DataSync Location S3 (%s): %s", d.Id(), err)
	}

	uri := aws.ToString(output.LocationUri)
	s3BucketName, err := globalIDFromLocationURI(aws.ToString(output.LocationUri))
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}
	subdirectory, err := subdirectoryFromLocationURI(aws.ToString(output.LocationUri))
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}
	locationARN, err := arn.Parse(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	d.Set("agent_arns", output.AgentArns)
	d.Set(names.AttrARN, output.LocationArn)
	s3BucketArn := fmt.Sprintf("arn:%s:s3:::%s", locationARN.Partition, s3BucketName)
	d.Set("s3_bucket_arn", s3BucketArn)
	if err := d.Set("s3_config", flattenS3Config(output.S3Config)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting s3_config: %s", err)
	}
	d.Set("s3_storage_class", output.S3StorageClass)
	d.Set("subdirectory", subdirectory)
	d.Set(names.AttrURI, uri)

	return diags
}

func resourceLocationS3Update(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// Tags only.

	return append(diags, resourceLocationS3Read(ctx, d, meta)...)
}

func resourceLocationS3Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DataSyncClient(ctx)

	log.Printf("[DEBUG] Deleting DataSync Location S3: %s", d.Id())
	_, err := conn.DeleteLocation(ctx, &datasync.DeleteLocationInput{
		LocationArn: aws.String(d.Id()),
	})

	if errs.IsAErrorMessageContains[*awstypes.InvalidRequestException](err, "not found") {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting DataSync Location S3 (%s): %s", d.Id(), err)
	}

	return diags
}

func findLocationS3ByARN(ctx context.Context, conn *datasync.Client, arn string) (*datasync.DescribeLocationS3Output, error) {
	input := &datasync.DescribeLocationS3Input{
		LocationArn: aws.String(arn),
	}

	output, err := conn.DescribeLocationS3(ctx, input)

	if errs.IsAErrorMessageContains[*awstypes.InvalidRequestException](err, "not found") {
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

func flattenS3Config(s3Config *awstypes.S3Config) []interface{} {
	if s3Config == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"bucket_access_role_arn": aws.ToString(s3Config.BucketAccessRoleArn),
	}

	return []interface{}{m}
}

func expandS3Config(l []interface{}) *awstypes.S3Config {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	s3Config := &awstypes.S3Config{
		BucketAccessRoleArn: aws.String(m["bucket_access_role_arn"].(string)),
	}

	return s3Config
}
