// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/go-uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	defaultVerifiedAccessLogVersion = "ocsf-1.0.0-rc.2"
)

// @SDKResource("aws_verifiedaccess_instance_logging_configuration", name="Verified Access Instance Logging Configuration")
func resourceVerifiedAccessInstanceLoggingConfiguration() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceVerifiedAccessInstanceLoggingConfigurationCreate,
		ReadWithoutTimeout:   resourceVerifiedAccessInstanceLoggingConfigurationRead,
		UpdateWithoutTimeout: resourceVerifiedAccessInstanceLoggingConfigurationUpdate,
		DeleteWithoutTimeout: resourceVerifiedAccessInstanceLoggingConfigurationDelete,

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
						names.AttrCloudWatchLogs: {
							Type:             schema.TypeList,
							MaxItems:         1,
							Optional:         true,
							DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrEnabled: {
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
									names.AttrEnabled: {
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
									names.AttrBucketName: {
										Type:     schema.TypeString,
										Optional: true,
									},
									"bucket_owner": {
										Type:         schema.TypeString,
										Optional:     true,
										Computed:     true, // Describe API returns this value if not set
										ValidateFunc: verify.ValidAccountID,
									},
									names.AttrEnabled: {
										Type:     schema.TypeBool,
										Required: true,
									},
									names.AttrPrefix: {
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

func resourceVerifiedAccessInstanceLoggingConfigurationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	vaiID := d.Get("verifiedaccess_instance_id").(string)

	uuid, err := uuid.GenerateUUID()
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "generating uuid for ClientToken for Verified Access Instance Logging Configuration %s): %s", vaiID, err)
	}

	input := &ec2.ModifyVerifiedAccessInstanceLoggingConfigurationInput{
		AccessLogs:               expandVerifiedAccessInstanceAccessLogs(d.Get("access_logs").([]interface{})),
		ClientToken:              aws.String(uuid), // can't use aws.String(id.UniqueId()), because it's not a valid uuid
		VerifiedAccessInstanceId: aws.String(vaiID),
	}

	output, err := conn.ModifyVerifiedAccessInstanceLoggingConfiguration(ctx, input)

	if err != nil || output == nil {
		return sdkdiag.AppendErrorf(diags, "creating Verified Access Instance Logging Configuration (%s): %s", vaiID, err)
	}

	d.SetId(vaiID)

	return append(diags, resourceVerifiedAccessInstanceLoggingConfigurationRead(ctx, d, meta)...)
}

func resourceVerifiedAccessInstanceLoggingConfigurationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	vaiID := d.Id()
	output, err := findVerifiedAccessInstanceLoggingConfigurationByInstanceID(ctx, conn, vaiID)

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

func resourceVerifiedAccessInstanceLoggingConfigurationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	vaiID := d.Id()

	if d.HasChange("access_logs") {
		uuid, err := uuid.GenerateUUID()
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "generating uuid for ClientToken for Verified Access Instance Logging Configuration %s): %s", vaiID, err)
		}

		input := &ec2.ModifyVerifiedAccessInstanceLoggingConfigurationInput{
			AccessLogs:               expandVerifiedAccessInstanceAccessLogs(d.Get("access_logs").([]interface{})),
			ClientToken:              aws.String(uuid), // can't use aws.String(id.UniqueId()), because it's not a valid uuid
			VerifiedAccessInstanceId: aws.String(vaiID),
		}

		_, err = conn.ModifyVerifiedAccessInstanceLoggingConfiguration(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Verified Access Instance Logging Configuration (%s): %s", vaiID, err)
		}
	}

	return append(diags, resourceVerifiedAccessInstanceLoggingConfigurationRead(ctx, d, meta)...)
}

func resourceVerifiedAccessInstanceLoggingConfigurationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	vaiID := d.Id()

	// create structure for reset
	resetObject := &types.VerifiedAccessLogOptions{
		CloudWatchLogs: &types.VerifiedAccessLogCloudWatchLogsDestinationOptions{
			Enabled: aws.Bool(false),
		},
		KinesisDataFirehose: &types.VerifiedAccessLogKinesisDataFirehoseDestinationOptions{
			Enabled: aws.Bool(false),
		},
		S3: &types.VerifiedAccessLogS3DestinationOptions{
			Enabled: aws.Bool(false),
		},
		IncludeTrustContext: aws.Bool(false),
		// reset log_version because ocsf-0.1 is not compatible with enabling include_trust_context
		// without reset, if practitioners previously applied and destroyed with ocsf-0.1,
		// ocsf-0.1 will be the new "default" value, leading to errors with include_trust_context
		LogVersion: aws.String(defaultVerifiedAccessLogVersion),
	}

	uuid, err := uuid.GenerateUUID()
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "generating uuid for ClientToken for Verified Access Instance Logging Configuration %s): %s", vaiID, err)
	}

	log.Printf("[INFO] Deleting Verified Access Instance Logging Configuration: %s", vaiID)
	input := &ec2.ModifyVerifiedAccessInstanceLoggingConfigurationInput{
		AccessLogs:               resetObject,
		ClientToken:              aws.String(uuid), // can't use aws.String(id.UniqueId()), because it's not a valid uuid
		VerifiedAccessInstanceId: aws.String(vaiID),
	}

	_, err = conn.ModifyVerifiedAccessInstanceLoggingConfiguration(ctx, input)

	if tfawserr.ErrCodeEquals(err, errCodeInvalidVerifiedAccessInstanceIdNotFound) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Verified Access Instance Logging Configuration (%s): %s", vaiID, err)
	}

	return diags
}

func expandVerifiedAccessInstanceAccessLogs(accessLogs []interface{}) *types.VerifiedAccessLogOptions {
	if len(accessLogs) == 0 || accessLogs[0] == nil {
		return nil
	}

	tfMap, ok := accessLogs[0].(map[string]interface{})
	if !ok {
		return nil
	}

	result := &types.VerifiedAccessLogOptions{}

	if v, ok := tfMap[names.AttrCloudWatchLogs].([]interface{}); ok && len(v) > 0 {
		result.CloudWatchLogs = expandVerifiedAccessLogCloudWatchLogs(v)
	}

	if v, ok := tfMap["include_trust_context"].(bool); ok {
		result.IncludeTrustContext = aws.Bool(v)
	}

	if v, ok := tfMap["kinesis_data_firehose"].([]interface{}); ok && len(v) > 0 {
		result.KinesisDataFirehose = expandVerifiedAccessLogKinesisDataFirehose(v)
	}

	if v, ok := tfMap["log_version"].(string); ok && v != "" {
		result.LogVersion = aws.String(v)
	}

	if v, ok := tfMap["s3"].([]interface{}); ok && len(v) > 0 {
		result.S3 = expandVerifiedAccessLogS3(v)
	}

	return result
}

func expandVerifiedAccessLogCloudWatchLogs(cloudWatchLogs []interface{}) *types.VerifiedAccessLogCloudWatchLogsDestinationOptions {
	if len(cloudWatchLogs) == 0 || cloudWatchLogs[0] == nil {
		return nil
	}

	tfMap, ok := cloudWatchLogs[0].(map[string]interface{})
	if !ok {
		return nil
	}

	result := &types.VerifiedAccessLogCloudWatchLogsDestinationOptions{
		Enabled: aws.Bool(tfMap[names.AttrEnabled].(bool)),
	}

	if v, ok := tfMap["log_group"].(string); ok && v != "" {
		result.LogGroup = aws.String(v)
	}

	return result
}

func expandVerifiedAccessLogKinesisDataFirehose(kinesisDataFirehose []interface{}) *types.VerifiedAccessLogKinesisDataFirehoseDestinationOptions {
	if len(kinesisDataFirehose) == 0 || kinesisDataFirehose[0] == nil {
		return nil
	}

	tfMap, ok := kinesisDataFirehose[0].(map[string]interface{})
	if !ok {
		return nil
	}

	result := &types.VerifiedAccessLogKinesisDataFirehoseDestinationOptions{
		Enabled: aws.Bool(tfMap[names.AttrEnabled].(bool)),
	}

	if v, ok := tfMap["delivery_stream"].(string); ok && v != "" {
		result.DeliveryStream = aws.String(v)
	}

	return result
}

func expandVerifiedAccessLogS3(s3 []interface{}) *types.VerifiedAccessLogS3DestinationOptions {
	if len(s3) == 0 || s3[0] == nil {
		return nil
	}

	tfMap, ok := s3[0].(map[string]interface{})
	if !ok {
		return nil
	}

	result := &types.VerifiedAccessLogS3DestinationOptions{
		Enabled: aws.Bool(tfMap[names.AttrEnabled].(bool)),
	}

	if v, ok := tfMap[names.AttrBucketName].(string); ok && v != "" {
		result.BucketName = aws.String(v)
	}

	// if enabled is true, pass bucket owner, otherwise don't pass it
	// api error InvalidParameterCombination: The parameter AccessLogs.S3.BucketOwner cannot be used when AccessLogs.S3.Enabled is false
	if v, ok := tfMap[names.AttrEnabled].(bool); ok && v {
		if v, ok := tfMap["bucket_owner"].(string); ok && v != "" {
			result.BucketOwner = aws.String(v)
		}
	}

	if v, ok := tfMap[names.AttrPrefix].(string); ok && v != "" {
		result.Prefix = aws.String(v)
	}

	return result
}

func flattenVerifiedAccessInstanceAccessLogs(apiObject *types.VerifiedAccessLogs) []interface{} {
	tfMap := map[string]interface{}{}

	if v := apiObject.CloudWatchLogs; v != nil {
		tfMap[names.AttrCloudWatchLogs] = flattenVerifiedAccessLogCloudWatchLogs(v)
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
		names.AttrEnabled: apiObject.Enabled,
	}

	if v := apiObject.LogGroup; v != nil {
		tfMap["log_group"] = aws.ToString(v)
	}

	return []interface{}{tfMap}
}

func flattenVerifiedAccessLogKinesisDataFirehose(apiObject *types.VerifiedAccessLogKinesisDataFirehoseDestination) []interface{} {
	tfMap := map[string]interface{}{
		names.AttrEnabled: apiObject.Enabled,
	}

	if v := apiObject.DeliveryStream; v != nil {
		tfMap["delivery_stream"] = aws.ToString(v)
	}

	return []interface{}{tfMap}
}

func flattenVerifiedAccessLogS3(apiObject *types.VerifiedAccessLogS3Destination) []interface{} {
	tfMap := map[string]interface{}{
		names.AttrEnabled: apiObject.Enabled,
	}

	if v := apiObject.BucketName; v != nil {
		tfMap[names.AttrBucketName] = aws.ToString(v)
	}

	if v := apiObject.BucketOwner; v != nil {
		tfMap["bucket_owner"] = aws.ToString(v)
	}

	if v := apiObject.Prefix; v != nil {
		tfMap[names.AttrPrefix] = aws.ToString(v)
	}

	return []interface{}{tfMap}
}
