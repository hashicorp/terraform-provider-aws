// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKResource("aws_verifiedaccess_instance_logging_configuration", name="Verified Access Instance Logging Configuration")
func ResourceVerifiedAccessInstanceLoggingConfiguration() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: resourceVerifiedAccessInstanceLoggingConfigurationRead,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"access_logs": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cloudwatch_logs": {
							Type:             schema.TypeList,
							MaxItems:         1,
							Optional:         true,
							DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"enabled": {
										Type:     schema.TypeBool,
										Required: true,
									},
									"log_group": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
						"include_trust_context": {
							Type:     schema.TypeBool,
							Optional: true,
							Computed: true,
						},
						"kinesis_data_firehose": {
							Type:             schema.TypeList,
							Optional:         true,
							MaxItems:         1,
							DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"delivery_stream": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"enabled": {
										Type:     schema.TypeBool,
										Required: true,
									},
								},
							},
						},
						"log_version": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"s3": {
							Type:             schema.TypeList,
							Optional:         true,
							MaxItems:         1,
							DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"bucket_name": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"bucket_owner": {
										Type:         schema.TypeString,
										Optional:     true,
										Computed:     true, // Describe API returns this value if not set
										ValidateFunc: verify.ValidAccountID,
									},
									"enabled": {
										Type:     schema.TypeBool,
										Required: true,
									},
									"prefix": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
					},
				},
			},
			"verifiedaccess_instance_id": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
		},
	}
}

func resourceVerifiedAccessInstanceLoggingConfigurationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	vaiID := d.Id()

	output, err := FindVerifiedAccessInstanceLoggingConfigurationByInstanceID(ctx, conn, vaiID)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 Verified Access Instance Logging Configuration (%s) not found, removing from state", vaiID)
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Verified Access Instance Logging Configuration (%s): %s", vaiID, err)
	}

	if v := output.AccessLogs; v != nil {
		if err := d.Set("access_logs", flattenVerifiedAccessInstanceAccessLogs(v)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting verified access instance access logs: %s", err)
		}
	} else {
		d.Set("access_logs", nil)
	}

	d.Set("verifiedaccess_instance_id", vaiID)

	return diags
}

func flattenVerifiedAccessInstanceAccessLogs(apiObject *types.VerifiedAccessLogs) []interface{} {
	tfMap := map[string]interface{}{}

	if v := apiObject.CloudWatchLogs; v != nil {
		tfMap["cloudwatch_logs"] = flattenVerifiedAccessLogCloudWatchLogs(v)
	}

	if v := apiObject.IncludeTrustContext; v != nil {
		tfMap["include_trust_context"] = aws.ToBool(v)
	}

	if v := apiObject.KinesisDataFirehose; v != nil {
		tfMap["kinesis_data_firehose"] = flattenVerifiedAccessLogKinesisDataFirehose(v)
	}

	if v := apiObject.LogVersion; v != nil {
		tfMap["log_version"] = aws.ToString(v)
	}

	if v := apiObject.S3; v != nil {
		tfMap["s3"] = flattenVerifiedAccessLogS3(v)
	}

	return []interface{}{tfMap}
}

func flattenVerifiedAccessLogCloudWatchLogs(apiObject *types.VerifiedAccessLogCloudWatchLogsDestination) []interface{} {
	tfMap := map[string]interface{}{
		"enabled": bool(*apiObject.Enabled),
	}

	if v := apiObject.LogGroup; v != nil {
		tfMap["log_group"] = aws.ToString(v)
	}

	return []interface{}{tfMap}
}

func flattenVerifiedAccessLogKinesisDataFirehose(apiObject *types.VerifiedAccessLogKinesisDataFirehoseDestination) []interface{} {
	tfMap := map[string]interface{}{
		"enabled": bool(*apiObject.Enabled),
	}

	if v := apiObject.DeliveryStream; v != nil {
		tfMap["delivery_stream"] = aws.ToString(v)
	}

	return []interface{}{tfMap}
}

func flattenVerifiedAccessLogS3(apiObject *types.VerifiedAccessLogS3Destination) []interface{} {
	tfMap := map[string]interface{}{
		"enabled": bool(*apiObject.Enabled),
	}

	if v := apiObject.BucketName; v != nil {
		tfMap["bucket_name"] = aws.ToString(v)
	}

	if v := apiObject.BucketOwner; v != nil {
		tfMap["bucket_owner"] = aws.ToString(v)
	}

	if v := apiObject.Prefix; v != nil {
		tfMap["prefix"] = aws.ToString(v)
	}

	return []interface{}{tfMap}
}
