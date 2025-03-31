// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package macie2

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/macie2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/macie2/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_macie2_classification_export_configuration", name="Classification Export Configuration")
func resourceClassificationExportConfiguration() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceClassificationExportConfigurationCreate,
		UpdateWithoutTimeout: resourceClassificationExportConfigurationUpdate,
		DeleteWithoutTimeout: resourceClassificationExportConfigurationDelete,
		ReadWithoutTimeout:   resourceClassificationExportConfigurationRead,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"s3_destination": {
				Type:         schema.TypeList,
				Optional:     true,
				MaxItems:     1,
				AtLeastOneOf: []string{"s3_destination"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrBucketName: {
							Type:     schema.TypeString,
							Required: true,
						},
						"key_prefix": {
							Type:     schema.TypeString,
							Optional: true,
						},
						names.AttrKMSKeyARN: {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: verify.ValidARN,
						},
					},
				},
			},
		},
	}
}

func resourceClassificationExportConfigurationCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).Macie2Client(ctx)

	if d.IsNewResource() {
		output, err := conn.GetClassificationExportConfiguration(ctx, &macie2.GetClassificationExportConfigurationInput{})
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading Macie classification export configuration failed: %s", err)
		}

		if (awstypes.ClassificationExportConfiguration{}) != *output.Configuration { // nosemgrep:ci.semgrep.aws.prefer-pointer-conversion-conditional
			return sdkdiag.AppendErrorf(diags, "creating Macie classification export configuration: a configuration already exists")
		}
	}

	input := macie2.PutClassificationExportConfigurationInput{
		Configuration: &awstypes.ClassificationExportConfiguration{},
	}

	if v, ok := d.GetOk("s3_destination"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		input.Configuration.S3Destination = expandClassificationExportConfiguration(v.([]any)[0].(map[string]any))
	}

	log.Printf("[DEBUG] Creating Macie classification export configuration: %+v", input)

	_, err := conn.PutClassificationExportConfiguration(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Macie classification export configuration failed: %s", err)
	}

	return append(diags, resourceClassificationExportConfigurationRead(ctx, d, meta)...)
}

func resourceClassificationExportConfigurationUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).Macie2Client(ctx)

	input := macie2.PutClassificationExportConfigurationInput{
		Configuration: &awstypes.ClassificationExportConfiguration{},
	}

	if v, ok := d.GetOk("s3_destination"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		input.Configuration.S3Destination = expandClassificationExportConfiguration(v.([]any)[0].(map[string]any))
	} else {
		input.Configuration.S3Destination = nil
	}

	log.Printf("[DEBUG] Creating Macie classification export configuration: %+v", input)

	_, err := conn.PutClassificationExportConfiguration(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Macie classification export configuration failed: %s", err)
	}

	return append(diags, resourceClassificationExportConfigurationRead(ctx, d, meta)...)
}

func resourceClassificationExportConfigurationRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).Macie2Client(ctx)

	input := macie2.GetClassificationExportConfigurationInput{} // api does not have a getById() like endpoint.
	output, err := conn.GetClassificationExportConfiguration(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Macie classification export configuration failed: %s", err)
	}

	if (awstypes.ClassificationExportConfiguration{}) != *output.Configuration { // nosemgrep:ci.semgrep.aws.prefer-pointer-conversion-conditional
		if (awstypes.S3Destination{}) != *output.Configuration.S3Destination { // nosemgrep:ci.semgrep.aws.prefer-pointer-conversion-conditional
			var flattenedS3Destination = flattenClassificationExportConfigurationS3DestinationResult(output.Configuration.S3Destination)
			if err := d.Set("s3_destination", []any{flattenedS3Destination}); err != nil {
				return sdkdiag.AppendErrorf(diags, "setting Macie classification export configuration s3_destination: %s", err)
			}
		}
		d.SetId(fmt.Sprintf("%s:%s:%s", "macie:classification_export_configuration", meta.(*conns.AWSClient).AccountID(ctx), meta.(*conns.AWSClient).Region(ctx)))
	}

	return diags
}

func resourceClassificationExportConfigurationDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).Macie2Client(ctx)

	input := macie2.PutClassificationExportConfigurationInput{
		Configuration: &awstypes.ClassificationExportConfiguration{},
	}

	log.Printf("[DEBUG] deleting Macie classification export configuration: %+v", input)

	_, err := conn.PutClassificationExportConfiguration(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Macie classification export configuration failed: %s", err)
	}

	return diags
}

func expandClassificationExportConfiguration(tfMap map[string]any) *awstypes.S3Destination {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.S3Destination{}

	if v, ok := tfMap[names.AttrBucketName].(string); ok {
		apiObject.BucketName = aws.String(v)
	}

	if v, ok := tfMap["key_prefix"].(string); ok {
		apiObject.KeyPrefix = aws.String(v)
	}

	if v, ok := tfMap[names.AttrKMSKeyARN].(string); ok {
		apiObject.KmsKeyArn = aws.String(v)
	}

	return apiObject
}

func flattenClassificationExportConfigurationS3DestinationResult(apiObject *awstypes.S3Destination) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if v := apiObject.BucketName; v != nil {
		tfMap[names.AttrBucketName] = aws.ToString(v)
	}

	if v := apiObject.KeyPrefix; v != nil {
		tfMap["key_prefix"] = aws.ToString(v)
	}

	if v := apiObject.KmsKeyArn; v != nil {
		tfMap[names.AttrKMSKeyARN] = aws.ToString(v)
	}

	return tfMap
}
