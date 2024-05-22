// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package wafv2

import (
	"bytes"
	"context"
	"fmt"
	"log"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/wafv2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/wafv2/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_wafv2_web_acl_logging_configuration", name="Web ACL Logging Configuration")
func resourceWebACLLoggingConfiguration() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceWebACLLoggingConfigurationPut,
		ReadWithoutTimeout:   resourceWebACLLoggingConfigurationRead,
		UpdateWithoutTimeout: resourceWebACLLoggingConfigurationPut,
		DeleteWithoutTimeout: resourceWebACLLoggingConfigurationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				"log_destination_configs": {
					Type:     schema.TypeSet,
					Required: true,
					ForceNew: true,
					MinItems: 1,
					MaxItems: 100,
					Elem: &schema.Schema{
						Type:         schema.TypeString,
						ValidateFunc: verify.ValidARN,
					},
					Description: "AWS Kinesis Firehose Delivery Stream ARNs",
				},
				"logging_filter": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"default_behavior": {
								Type:             schema.TypeString,
								Required:         true,
								ValidateDiagFunc: enum.Validate[awstypes.FilterBehavior](),
							},
							names.AttrFilter: {
								Type:     schema.TypeSet,
								Required: true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"behavior": {
											Type:             schema.TypeString,
											Required:         true,
											ValidateDiagFunc: enum.Validate[awstypes.FilterBehavior](),
										},
										names.AttrCondition: {
											Type:     schema.TypeSet,
											Required: true,
											MinItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"action_condition": {
														Type:     schema.TypeList,
														Optional: true,
														MaxItems: 1,
														Elem: &schema.Resource{
															Schema: map[string]*schema.Schema{
																names.AttrAction: {
																	Type:             schema.TypeString,
																	Required:         true,
																	ValidateDiagFunc: enum.Validate[awstypes.ActionValue](),
																},
															},
														},
													},
													"label_name_condition": {
														Type:     schema.TypeList,
														Optional: true,
														MaxItems: 1,
														Elem: &schema.Resource{
															Schema: map[string]*schema.Schema{
																"label_name": {
																	Type:     schema.TypeString,
																	Required: true,
																	ValidateFunc: validation.All(
																		validation.StringLenBetween(1, 1024),
																		validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z_\-:]+$`), "must contain only alphanumeric characters, underscores, hyphens, and colons"),
																	),
																},
															},
														},
													},
												},
											},
										},
										"requirement": {
											Type:             schema.TypeString,
											Required:         true,
											ValidateDiagFunc: enum.Validate[awstypes.FilterRequirement](),
										},
									},
								},
							},
						},
					},
				},
				"redacted_fields": {
					// To allow this argument and its nested fields with Empty Schemas (e.g. "method")
					// to be correctly interpreted, this argument must be of type List,
					// otherwise, at apply-time a field configured as an empty block
					// (e.g. body {}) will result in a nil redacted_fields attribute
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 100,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"method":       emptySchema(),
							"query_string": emptySchema(),
							"single_header": {
								Type:     schema.TypeList,
								Optional: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										names.AttrName: {
											Type:     schema.TypeString,
											Required: true,
											ValidateFunc: validation.All(
												validation.StringLenBetween(1, 40),
												// The value is returned in lower case by the API.
												// Trying to solve it with StateFunc and/or DiffSuppressFunc resulted in hash problem of the rule field or didn't work.
												validation.StringMatch(regexache.MustCompile(`^[0-9a-z_-]+$`), "must contain only lowercase alphanumeric characters, underscores, and hyphens"),
											),
										},
									},
								},
							},
							"uri_path": emptySchema(),
						},
					},
					Description:      "Parts of the request to exclude from logs",
					DiffSuppressFunc: suppressRedactedFieldsDiff,
				},
				names.AttrResourceARN: {
					Type:         schema.TypeString,
					Required:     true,
					ForceNew:     true,
					ValidateFunc: verify.ValidARN,
					Description:  "AWS WebACL ARN",
				},
			}
		},
	}
}

func resourceWebACLLoggingConfigurationPut(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFV2Client(ctx)

	resourceARN := d.Get(names.AttrResourceARN).(string)
	config := &awstypes.LoggingConfiguration{
		LogDestinationConfigs: flex.ExpandStringValueSet(d.Get("log_destination_configs").(*schema.Set)),
		ResourceArn:           aws.String(resourceARN),
	}

	if v, ok := d.GetOk("logging_filter"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		config.LoggingFilter = expandLoggingFilter(v.([]interface{}))
	}

	if v, ok := d.GetOk("redacted_fields"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		config.RedactedFields = expandRedactedFields(v.([]interface{}))
	} else {
		config.RedactedFields = []awstypes.FieldToMatch{}
	}

	input := &wafv2.PutLoggingConfigurationInput{
		LoggingConfiguration: config,
	}

	output, err := conn.PutLoggingConfiguration(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "putting WAFv2 WebACL Logging Configuration (%s): %s", resourceARN, err)
	}

	if d.IsNewResource() {
		d.SetId(aws.ToString(output.LoggingConfiguration.ResourceArn))
	}

	return append(diags, resourceWebACLLoggingConfigurationRead(ctx, d, meta)...)
}

func resourceWebACLLoggingConfigurationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFV2Client(ctx)

	loggingConfig, err := findLoggingConfigurationByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] WAFv2 WebACL Logging Configuration (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading WAFv2 WebACL Logging Configuration (%s): %s", d.Id(), err)
	}

	if err := d.Set("log_destination_configs", flex.FlattenStringValueList(loggingConfig.LogDestinationConfigs)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting log_destination_configs: %s", err)
	}
	if err := d.Set("logging_filter", flattenLoggingFilter(loggingConfig.LoggingFilter)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting logging_filter: %s", err)
	}
	if err := d.Set("redacted_fields", flattenRedactedFields(loggingConfig.RedactedFields)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting redacted_fields: %s", err)
	}
	d.Set(names.AttrResourceARN, loggingConfig.ResourceArn)

	return diags
}

func resourceWebACLLoggingConfigurationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFV2Client(ctx)

	log.Printf("[INFO] Deleting WAFv2 WebACL Logging Configuration: %s", d.Id())
	_, err := conn.DeleteLoggingConfiguration(ctx, &wafv2.DeleteLoggingConfigurationInput{
		ResourceArn: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.WAFNonexistentItemException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting WAFv2 WebACL Logging Configuration (%s): %s", d.Id(), err)
	}

	return diags
}

func findLoggingConfigurationByARN(ctx context.Context, conn *wafv2.Client, arn string) (*awstypes.LoggingConfiguration, error) {
	input := &wafv2.GetLoggingConfigurationInput{
		ResourceArn: aws.String(arn),
	}

	output, err := conn.GetLoggingConfiguration(ctx, input)

	if errs.IsA[*awstypes.WAFNonexistentItemException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.LoggingConfiguration == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.LoggingConfiguration, nil
}

func expandLoggingFilter(l []interface{}) *awstypes.LoggingFilter {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})

	if !ok {
		return nil
	}

	loggingFilter := &awstypes.LoggingFilter{}

	if v, ok := tfMap["default_behavior"].(string); ok && v != "" {
		loggingFilter.DefaultBehavior = awstypes.FilterBehavior(v)
	}

	if v, ok := tfMap[names.AttrFilter].(*schema.Set); ok && v.Len() > 0 {
		loggingFilter.Filters = expandFilters(v.List())
	}

	return loggingFilter
}

func expandFilters(l []interface{}) []awstypes.Filter {
	if len(l) == 0 {
		return nil
	}

	var filters []awstypes.Filter

	for _, tfMapRaw := range l {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		filter := awstypes.Filter{}

		if v, ok := tfMap["behavior"].(string); ok && v != "" {
			filter.Behavior = awstypes.FilterBehavior(v)
		}

		if v, ok := tfMap[names.AttrCondition].(*schema.Set); ok && v.Len() > 0 {
			filter.Conditions = expandFilterConditions(v.List())
		}

		if v, ok := tfMap["requirement"].(string); ok && v != "" {
			filter.Requirement = awstypes.FilterRequirement(v)
		}

		filters = append(filters, filter)
	}

	return filters
}

func expandFilterConditions(l []interface{}) []awstypes.Condition {
	if len(l) == 0 {
		return nil
	}

	var conditions []awstypes.Condition

	for _, tfMapRaw := range l {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		condition := awstypes.Condition{}

		if v, ok := tfMap["action_condition"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
			condition.ActionCondition = expandActionCondition(v)
		}

		if v, ok := tfMap["label_name_condition"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
			condition.LabelNameCondition = expandLabelNameCondition(v)
		}

		conditions = append(conditions, condition)
	}

	return conditions
}

func expandActionCondition(l []interface{}) *awstypes.ActionCondition {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})
	if !ok {
		return nil
	}

	condition := &awstypes.ActionCondition{}

	if v, ok := tfMap[names.AttrAction].(string); ok && v != "" {
		condition.Action = awstypes.ActionValue(v)
	}

	return condition
}

func expandLabelNameCondition(l []interface{}) *awstypes.LabelNameCondition {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})
	if !ok {
		return nil
	}

	condition := &awstypes.LabelNameCondition{}

	if v, ok := tfMap["label_name"].(string); ok && v != "" {
		condition.LabelName = aws.String(v)
	}

	return condition
}

func expandRedactedFields(fields []interface{}) []awstypes.FieldToMatch {
	redactedFields := make([]awstypes.FieldToMatch, 0, len(fields))
	for _, field := range fields {
		redactedFields = append(redactedFields, expandRedactedField(field))
	}
	return redactedFields
}

func expandRedactedField(field interface{}) awstypes.FieldToMatch {
	m := field.(map[string]interface{})

	f := awstypes.FieldToMatch{}

	// While the FieldToMatch struct allows more than 1 of its fields to be set,
	// the WAFv2 API does not. In addition, in the context of Logging Configuration requests,
	// the WAFv2 API only supports the following redacted fields.
	// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/14244
	if v, ok := m["method"]; ok && len(v.([]interface{})) > 0 {
		f.Method = &awstypes.Method{}
	}

	if v, ok := m["query_string"]; ok && len(v.([]interface{})) > 0 {
		f.QueryString = &awstypes.QueryString{}
	}

	if v, ok := m["single_header"]; ok && len(v.([]interface{})) > 0 {
		f.SingleHeader = expandSingleHeader(m["single_header"].([]interface{}))
	}

	if v, ok := m["uri_path"]; ok && len(v.([]interface{})) > 0 {
		f.UriPath = &awstypes.UriPath{}
	}

	return f
}

func flattenLoggingFilter(filter *awstypes.LoggingFilter) []interface{} {
	if filter == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"default_behavior": string(filter.DefaultBehavior),
		names.AttrFilter:   flattenFilters(filter.Filters),
	}

	return []interface{}{m}
}

func flattenFilters(f []awstypes.Filter) []interface{} {
	if len(f) == 0 {
		return []interface{}{}
	}

	var filters []interface{}

	for _, filter := range f {
		m := map[string]interface{}{
			"behavior":          string(filter.Behavior),
			names.AttrCondition: flattenFilterConditions(filter.Conditions),
			"requirement":       string(filter.Requirement),
		}

		filters = append(filters, m)
	}

	return filters
}

func flattenFilterConditions(c []awstypes.Condition) []interface{} {
	if len(c) == 0 {
		return []interface{}{}
	}

	var conditions []interface{}

	for _, condition := range c {
		m := map[string]interface{}{
			"action_condition":     flattenActionCondition(condition.ActionCondition),
			"label_name_condition": flattenLabelNameCondition(condition.LabelNameCondition),
		}

		conditions = append(conditions, m)
	}

	return conditions
}

func flattenActionCondition(a *awstypes.ActionCondition) []interface{} {
	if a == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		names.AttrAction: string(a.Action),
	}

	return []interface{}{m}
}

func flattenLabelNameCondition(l *awstypes.LabelNameCondition) []interface{} {
	if l == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"label_name": aws.ToString(l.LabelName),
	}

	return []interface{}{m}
}

func flattenRedactedFields(fields []awstypes.FieldToMatch) []interface{} {
	redactedFields := make([]interface{}, 0, len(fields))
	for _, field := range fields {
		redactedFields = append(redactedFields, flattenRedactedField(field))
	}
	return redactedFields
}

func flattenRedactedField(f awstypes.FieldToMatch) map[string]interface{} {
	m := map[string]interface{}{}

	// In the context of Logging Configuration requests,
	// the WAFv2 API only supports the following redacted fields.
	// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/14244
	if f.Method != nil {
		m["method"] = make([]map[string]interface{}, 1)
	}

	if f.QueryString != nil {
		m["query_string"] = make([]map[string]interface{}, 1)
	}

	if f.SingleHeader != nil {
		m["single_header"] = flattenSingleHeader(f.SingleHeader)
	}

	if f.UriPath != nil {
		m["uri_path"] = make([]map[string]interface{}, 1)
	}

	return m
}

// redactedFieldsHash takes a map[string]interface{} as input and generates a
// unique hashcode, taking into account keys defined in the resource's schema
// are present even if not explicitly configured
func redactedFieldsHash(v interface{}) int {
	var buf bytes.Buffer
	m, ok := v.(map[string]interface{})
	if !ok {
		return 0
	}
	if v, ok := m["method"].([]interface{}); ok && len(v) > 0 {
		buf.WriteString("method-")
	}
	if v, ok := m["query_string"].([]interface{}); ok && len(v) > 0 {
		buf.WriteString("query_string-")
	}
	if v, ok := m["uri_path"].([]interface{}); ok && len(v) > 0 {
		buf.WriteString("uri_path-")
	}
	if v, ok := m["single_header"].([]interface{}); ok && len(v) > 0 {
		sh, ok := v[0].(map[string]interface{})
		if ok {
			if name, ok := sh[names.AttrName].(string); ok {
				buf.WriteString(fmt.Sprintf("%s-", name))
			}
		}
	}

	return create.StringHashcode(buf.String())
}

func suppressRedactedFieldsDiff(k, old, new string, d *schema.ResourceData) bool {
	o, n := d.GetChange("redacted_fields")
	oList := o.([]interface{})
	nList := n.([]interface{})

	if len(oList) == 0 && len(nList) == 0 {
		return true
	}

	if len(oList) == 0 && len(nList) != 0 {
		// account for empty block
		return nList[0] == nil
	}

	if len(oList) != 0 && len(nList) == 0 {
		// account for empty block
		return oList[0] == nil
	}

	oldSet := schema.NewSet(redactedFieldsHash, oList)
	newSet := schema.NewSet(redactedFieldsHash, nList)
	return oldSet.Equal(newSet)
}
