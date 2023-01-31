package cloudfront

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudfront"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceRealtimeLogConfig() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceRealtimeLogConfigCreate,
		ReadWithoutTimeout:   resourceRealtimeLogConfigRead,
		UpdateWithoutTimeout: resourceRealtimeLogConfigUpdate,
		DeleteWithoutTimeout: resourceRealtimeLogConfigDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"endpoint": {
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
									"role_arn": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: verify.ValidARN,
									},
									"stream_arn": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: verify.ValidARN,
									},
								},
							},
						},
						"stream_type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(StreamType_Values(), false),
						},
					},
				},
			},
			"fields": {
				Type:     schema.TypeSet,
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"name": {
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
	conn := meta.(*conns.AWSClient).CloudFrontConn()

	name := d.Get("name").(string)
	input := &cloudfront.CreateRealtimeLogConfigInput{
		Name: aws.String(name),
	}

	if v, ok := d.GetOk("endpoint"); ok && len(v.([]interface{})) > 0 {
		input.EndPoints = expandEndPoints(v.([]interface{}))
	}

	if v, ok := d.GetOk("fields"); ok && v.(*schema.Set).Len() > 0 {
		input.Fields = flex.ExpandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("sampling_rate"); ok {
		input.SamplingRate = aws.Int64(int64(v.(int)))
	}

	log.Printf("[DEBUG] Creating CloudFront Real-time Log Config: %s", input)
	output, err := conn.CreateRealtimeLogConfigWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating CloudFront Real-time Log Config (%s): %s", name, err)
	}

	d.SetId(aws.StringValue(output.RealtimeLogConfig.ARN))

	return append(diags, resourceRealtimeLogConfigRead(ctx, d, meta)...)
}

func resourceRealtimeLogConfigRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudFrontConn()

	logConfig, err := FindRealtimeLogConfigByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] CloudFront Real-time Log Config (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading CloudFront Real-time Log Config (%s): %s", d.Id(), err)
	}

	d.Set("arn", logConfig.ARN)
	if err := d.Set("endpoint", flattenEndPoints(logConfig.EndPoints)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting endpoint: %s", err)
	}
	d.Set("fields", aws.StringValueSlice(logConfig.Fields))
	d.Set("name", logConfig.Name)
	d.Set("sampling_rate", logConfig.SamplingRate)

	return diags
}

func resourceRealtimeLogConfigUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudFrontConn()

	//
	// https://docs.aws.amazon.com/cloudfront/latest/APIReference/API_UpdateRealtimeLogConfig.html:
	// "When you update a real-time log configuration, all the parameters are updated with the values provided in the request. You cannot update some parameters independent of others."
	//
	input := &cloudfront.UpdateRealtimeLogConfigInput{
		ARN: aws.String(d.Id()),
	}

	if v, ok := d.GetOk("endpoint"); ok && len(v.([]interface{})) > 0 {
		input.EndPoints = expandEndPoints(v.([]interface{}))
	}

	if v, ok := d.GetOk("fields"); ok && v.(*schema.Set).Len() > 0 {
		input.Fields = flex.ExpandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("sampling_rate"); ok {
		input.SamplingRate = aws.Int64(int64(v.(int)))
	}

	log.Printf("[DEBUG] Updating CloudFront Real-time Log Config: %s", input)
	_, err := conn.UpdateRealtimeLogConfigWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating CloudFront Real-time Log Config (%s): %s", d.Id(), err)
	}

	return append(diags, resourceRealtimeLogConfigRead(ctx, d, meta)...)
}

func resourceRealtimeLogConfigDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudFrontConn()

	log.Printf("[DEBUG] Deleting CloudFront Real-time Log Config (%s)", d.Id())
	_, err := conn.DeleteRealtimeLogConfigWithContext(ctx, &cloudfront.DeleteRealtimeLogConfigInput{
		ARN: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, cloudfront.ErrCodeNoSuchRealtimeLogConfig) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting CloudFront Real-time Log Config (%s): %s", d.Id(), err)
	}

	return diags
}

func expandEndPoint(tfMap map[string]interface{}) *cloudfront.EndPoint {
	if tfMap == nil {
		return nil
	}

	apiObject := &cloudfront.EndPoint{}

	if v, ok := tfMap["kinesis_stream_config"].([]interface{}); ok && len(v) > 0 {
		apiObject.KinesisStreamConfig = expandKinesisStreamConfig(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["stream_type"].(string); ok && v != "" {
		apiObject.StreamType = aws.String(v)
	}

	return apiObject
}

func expandEndPoints(tfList []interface{}) []*cloudfront.EndPoint {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*cloudfront.EndPoint

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandEndPoint(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandKinesisStreamConfig(tfMap map[string]interface{}) *cloudfront.KinesisStreamConfig {
	if tfMap == nil {
		return nil
	}

	apiObject := &cloudfront.KinesisStreamConfig{}

	if v, ok := tfMap["role_arn"].(string); ok && v != "" {
		apiObject.RoleARN = aws.String(v)
	}

	if v, ok := tfMap["stream_arn"].(string); ok && v != "" {
		apiObject.StreamARN = aws.String(v)
	}

	return apiObject
}

func flattenEndPoint(apiObject *cloudfront.EndPoint) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := flattenKinesisStreamConfig(apiObject.KinesisStreamConfig); len(v) > 0 {
		tfMap["kinesis_stream_config"] = []interface{}{v}
	}

	if v := apiObject.StreamType; v != nil {
		tfMap["stream_type"] = aws.StringValue(v)
	}

	return tfMap
}

func flattenEndPoints(apiObjects []*cloudfront.EndPoint) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		if v := flattenEndPoint(apiObject); len(v) > 0 {
			tfList = append(tfList, v)
		}
	}

	return tfList
}

func flattenKinesisStreamConfig(apiObject *cloudfront.KinesisStreamConfig) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.RoleARN; v != nil {
		tfMap["role_arn"] = aws.StringValue(v)
	}

	if v := apiObject.StreamARN; v != nil {
		tfMap["stream_arn"] = aws.StringValue(v)
	}

	return tfMap
}
