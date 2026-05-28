// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package logs

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/sdkv2"
	"github.com/hashicorp/terraform-provider-aws/internal/sdkv2/types/nullable"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_cloudwatch_log_metric_filter", name="Metric Filter")
// @IdentityAttribute("log_group_name")
// @IdentityAttribute("name")
// @ImportIDHandler("metricFilterImportID")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types;awstypes;awstypes.MetricFilter")
// @Testing(importStateIdFunc=testAccMetricFilterImportStateIDFunc)
// @Testing(preIdentityVersion="v6.41.0")
func resourceMetricFilter() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceMetricFilterPut,
		ReadWithoutTimeout:   resourceMetricFilterRead,
		UpdateWithoutTimeout: resourceMetricFilterPut,
		DeleteWithoutTimeout: resourceMetricFilterDelete,

		Schema: map[string]*schema.Schema{
			"apply_on_transformed_logs": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
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
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: verify.StringUTF8LenBetween(0, 1024),
				StateFunc:        sdkv2.TrimSpaceSchemaStateFunc,
			},
		},
	}
}

func resourceMetricFilterPut(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).LogsClient(ctx)

	name := d.Get(names.AttrName).(string)
	logGroupName := d.Get(names.AttrLogGroupName).(string)
	input := cloudwatchlogs.PutMetricFilterInput{
		FilterName:            aws.String(name),
		FilterPattern:         aws.String(strings.TrimSpace(d.Get("pattern").(string))),
		LogGroupName:          aws.String(logGroupName),
		MetricTransformations: expandMetricTransformations(d.Get("metric_transformation").([]any)),
	}

	if v, ok := d.GetOk("apply_on_transformed_logs"); ok {
		input.ApplyOnTransformedLogs = v.(bool)
	}

	// Creating multiple filters on the same log group can sometimes cause
	// clashes, so use a mutex here (and on deletion) to serialise actions on
	// log groups.
	mutexKey := fmt.Sprintf(`log-group-%s`, logGroupName)
	conns.GlobalMutexKV.Lock(mutexKey)
	defer conns.GlobalMutexKV.Unlock(mutexKey)

	_, err := conn.PutMetricFilter(ctx, &input)

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

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] CloudWatch Logs Metric Filter (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading CloudWatch Logs Metric Filter (%s): %s", d.Id(), err)
	}

	if err := resourceMetricFilterFlatten(ctx, mf, d); err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

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
	input := cloudwatchlogs.DeleteMetricFilterInput{
		FilterName:   aws.String(d.Id()),
		LogGroupName: aws.String(d.Get(names.AttrLogGroupName).(string)),
	}
	_, err := conn.DeleteMetricFilter(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting CloudWatch Logs Metric Filter (%s): %s", d.Id(), err)
	}

	return diags
}

func resourceMetricFilterFlatten(_ context.Context, mf *awstypes.MetricFilter, d *schema.ResourceData) error {
	d.Set("apply_on_transformed_logs", mf.ApplyOnTransformedLogs)
	d.Set(names.AttrLogGroupName, mf.LogGroupName)
	if err := d.Set("metric_transformation", flattenMetricTransformations(mf.MetricTransformations)); err != nil {
		return fmt.Errorf("setting metric_transformation: %w", err)
	}
	d.Set(names.AttrName, mf.FilterName)
	d.Set("pattern", mf.FilterPattern)

	return nil
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
	output, err := findMetricFilters(ctx, conn, input, filter, tfslices.WithReturnFirstMatch)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findMetricFilters(ctx context.Context, conn *cloudwatchlogs.Client, input *cloudwatchlogs.DescribeMetricFiltersInput, filter tfslices.Predicate[*awstypes.MetricFilter], optFns ...tfslices.FinderOptionsFunc) ([]awstypes.MetricFilter, error) {
	var output []awstypes.MetricFilter
	opts := tfslices.NewFinderOptions(optFns)

	pages := cloudwatchlogs.NewDescribeMetricFiltersPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError: err,
			}
		}

		if err != nil {
			return nil, err
		}

		for _, v := range page.MetricFilters {
			if filter(&v) {
				output = append(output, v)
				if opts.ReturnFirstMatch() {
					return output, nil
				}
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
		tfMap[names.AttrDefaultValue] = flex.Float64ToStringValue(v)
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

const metricFilterImportIDSeparator = ":"

func metricFilterParseImportID(id string) (string, string, error) {
	parts := strings.Split(id, metricFilterImportIDSeparator)

	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected log-group-name%[2]sfilter-name", id, metricFilterImportIDSeparator)
}

var (
	_ inttypes.SDKv2ImportID = metricFilterImportID{}
)

type metricFilterImportID struct{}

func (metricFilterImportID) Parse(id string) (string, map[string]any, error) {
	logGroupName, filterName, err := metricFilterParseImportID(id)
	if err != nil {
		return "", nil, err
	}

	result := map[string]any{
		names.AttrLogGroupName: logGroupName,
		names.AttrName:         filterName,
	}

	return filterName, result, nil
}

func (metricFilterImportID) Create(d *schema.ResourceData) string {
	return d.Get(names.AttrName).(string)
}
