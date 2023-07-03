// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package wafv2

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/wafv2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKResource("aws_wafv2_web_acl_logging_configuration")
func ResourceWebACLLoggingConfiguration() *schema.Resource {
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
								Type:         schema.TypeString,
								Required:     true,
								ValidateFunc: validation.StringInSlice(wafv2.FilterBehavior_Values(), false),
							},
							"filter": {
								Type:     schema.TypeSet,
								Required: true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"behavior": {
											Type:         schema.TypeString,
											Required:     true,
											ValidateFunc: validation.StringInSlice(wafv2.FilterBehavior_Values(), false),
										},
										"condition": {
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
																"action": {
																	Type:         schema.TypeString,
																	Required:     true,
																	ValidateFunc: validation.StringInSlice(wafv2.ActionValue_Values(), false),
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
																		validation.StringMatch(regexp.MustCompile(`^[0-9A-Za-z_\-:]+$`), "must contain only alphanumeric characters, underscores, hyphens, and colons"),
																	),
																},
															},
														},
													},
												},
											},
										},
										"requirement": {
											Type:         schema.TypeString,
											Required:     true,
											ValidateFunc: validation.StringInSlice(wafv2.FilterRequirement_Values(), false),
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
										"name": {
											Type:     schema.TypeString,
											Required: true,
											ValidateFunc: validation.All(
												validation.StringLenBetween(1, 40),
												// The value is returned in lower case by the API.
												// Trying to solve it with StateFunc and/or DiffSuppressFunc resulted in hash problem of the rule field or didn't work.
												validation.StringMatch(regexp.MustCompile(`^[a-z0-9-_]+$`), "must contain only lowercase alphanumeric characters, underscores, and hyphens"),
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
				"resource_arn": {
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
	conn := meta.(*conns.AWSClient).WAFV2Conn(ctx)

	resourceARN := d.Get("resource_arn").(string)
	config := &wafv2.LoggingConfiguration{
		LogDestinationConfigs: flex.ExpandStringSet(d.Get("log_destination_configs").(*schema.Set)),
		ResourceArn:           aws.String(resourceARN),
	}

	if v, ok := d.GetOk("logging_filter"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		config.LoggingFilter = expandLoggingFilter(v.([]interface{}))
	}

	if v, ok := d.GetOk("redacted_fields"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		config.RedactedFields = expandRedactedFields(v.([]interface{}))
	} else {
		config.RedactedFields = []*wafv2.FieldToMatch{}
	}

	input := &wafv2.PutLoggingConfigurationInput{
		LoggingConfiguration: config,
	}

	output, err := conn.PutLoggingConfigurationWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "putting WAFv2 WebACL Logging Configuration (%s): %s", resourceARN, err)
	}

	if d.IsNewResource() {
		d.SetId(aws.StringValue(output.LoggingConfiguration.ResourceArn))
	}

	return append(diags, resourceWebACLLoggingConfigurationRead(ctx, d, meta)...)
}

func resourceWebACLLoggingConfigurationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFV2Conn(ctx)

	loggingConfig, err := FindLoggingConfigurationByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] WAFv2 WebACL Logging Configuration (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return diag.Errorf("reading WAFv2 WebACL Logging Configuration (%s): %s", d.Id(), err)
	}

	if err := d.Set("log_destination_configs", flex.FlattenStringList(loggingConfig.LogDestinationConfigs)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting log_destination_configs: %s", err)
	}
	if err := d.Set("logging_filter", flattenLoggingFilter(loggingConfig.LoggingFilter)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting logging_filter: %s", err)
	}
	if err := d.Set("redacted_fields", flattenRedactedFields(loggingConfig.RedactedFields)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting redacted_fields: %s", err)
	}
	d.Set("resource_arn", loggingConfig.ResourceArn)

	return diags
}

func resourceWebACLLoggingConfigurationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFV2Conn(ctx)

	log.Printf("[INFO] Deleting WAFv2 WebACL Logging Configuration: %s", d.Id())
	_, err := conn.DeleteLoggingConfigurationWithContext(ctx, &wafv2.DeleteLoggingConfigurationInput{
		ResourceArn: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, wafv2.ErrCodeWAFNonexistentItemException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting WAFv2 WebACL Logging Configuration (%s): %s", d.Id(), err)
	}

	return diags
}

func FindLoggingConfigurationByARN(ctx context.Context, conn *wafv2.WAFV2, arn string) (*wafv2.LoggingConfiguration, error) {
	input := &wafv2.GetLoggingConfigurationInput{
		ResourceArn: aws.String(arn),
	}

	output, err := conn.GetLoggingConfigurationWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, wafv2.ErrCodeWAFNonexistentItemException) {
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

func expandLoggingFilter(l []interface{}) *wafv2.LoggingFilter {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})

	if !ok {
		return nil
	}

	loggingFilter := &wafv2.LoggingFilter{}

	if v, ok := tfMap["default_behavior"].(string); ok && v != "" {
		loggingFilter.DefaultBehavior = aws.String(v)
	}

	if v, ok := tfMap["filter"].(*schema.Set); ok && v.Len() > 0 {
		loggingFilter.Filters = expandFilters(v.List())
	}

	return loggingFilter
}

func expandFilters(l []interface{}) []*wafv2.Filter {
	if len(l) == 0 {
		return nil
	}

	var filters []*wafv2.Filter

	for _, tfMapRaw := range l {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		filter := &wafv2.Filter{}

		if v, ok := tfMap["behavior"].(string); ok && v != "" {
			filter.Behavior = aws.String(v)
		}

		if v, ok := tfMap["condition"].(*schema.Set); ok && v.Len() > 0 {
			filter.Conditions = expandFilterConditions(v.List())
		}

		if v, ok := tfMap["requirement"].(string); ok && v != "" {
			filter.Requirement = aws.String(v)
		}

		filters = append(filters, filter)
	}

	return filters
}

func expandFilterConditions(l []interface{}) []*wafv2.Condition {
	if len(l) == 0 {
		return nil
	}

	var conditions []*wafv2.Condition

	for _, tfMapRaw := range l {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		condition := &wafv2.Condition{}

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

func expandActionCondition(l []interface{}) *wafv2.ActionCondition {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})
	if !ok {
		return nil
	}

	condition := &wafv2.ActionCondition{}

	if v, ok := tfMap["action"].(string); ok && v != "" {
		condition.Action = aws.String(v)
	}

	return condition
}

func expandLabelNameCondition(l []interface{}) *wafv2.LabelNameCondition {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})
	if !ok {
		return nil
	}

	condition := &wafv2.LabelNameCondition{}

	if v, ok := tfMap["label_name"].(string); ok && v != "" {
		condition.LabelName = aws.String(v)
	}

	return condition
}

func expandRedactedFields(fields []interface{}) []*wafv2.FieldToMatch {
	redactedFields := make([]*wafv2.FieldToMatch, 0, len(fields))
	for _, field := range fields {
		redactedFields = append(redactedFields, expandRedactedField(field))
	}
	return redactedFields
}

func expandRedactedField(field interface{}) *wafv2.FieldToMatch {
	m := field.(map[string]interface{})

	f := &wafv2.FieldToMatch{}

	// While the FieldToMatch struct allows more than 1 of its fields to be set,
	// the WAFv2 API does not. In addition, in the context of Logging Configuration requests,
	// the WAFv2 API only supports the following redacted fields.
	// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/14244
	if v, ok := m["method"]; ok && len(v.([]interface{})) > 0 {
		f.Method = &wafv2.Method{}
	}

	if v, ok := m["query_string"]; ok && len(v.([]interface{})) > 0 {
		f.QueryString = &wafv2.QueryString{}
	}

	if v, ok := m["single_header"]; ok && len(v.([]interface{})) > 0 {
		f.SingleHeader = expandSingleHeader(m["single_header"].([]interface{}))
	}

	if v, ok := m["uri_path"]; ok && len(v.([]interface{})) > 0 {
		f.UriPath = &wafv2.UriPath{}
	}

	return f
}

func flattenLoggingFilter(filter *wafv2.LoggingFilter) []interface{} {
	if filter == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"default_behavior": aws.StringValue(filter.DefaultBehavior),
		"filter":           flattenFilters(filter.Filters),
	}

	return []interface{}{m}
}

func flattenFilters(f []*wafv2.Filter) []interface{} {
	if len(f) == 0 {
		return []interface{}{}
	}

	var filters []interface{}

	for _, filter := range f {
		if filter == nil {
			continue
		}

		m := map[string]interface{}{
			"behavior":    aws.StringValue(filter.Behavior),
			"condition":   flattenFilterConditions(filter.Conditions),
			"requirement": aws.StringValue(filter.Requirement),
		}

		filters = append(filters, m)
	}

	return filters
}

func flattenFilterConditions(c []*wafv2.Condition) []interface{} {
	if len(c) == 0 {
		return []interface{}{}
	}

	var conditions []interface{}

	for _, condition := range c {
		if condition == nil {
			continue
		}

		m := map[string]interface{}{
			"action_condition":     flattenActionCondition(condition.ActionCondition),
			"label_name_condition": flattenLabelNameCondition(condition.LabelNameCondition),
		}

		conditions = append(conditions, m)
	}

	return conditions
}

func flattenActionCondition(a *wafv2.ActionCondition) []interface{} {
	if a == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"action": aws.StringValue(a.Action),
	}

	return []interface{}{m}
}

func flattenLabelNameCondition(l *wafv2.LabelNameCondition) []interface{} {
	if l == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"label_name": aws.StringValue(l.LabelName),
	}

	return []interface{}{m}
}

func flattenRedactedFields(fields []*wafv2.FieldToMatch) []interface{} {
	redactedFields := make([]interface{}, 0, len(fields))
	for _, field := range fields {
		redactedFields = append(redactedFields, flattenRedactedField(field))
	}
	return redactedFields
}

func flattenRedactedField(f *wafv2.FieldToMatch) map[string]interface{} {
	m := map[string]interface{}{}

	if f == nil {
		return m
	}

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
			if name, ok := sh["name"].(string); ok {
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
