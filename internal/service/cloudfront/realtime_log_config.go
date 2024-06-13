// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudfront

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_cloudfront_realtime_log_config", name="Real-time Log Config")
func resourceRealtimeLogConfig() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceRealtimeLogConfigCreate,
		ReadWithoutTimeout:   resourceRealtimeLogConfigRead,
		UpdateWithoutTimeout: resourceRealtimeLogConfigUpdate,
		DeleteWithoutTimeout: resourceRealtimeLogConfigDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrEndpoint: {
				Type:     schema.TypeList,
				Required: true,
				MinItems: 1,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"kinesis_stream_config": {
							Type:     schema.TypeList,
							Required: true,
							MinItems: 1,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrRoleARN: {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: verify.ValidARN,
									},
									names.AttrStreamARN: {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: verify.ValidARN,
									},
								},
							},
						},
						"stream_type": {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[streamType](),
						},
					},
				},
			},
			"fields": {
				Type:     schema.TypeSet,
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"sampling_rate": {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntBetween(1, 100),
			},
		},
	}
}

func resourceRealtimeLogConfigCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudFrontClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &cloudfront.CreateRealtimeLogConfigInput{
		Name: aws.String(name),
	}

	if v, ok := d.GetOk(names.AttrEndpoint); ok && len(v.([]interface{})) > 0 {
		input.EndPoints = expandEndPoints(v.([]interface{}))
	}

	if v, ok := d.GetOk("fields"); ok && v.(*schema.Set).Len() > 0 {
		input.Fields = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("sampling_rate"); ok {
		input.SamplingRate = aws.Int64(int64(v.(int)))
	}

	output, err := conn.CreateRealtimeLogConfig(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating CloudFront Real-time Log Config (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.RealtimeLogConfig.ARN))

	return append(diags, resourceRealtimeLogConfigRead(ctx, d, meta)...)
}

func resourceRealtimeLogConfigRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudFrontClient(ctx)

	logConfig, err := findRealtimeLogConfigByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] CloudFront Real-time Log Config (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading CloudFront Real-time Log Config (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, logConfig.ARN)
	if err := d.Set(names.AttrEndpoint, flattenEndPoints(logConfig.EndPoints)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting endpoint: %s", err)
	}
	d.Set("fields", logConfig.Fields)
	d.Set(names.AttrName, logConfig.Name)
	d.Set("sampling_rate", logConfig.SamplingRate)

	return diags
}

func resourceRealtimeLogConfigUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudFrontClient(ctx)

	//
	// https://docs.aws.amazon.com/cloudfront/latest/APIReference/API_UpdateRealtimeLogConfig.html:
	// "When you update a real-time log configuration, all the parameters are updated with the values provided in the request. You cannot update some parameters independent of others."
	//
	input := &cloudfront.UpdateRealtimeLogConfigInput{
		ARN: aws.String(d.Id()),
	}

	if v, ok := d.GetOk(names.AttrEndpoint); ok && len(v.([]interface{})) > 0 {
		input.EndPoints = expandEndPoints(v.([]interface{}))
	}

	if v, ok := d.GetOk("fields"); ok && v.(*schema.Set).Len() > 0 {
		input.Fields = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("sampling_rate"); ok {
		input.SamplingRate = aws.Int64(int64(v.(int)))
	}

	_, err := conn.UpdateRealtimeLogConfig(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating CloudFront Real-time Log Config (%s): %s", d.Id(), err)
	}

	return append(diags, resourceRealtimeLogConfigRead(ctx, d, meta)...)
}

func resourceRealtimeLogConfigDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudFrontClient(ctx)

	log.Printf("[DEBUG] Deleting CloudFront Real-time Log Config: %s", d.Id())
	_, err := conn.DeleteRealtimeLogConfig(ctx, &cloudfront.DeleteRealtimeLogConfigInput{
		ARN: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.NoSuchRealtimeLogConfig](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting CloudFront Real-time Log Config (%s): %s", d.Id(), err)
	}

	return diags
}

func findRealtimeLogConfigByARN(ctx context.Context, conn *cloudfront.Client, arn string) (*awstypes.RealtimeLogConfig, error) {
	input := &cloudfront.GetRealtimeLogConfigInput{
		ARN: aws.String(arn),
	}

	return findRealtimeLogConfig(ctx, conn, input)
}

func findRealtimeLogConfig(ctx context.Context, conn *cloudfront.Client, input *cloudfront.GetRealtimeLogConfigInput) (*awstypes.RealtimeLogConfig, error) {
	output, err := conn.GetRealtimeLogConfig(ctx, input)

	if errs.IsA[*awstypes.NoSuchRealtimeLogConfig](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.RealtimeLogConfig == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.RealtimeLogConfig, nil
}

func expandEndPoint(tfMap map[string]interface{}) *awstypes.EndPoint {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.EndPoint{}

	if v, ok := tfMap["kinesis_stream_config"].([]interface{}); ok && len(v) > 0 {
		apiObject.KinesisStreamConfig = expandKinesisStreamConfig(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["stream_type"].(string); ok && v != "" {
		apiObject.StreamType = aws.String(v)
	}

	return apiObject
}

func expandEndPoints(tfList []interface{}) []awstypes.EndPoint {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.EndPoint

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandEndPoint(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func expandKinesisStreamConfig(tfMap map[string]interface{}) *awstypes.KinesisStreamConfig {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.KinesisStreamConfig{}

	if v, ok := tfMap[names.AttrRoleARN].(string); ok && v != "" {
		apiObject.RoleARN = aws.String(v)
	}

	if v, ok := tfMap[names.AttrStreamARN].(string); ok && v != "" {
		apiObject.StreamARN = aws.String(v)
	}

	return apiObject
}

func flattenEndPoint(apiObject *awstypes.EndPoint) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := flattenKinesisStreamConfig(apiObject.KinesisStreamConfig); len(v) > 0 {
		tfMap["kinesis_stream_config"] = []interface{}{v}
	}

	if v := apiObject.StreamType; v != nil {
		tfMap["stream_type"] = aws.ToString(v)
	}

	return tfMap
}

func flattenEndPoints(apiObjects []awstypes.EndPoint) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if v := flattenEndPoint(&apiObject); len(v) > 0 {
			tfList = append(tfList, v)
		}
	}

	return tfList
}

func flattenKinesisStreamConfig(apiObject *awstypes.KinesisStreamConfig) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.RoleARN; v != nil {
		tfMap[names.AttrRoleARN] = aws.ToString(v)
	}

	if v := apiObject.StreamARN; v != nil {
		tfMap[names.AttrStreamARN] = aws.ToString(v)
	}

	return tfMap
}
