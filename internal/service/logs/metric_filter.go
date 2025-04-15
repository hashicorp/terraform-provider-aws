// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package logs

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/sdkv2/types/nullable"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_cloudwatch_log_metric_filter", name="Metric Filter")
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
			names.AttrLogGroupName: {
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
						names.AttrDefaultValue: {
							Type:         nullable.TypeNullableFloat,
							Optional:     true,
							ValidateFunc: nullable.ValidateTypeStringNullableFloat,
						},
						"dimensions": {
							Type:     schema.TypeMap,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						names.AttrName: {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validLogMetricFilterTransformationName,
						},
						names.AttrNamespace: {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validLogMetricFilterTransformationName,
						},
						names.AttrUnit: {
							Type:             schema.TypeString,
							Optional:         true,
							Default:          awstypes.StandardUnitNone,
							ValidateDiagFunc: enum.Validate[awstypes.StandardUnit](),
						},
						names.AttrValue: {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(0, 100),
						},
					},
				},
			},
			names.AttrName: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validLogMetricFilterName,
			},
			"pattern": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(0, 1024),
				StateFunc: func(v any) string {
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

func resourceMetricFilterPut(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LogsClient(ctx)

	name := d.Get(names.AttrName).(string)
	logGroupName := d.Get(names.AttrLogGroupName).(string)
	input := &cloudwatchlogs.PutMetricFilterInput{
		FilterName:            aws.String(name),
		FilterPattern:         aws.String(strings.TrimSpace(d.Get("pattern").(string))),
		LogGroupName:          aws.String(logGroupName),
		MetricTransformations: expandMetricTransformations(d.Get("metric_transformation").([]any)),
	}

	// Creating multiple filters on the same log group can sometimes cause
	// clashes, so use a mutex here (and on deletion) to serialise actions on
	// log groups.
	mutexKey := fmt.Sprintf(`log-group-%s`, logGroupName)
	conns.GlobalMutexKV.Lock(mutexKey)
	defer conns.GlobalMutexKV.Unlock(mutexKey)

	_, err := conn.PutMetricFilter(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "putting CloudWatch Logs Metric Filter (%s): %s", d.Id(), err)
	}

	if d.IsNewResource() {
		d.SetId(name)
	}

	return append(diags, resourceMetricFilterRead(ctx, d, meta)...)
}

func resourceMetricFilterRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LogsClient(ctx)

	mf, err := findMetricFilterByTwoPartKey(ctx, conn, d.Get(names.AttrLogGroupName).(string), d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] CloudWatch Logs Metric Filter (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading CloudWatch Logs Metric Filter (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrLogGroupName, mf.LogGroupName)
	if err := d.Set("metric_transformation", flattenMetricTransformations(mf.MetricTransformations)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting metric_transformation: %s", err)
	}
	d.Set(names.AttrName, mf.FilterName)
	d.Set("pattern", mf.FilterPattern)

	return diags
}

func resourceMetricFilterDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LogsClient(ctx)

	// Creating multiple filters on the same log group can sometimes cause
	// clashes, so use a mutex here (and on creation) to serialise actions on
	// log groups.
	mutexKey := fmt.Sprintf(`log-group-%s`, d.Get(names.AttrLogGroupName))
	conns.GlobalMutexKV.Lock(mutexKey)
	defer conns.GlobalMutexKV.Unlock(mutexKey)

	log.Printf("[INFO] Deleting CloudWatch Logs Metric Filter: %s", d.Id())
	_, err := conn.DeleteMetricFilter(ctx, &cloudwatchlogs.DeleteMetricFilterInput{
		FilterName:   aws.String(d.Id()),
		LogGroupName: aws.String(d.Get(names.AttrLogGroupName).(string)),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting CloudWatch Logs Metric Filter (%s): %s", d.Id(), err)
	}

	return diags
}

func resourceMetricFilterImport(d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
	idParts := strings.Split(d.Id(), ":")
	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		return nil, fmt.Errorf("Unexpected format of ID (%q), expected <log_group_name>:<name>", d.Id())
	}
	logGroupName := idParts[0]
	name := idParts[1]
	d.Set(names.AttrLogGroupName, logGroupName)
	d.Set(names.AttrName, name)
	d.SetId(name)
	return []*schema.ResourceData{d}, nil
}

func findMetricFilterByTwoPartKey(ctx context.Context, conn *cloudwatchlogs.Client, logGroupName, name string) (*awstypes.MetricFilter, error) {
	input := cloudwatchlogs.DescribeMetricFiltersInput{
		FilterNamePrefix: aws.String(name),
		LogGroupName:     aws.String(logGroupName),
	}

	return findMetricFilter(ctx, conn, &input, func(v *awstypes.MetricFilter) bool {
		return aws.ToString(v.FilterName) == name
	})
}

func findMetricFilter(ctx context.Context, conn *cloudwatchlogs.Client, input *cloudwatchlogs.DescribeMetricFiltersInput, filter tfslices.Predicate[*awstypes.MetricFilter]) (*awstypes.MetricFilter, error) {
	output, err := findMetricFilters(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findMetricFilters(ctx context.Context, conn *cloudwatchlogs.Client, input *cloudwatchlogs.DescribeMetricFiltersInput, filter tfslices.Predicate[*awstypes.MetricFilter]) ([]awstypes.MetricFilter, error) {
	var output []awstypes.MetricFilter

	pages := cloudwatchlogs.NewDescribeMetricFiltersPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		for _, v := range page.MetricFilters {
			if filter(&v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}

func expandMetricTransformation(tfMap map[string]any) *awstypes.MetricTransformation {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.MetricTransformation{}

	if v, ok := tfMap[names.AttrDefaultValue].(string); ok {
		if v, null, _ := nullable.Float(v).ValueFloat64(); !null {
			apiObject.DefaultValue = aws.Float64(v)
		}
	}

	if v, ok := tfMap["dimensions"].(map[string]any); ok && len(v) > 0 {
		apiObject.Dimensions = flex.ExpandStringValueMap(v)
	}

	if v, ok := tfMap[names.AttrName].(string); ok && v != "" {
		apiObject.MetricName = aws.String(v)
	}

	if v, ok := tfMap[names.AttrNamespace].(string); ok && v != "" {
		apiObject.MetricNamespace = aws.String(v)
	}

	if v, ok := tfMap[names.AttrUnit].(string); ok && v != "" {
		apiObject.Unit = awstypes.StandardUnit(v)
	}

	if v, ok := tfMap[names.AttrValue].(string); ok && v != "" {
		apiObject.MetricValue = aws.String(v)
	}

	return apiObject
}

func expandMetricTransformations(tfList []any) []awstypes.MetricTransformation {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.MetricTransformation

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)

		if !ok {
			continue
		}

		apiObject := expandMetricTransformation(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func flattenMetricTransformation(apiObject awstypes.MetricTransformation) map[string]any {
	tfMap := map[string]any{
		names.AttrUnit: apiObject.Unit,
	}

	if v := apiObject.DefaultValue; v != nil {
		tfMap[names.AttrDefaultValue] = strconv.FormatFloat(aws.ToFloat64(v), 'f', -1, 64)
	}

	if v := apiObject.Dimensions; v != nil {
		tfMap["dimensions"] = v
	}

	if v := apiObject.MetricName; v != nil {
		tfMap[names.AttrName] = aws.ToString(v)
	}

	if v := apiObject.MetricNamespace; v != nil {
		tfMap[names.AttrNamespace] = aws.ToString(v)
	}

	if v := apiObject.MetricValue; v != nil {
		tfMap[names.AttrValue] = aws.ToString(v)
	}

	return tfMap
}

func flattenMetricTransformations(apiObjects []awstypes.MetricTransformation) []any {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []any

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenMetricTransformation(apiObject))
	}

	return tfList
}
