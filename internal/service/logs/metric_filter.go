package logs

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/experimental/nullable"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func init() {
	_sp.registerSDKResourceFactory("aws_cloudwatch_log_metric_filter", resourceMetricFilter)
}

func resourceMetricFilter() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceMetricFilterPut,
		ReadWithoutTimeout:   resourceMetricFilterRead,
		UpdateWithoutTimeout: resourceMetricFilterPut,
		DeleteWithoutTimeout: resourceMetricFilterDelete,

		Importer: &schema.ResourceImporter{
			State: resourceMetricFilterImport,
		},

		Schema: map[string]*schema.Schema{
			"log_group_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validLogGroupName,
			},
			"metric_transformation": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"default_value": {
							Type:         nullable.TypeNullableFloat,
							Optional:     true,
							ValidateFunc: nullable.ValidateTypeStringNullableFloat,
						},
						"dimensions": {
							Type:     schema.TypeMap,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"name": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validLogMetricFilterTransformationName,
						},
						"namespace": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validLogMetricFilterTransformationName,
						},
						"unit": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      cloudwatchlogs.StandardUnitNone,
							ValidateFunc: validation.StringInSlice(cloudwatchlogs.StandardUnit_Values(), false),
						},
						"value": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(0, 100),
						},
					},
				},
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validLogMetricFilterName,
			},
			"pattern": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(0, 1024),
				StateFunc: func(v interface{}) string {
					s, ok := v.(string)
					if !ok {
						return ""
					}
					return strings.TrimSpace(s)
				},
			},
		},
	}
}

func resourceMetricFilterPut(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).LogsConn()

	name := d.Get("name").(string)
	logGroupName := d.Get("log_group_name").(string)
	input := &cloudwatchlogs.PutMetricFilterInput{
		FilterName:            aws.String(name),
		FilterPattern:         aws.String(strings.TrimSpace(d.Get("pattern").(string))),
		LogGroupName:          aws.String(logGroupName),
		MetricTransformations: expandMetricTransformations(d.Get("metric_transformation").([]interface{})),
	}

	// Creating multiple filters on the same log group can sometimes cause
	// clashes, so use a mutex here (and on deletion) to serialise actions on
	// log groups.
	mutexKey := fmt.Sprintf(`log-group-%s`, logGroupName)
	conns.GlobalMutexKV.Lock(mutexKey)
	defer conns.GlobalMutexKV.Unlock(mutexKey)

	_, err := conn.PutMetricFilterWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("putting CloudWatch Logs Metric Filter (%s): %s", d.Id(), err)
	}

	if d.IsNewResource() {
		d.SetId(name)
	}

	return resourceMetricFilterRead(ctx, d, meta)
}

func resourceMetricFilterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).LogsConn()

	mf, err := FindMetricFilterByTwoPartKey(ctx, conn, d.Get("log_group_name").(string), d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] CloudWatch Logs Metric Filter (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading CloudWatch Logs Metric Filter (%s): %s", d.Id(), err)
	}

	d.Set("log_group_name", mf.LogGroupName)
	if err := d.Set("metric_transformation", flattenMetricTransformations(mf.MetricTransformations)); err != nil {
		return diag.Errorf("setting metric_transformation: %s", err)
	}
	d.Set("name", mf.FilterName)
	d.Set("pattern", mf.FilterPattern)

	return nil
}

func resourceMetricFilterDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).LogsConn()

	// Creating multiple filters on the same log group can sometimes cause
	// clashes, so use a mutex here (and on creation) to serialise actions on
	// log groups.
	mutexKey := fmt.Sprintf(`log-group-%s`, d.Get(`log_group_name`))
	conns.GlobalMutexKV.Lock(mutexKey)
	defer conns.GlobalMutexKV.Unlock(mutexKey)

	log.Printf("[INFO] Deleting CloudWatch Logs Metric Filter: %s", d.Id())
	_, err := conn.DeleteMetricFilterWithContext(ctx, &cloudwatchlogs.DeleteMetricFilterInput{
		FilterName:   aws.String(d.Id()),
		LogGroupName: aws.String(d.Get("log_group_name").(string)),
	})

	if tfawserr.ErrCodeEquals(err, cloudwatchlogs.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return diag.Errorf("deleting CloudWatch Logs Metric Filter (%s): %s", d.Id(), err)
	}

	return nil
}

func resourceMetricFilterImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	idParts := strings.Split(d.Id(), ":")
	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		return nil, fmt.Errorf("Unexpected format of ID (%q), expected <log_group_name>:<name>", d.Id())
	}
	logGroupName := idParts[0]
	name := idParts[1]
	d.Set("log_group_name", logGroupName)
	d.Set("name", name)
	d.SetId(name)
	return []*schema.ResourceData{d}, nil
}

func FindMetricFilterByTwoPartKey(ctx context.Context, conn *cloudwatchlogs.CloudWatchLogs, logGroupName, name string) (*cloudwatchlogs.MetricFilter, error) {
	input := &cloudwatchlogs.DescribeMetricFiltersInput{
		FilterNamePrefix: aws.String(name),
		LogGroupName:     aws.String(logGroupName),
	}
	var output *cloudwatchlogs.MetricFilter

	err := conn.DescribeMetricFiltersPagesWithContext(ctx, input, func(page *cloudwatchlogs.DescribeMetricFiltersOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.MetricFilters {
			if aws.StringValue(v.FilterName) == name {
				output = v

				return false
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, cloudwatchlogs.ErrCodeResourceNotFoundException) {
		return nil, &resource.NotFoundError{
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

func expandMetricTransformation(tfMap map[string]interface{}) *cloudwatchlogs.MetricTransformation {
	if tfMap == nil {
		return nil
	}

	apiObject := &cloudwatchlogs.MetricTransformation{}

	if v, ok := tfMap["default_value"].(string); ok {
		if v, null, _ := nullable.Float(v).Value(); !null {
			apiObject.DefaultValue = aws.Float64(v)
		}
	}

	if v, ok := tfMap["dimensions"].(map[string]interface{}); ok && len(v) > 0 {
		apiObject.Dimensions = flex.ExpandStringMap(v)
	}

	if v, ok := tfMap["name"].(string); ok && v != "" {
		apiObject.MetricName = aws.String(v)
	}

	if v, ok := tfMap["namespace"].(string); ok && v != "" {
		apiObject.MetricNamespace = aws.String(v)
	}

	if v, ok := tfMap["unit"].(string); ok && v != "" {
		apiObject.Unit = aws.String(v)
	}

	if v, ok := tfMap["value"].(string); ok && v != "" {
		apiObject.MetricValue = aws.String(v)
	}

	return apiObject
}

func expandMetricTransformations(tfList []interface{}) []*cloudwatchlogs.MetricTransformation {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*cloudwatchlogs.MetricTransformation

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandMetricTransformation(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func flattenMetricTransformation(apiObject *cloudwatchlogs.MetricTransformation) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.DefaultValue; v != nil {
		tfMap["default_value"] = strconv.FormatFloat(aws.Float64Value(v), 'f', -1, 64)
	}

	if v := apiObject.Dimensions; v != nil {
		tfMap["dimensions"] = aws.StringValueMap(v)
	}

	if v := apiObject.MetricName; v != nil {
		tfMap["name"] = aws.StringValue(v)
	}

	if v := apiObject.MetricNamespace; v != nil {
		tfMap["namespace"] = aws.StringValue(v)
	}

	if v := apiObject.Unit; v != nil {
		tfMap["unit"] = aws.StringValue(v)
	}

	if v := apiObject.MetricValue; v != nil {
		tfMap["value"] = aws.StringValue(v)
	}

	return tfMap
}

func flattenMetricTransformations(apiObjects []*cloudwatchlogs.MetricTransformation) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenMetricTransformation(apiObject))
	}

	return tfList
}
