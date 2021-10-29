package config

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/configservice"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
)

func expandAccountAggregationSources(configured []interface{}) []*configservice.AccountAggregationSource {
	var results []*configservice.AccountAggregationSource
	for _, item := range configured {
		detail := item.(map[string]interface{})
		source := configservice.AccountAggregationSource{
			AllAwsRegions: aws.Bool(detail["all_regions"].(bool)),
		}

		if v, ok := detail["account_ids"]; ok {
			accountIDs := v.([]interface{})
			if len(accountIDs) > 0 {
				source.AccountIds = flex.ExpandStringList(accountIDs)
			}
		}

		if v, ok := detail["regions"]; ok {
			regions := v.([]interface{})
			if len(regions) > 0 {
				source.AwsRegions = flex.ExpandStringList(regions)
			}
		}

		results = append(results, &source)
	}
	return results
}

func expandOrganizationAggregationSource(configured map[string]interface{}) *configservice.OrganizationAggregationSource {
	source := configservice.OrganizationAggregationSource{
		AllAwsRegions: aws.Bool(configured["all_regions"].(bool)),
		RoleArn:       aws.String(configured["role_arn"].(string)),
	}

	if v, ok := configured["regions"]; ok {
		regions := v.([]interface{})
		if len(regions) > 0 {
			source.AwsRegions = flex.ExpandStringList(regions)
		}
	}

	return &source
}

func expandRecordingGroup(configured []interface{}) *configservice.RecordingGroup {
	recordingGroup := configservice.RecordingGroup{}
	group := configured[0].(map[string]interface{})

	if v, ok := group["all_supported"]; ok {
		recordingGroup.AllSupported = aws.Bool(v.(bool))
	}

	if v, ok := group["include_global_resource_types"]; ok {
		recordingGroup.IncludeGlobalResourceTypes = aws.Bool(v.(bool))
	}

	if v, ok := group["resource_types"]; ok {
		recordingGroup.ResourceTypes = flex.ExpandStringSet(v.(*schema.Set))
	}
	return &recordingGroup
}

func expandRuleScope(l []interface{}) *configservice.Scope {
	if len(l) == 0 || l[0] == nil {
		return nil
	}
	configured := l[0].(map[string]interface{})
	scope := &configservice.Scope{}

	if v, ok := configured["compliance_resource_id"].(string); ok && v != "" {
		scope.ComplianceResourceId = aws.String(v)
	}
	if v, ok := configured["compliance_resource_types"]; ok {
		l := v.(*schema.Set)
		if l.Len() > 0 {
			scope.ComplianceResourceTypes = flex.ExpandStringSet(l)
		}
	}
	if v, ok := configured["tag_key"].(string); ok && v != "" {
		scope.TagKey = aws.String(v)
	}
	if v, ok := configured["tag_value"].(string); ok && v != "" {
		scope.TagValue = aws.String(v)
	}

	return scope
}

func expandRuleSource(configured []interface{}) *configservice.Source {
	cfg := configured[0].(map[string]interface{})
	source := configservice.Source{
		Owner:            aws.String(cfg["owner"].(string)),
		SourceIdentifier: aws.String(cfg["source_identifier"].(string)),
	}
	if details, ok := cfg["source_detail"]; ok {
		source.SourceDetails = expandRuleSourceDetails(details.(*schema.Set))
	}
	return &source
}

func expandRuleSourceDetails(configured *schema.Set) []*configservice.SourceDetail {
	var results []*configservice.SourceDetail

	for _, item := range configured.List() {
		detail := item.(map[string]interface{})
		src := configservice.SourceDetail{}

		if msgType, ok := detail["message_type"].(string); ok && msgType != "" {
			src.MessageType = aws.String(msgType)
		}
		if eventSource, ok := detail["event_source"].(string); ok && eventSource != "" {
			src.EventSource = aws.String(eventSource)
		}
		if maxExecFreq, ok := detail["maximum_execution_frequency"].(string); ok && maxExecFreq != "" {
			src.MaximumExecutionFrequency = aws.String(maxExecFreq)
		}

		results = append(results, &src)
	}

	return results
}

func flattenAccountAggregationSources(sources []*configservice.AccountAggregationSource) []interface{} {
	var result []interface{}

	if len(sources) == 0 {
		return result
	}

	source := sources[0]
	m := make(map[string]interface{})
	m["account_ids"] = flex.FlattenStringList(source.AccountIds)
	m["all_regions"] = aws.BoolValue(source.AllAwsRegions)
	m["regions"] = flex.FlattenStringList(source.AwsRegions)
	result = append(result, m)
	return result
}

func flattenOrganizationAggregationSource(source *configservice.OrganizationAggregationSource) []interface{} {
	var result []interface{}

	if source == nil {
		return result
	}

	m := make(map[string]interface{})
	m["all_regions"] = aws.BoolValue(source.AllAwsRegions)
	m["regions"] = flex.FlattenStringList(source.AwsRegions)
	m["role_arn"] = aws.StringValue(source.RoleArn)
	result = append(result, m)
	return result
}

func flattenRecordingGroup(g *configservice.RecordingGroup) []map[string]interface{} {
	m := make(map[string]interface{}, 1)

	if g.AllSupported != nil {
		m["all_supported"] = *g.AllSupported
	}

	if g.IncludeGlobalResourceTypes != nil {
		m["include_global_resource_types"] = *g.IncludeGlobalResourceTypes
	}

	if g.ResourceTypes != nil && len(g.ResourceTypes) > 0 {
		m["resource_types"] = flex.FlattenStringSet(g.ResourceTypes)
	}

	return []map[string]interface{}{m}
}

func flattenRuleScope(scope *configservice.Scope) []interface{} {
	var items []interface{}

	m := make(map[string]interface{})
	if scope.ComplianceResourceId != nil {
		m["compliance_resource_id"] = *scope.ComplianceResourceId
	}
	if scope.ComplianceResourceTypes != nil {
		m["compliance_resource_types"] = flex.FlattenStringSet(scope.ComplianceResourceTypes)
	}
	if scope.TagKey != nil {
		m["tag_key"] = *scope.TagKey
	}
	if scope.TagValue != nil {
		m["tag_value"] = *scope.TagValue
	}

	items = append(items, m)
	return items
}

func flattenRuleSource(source *configservice.Source) []interface{} {
	var result []interface{}
	m := make(map[string]interface{})
	m["owner"] = *source.Owner
	m["source_identifier"] = *source.SourceIdentifier
	if len(source.SourceDetails) > 0 {
		m["source_detail"] = schema.NewSet(configRuleSourceDetailsHash, flattenRuleSourceDetails(source.SourceDetails))
	}
	result = append(result, m)
	return result
}

func flattenRuleSourceDetails(details []*configservice.SourceDetail) []interface{} {
	var items []interface{}
	for _, d := range details {
		m := make(map[string]interface{})
		if d.MessageType != nil {
			m["message_type"] = *d.MessageType
		}
		if d.EventSource != nil {
			m["event_source"] = *d.EventSource
		}
		if d.MaximumExecutionFrequency != nil {
			m["maximum_execution_frequency"] = *d.MaximumExecutionFrequency
		}

		items = append(items, m)
	}

	return items
}

func flattenSnapshotDeliveryProperties(p *configservice.ConfigSnapshotDeliveryProperties) []map[string]interface{} {
	m := make(map[string]interface{})

	if p.DeliveryFrequency != nil {
		m["delivery_frequency"] = *p.DeliveryFrequency
	}

	return []map[string]interface{}{m}
}
