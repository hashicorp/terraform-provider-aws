// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/securityhub/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	sdkschema "github.com/hashicorp/terraform-provider-aws/internal/sdkv2/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/sdkv2/types/nullable"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func InsightFiltersSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Required: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				names.AttrAWSAccountID:                                 stringFilterSchema(),
				"aws_account_name":                                     stringFilterSchema(),
				"company_name":                                         stringFilterSchema(),
				"compliance_associated_standards_id":                   stringFilterSchema(),
				"compliance_security_control_id":                       stringFilterSchema(),
				"compliance_security_control_parameters_name":          stringFilterSchema(),
				"compliance_security_control_parameters_value":         stringFilterSchema(),
				"compliance_status":                                    stringFilterSchema(),
				"confidence":                                           numberFilterSchema(),
				names.AttrCreatedAt:                                    dateFilterSchema(),
				"criticality":                                          numberFilterSchema(),
				names.AttrDescription:                                  stringFilterSchema(),
				"finding_provider_fields_confidence":                   numberFilterSchema(),
				"finding_provider_fields_criticality":                  numberFilterSchema(),
				"finding_provider_fields_related_findings_id":          stringFilterSchema(),
				"finding_provider_fields_related_findings_product_arn": stringFilterSchema(),
				"finding_provider_fields_severity_label":               stringFilterSchema(),
				"finding_provider_fields_severity_original":            stringFilterSchema(),
				"finding_provider_fields_types":                        stringFilterSchema(),
				"first_observed_at":                                    dateFilterSchema(),
				"generator_id":                                         stringFilterSchema(),
				names.AttrID:                                           stringFilterSchema(),
				"keyword":                                              keywordFilterSchema(),
				"last_observed_at":                                     dateFilterSchema(),
				"malware_name":                                         stringFilterSchema(),
				"malware_path":                                         stringFilterSchema(),
				"malware_state":                                        stringFilterSchema(),
				"malware_type":                                         stringFilterSchema(),
				"network_destination_domain":                           stringFilterSchema(),
				"network_destination_ipv4":                             ipFilterSchema(),
				"network_destination_ipv6":                             ipFilterSchema(),
				"network_destination_port":                             numberFilterSchema(),
				"network_direction":                                    stringFilterSchema(),
				"network_protocol":                                     stringFilterSchema(),
				"network_source_domain":                                stringFilterSchema(),
				"network_source_ipv4":                                  ipFilterSchema(),
				"network_source_ipv6":                                  ipFilterSchema(),
				"network_source_mac":                                   stringFilterSchema(),
				"network_source_port":                                  numberFilterSchema(),
				"note_text":                                            stringFilterSchema(),
				"note_updated_at":                                      dateFilterSchema(),
				"note_updated_by":                                      stringFilterSchema(),
				"process_launched_at":                                  dateFilterSchema(),
				"process_name":                                         stringFilterSchema(),
				"process_parent_pid":                                   numberFilterSchema(),
				"process_path":                                         stringFilterSchema(),
				"process_pid":                                          numberFilterSchema(),
				"process_terminated_at":                                dateFilterSchema(),
				"product_arn":                                          stringFilterSchema(),
				"product_fields":                                       mapFilterSchema(),
				"product_name":                                         stringFilterSchema(),
				"recommendation_text":                                  stringFilterSchema(),
				"record_state":                                         stringFilterSchema(),
				"related_findings_id":                                  stringFilterSchema(),
				"related_findings_product_arn":                         stringFilterSchema(),
				"resource_aws_ec2_instance_iam_instance_profile_arn": stringFilterSchema(),
				"resource_aws_ec2_instance_image_id":                 stringFilterSchema(),
				"resource_aws_ec2_instance_ipv4_addresses":           ipFilterSchema(),
				"resource_aws_ec2_instance_ipv6_addresses":           ipFilterSchema(),
				"resource_aws_ec2_instance_key_name":                 stringFilterSchema(),
				"resource_aws_ec2_instance_launched_at":              dateFilterSchema(),
				"resource_aws_ec2_instance_subnet_id":                stringFilterSchema(),
				"resource_aws_ec2_instance_type":                     stringFilterSchema(),
				"resource_aws_ec2_instance_vpc_id":                   stringFilterSchema(),
				"resource_aws_iam_access_key_created_at":             dateFilterSchema(),
				"resource_aws_iam_access_key_status":                 stringFilterSchema(),
				"resource_aws_iam_access_key_user_name":              stringFilterSchema(),
				"resource_aws_s3_bucket_owner_id":                    stringFilterSchema(),
				"resource_aws_s3_bucket_owner_name":                  stringFilterSchema(),
				"resource_container_image_id":                        stringFilterSchema(),
				"resource_container_image_name":                      stringFilterSchema(),
				"resource_container_launched_at":                     dateFilterSchema(),
				"resource_container_name":                            stringFilterSchema(),
				"resource_details_other":                             mapFilterSchema(),
				names.AttrResourceID:                                 stringFilterSchema(),
				"resource_partition":                                 stringFilterSchema(),
				"resource_region":                                    stringFilterSchema(),
				names.AttrResourceTags:                               mapFilterSchema(),
				names.AttrResourceType:                               stringFilterSchema(),
				"severity_label":                                     stringFilterSchema(),
				"source_url":                                         stringFilterSchema(),
				"threat_intel_indicator_category":                    stringFilterSchema(),
				"threat_intel_indicator_last_observed_at":            dateFilterSchema(),
				"threat_intel_indicator_source":                      stringFilterSchema(),
				"threat_intel_indicator_source_url":                  stringFilterSchema(),
				"threat_intel_indicator_type":                        stringFilterSchema(),
				"threat_intel_indicator_value":                       stringFilterSchema(),
				"title":                                              stringFilterSchema(),
				names.AttrType:                                       stringFilterSchema(),
				"updated_at":                                         dateFilterSchema(),
				"user_defined_values":                                mapFilterSchema(),
				"verification_state":                                 stringFilterSchema(),
				"workflow_status":                                    workflowStatusSchema(),
			},
		},
	}
}

var dateFilterSchema = sync.OnceValue(func() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		MaxItems: 20,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"date_range": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							names.AttrUnit: sdkschema.StringEnumSchema[types.DateRangeUnit](sdkschema.AttrRequired),
							names.AttrValue: {
								Type:     schema.TypeInt,
								Required: true,
							},
						},
					},
				},
				"end": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"start": {
					Type:     schema.TypeString,
					Optional: true,
				},
			},
		},
	}
})

var ipFilterSchema = sync.OnceValue(func() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		MaxItems: 20,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"cidr": {
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: verify.ValidCIDRNetworkAddress,
				},
			},
		},
	}
})

func keywordFilterSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		MaxItems: 20,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				names.AttrValue: {
					Type:     schema.TypeString,
					Required: true,
				},
			},
		},
	}
}

var mapFilterSchema = sync.OnceValue(func() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		MaxItems: 20,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"comparison": sdkschema.StringEnumSchema[types.MapFilterComparison](sdkschema.AttrRequired),
				names.AttrKey: {
					Type:     schema.TypeString,
					Required: true,
				},
				names.AttrValue: {
					Type:     schema.TypeString,
					Required: true,
				},
			},
		},
	}
})

var numberFilterSchema = sync.OnceValue(func() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		MaxItems: 20,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"eq": {
					Type:         nullable.TypeNullableFloat,
					Optional:     true,
					ValidateFunc: nullable.ValidateTypeStringNullableFloat,
				},
				"gte": {
					Type:         nullable.TypeNullableFloat,
					Optional:     true,
					ValidateFunc: nullable.ValidateTypeStringNullableFloat,
				},
				"lte": {
					Type:         nullable.TypeNullableFloat,
					Optional:     true,
					ValidateFunc: nullable.ValidateTypeStringNullableFloat,
				},
			},
		},
	}
})

var stringFilterSchema = sync.OnceValue(func() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		MaxItems: 20,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"comparison": sdkschema.StringEnumSchema[types.StringFilterComparison](sdkschema.AttrRequired),
				names.AttrValue: {
					Type:     schema.TypeString,
					Required: true,
				},
			},
		},
	}
})

func workflowStatusSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		MaxItems: 20,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"comparison":    sdkschema.StringEnumSchema[types.StringFilterComparison](sdkschema.AttrRequired),
				names.AttrValue: sdkschema.StringEnumSchema[types.WorkflowStatus](sdkschema.AttrRequired),
			},
		},
	}
}

func FlattenSecurityFindingFilters(apiObject *types.AwsSecurityFindingFilters) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{
		names.AttrAWSAccountID:                                 flattenStringFilters(apiObject.AwsAccountId),
		"aws_account_name":                                     flattenStringFilters(apiObject.AwsAccountName),
		"company_name":                                         flattenStringFilters(apiObject.CompanyName),
		"compliance_associated_standards_id":                   flattenStringFilters(apiObject.ComplianceAssociatedStandardsId),
		"compliance_security_control_id":                       flattenStringFilters(apiObject.ComplianceSecurityControlId),
		"compliance_security_control_parameters_name":          flattenStringFilters(apiObject.ComplianceSecurityControlParametersName),
		"compliance_security_control_parameters_value":         flattenStringFilters(apiObject.ComplianceSecurityControlParametersValue),
		"compliance_status":                                    flattenStringFilters(apiObject.ComplianceStatus),
		"confidence":                                           flattenNumberFilters(apiObject.Confidence),
		names.AttrCreatedAt:                                    flattenDateFilters(apiObject.CreatedAt),
		"criticality":                                          flattenNumberFilters(apiObject.Criticality),
		names.AttrDescription:                                  flattenStringFilters(apiObject.Description),
		"finding_provider_fields_confidence":                   flattenNumberFilters(apiObject.FindingProviderFieldsConfidence),
		"finding_provider_fields_criticality":                  flattenNumberFilters(apiObject.FindingProviderFieldsCriticality),
		"finding_provider_fields_related_findings_id":          flattenStringFilters(apiObject.FindingProviderFieldsRelatedFindingsId),
		"finding_provider_fields_related_findings_product_arn": flattenStringFilters(apiObject.FindingProviderFieldsRelatedFindingsProductArn),
		"finding_provider_fields_severity_label":               flattenStringFilters(apiObject.FindingProviderFieldsSeverityLabel),
		"finding_provider_fields_severity_original":            flattenStringFilters(apiObject.FindingProviderFieldsSeverityOriginal),
		"finding_provider_fields_types":                        flattenStringFilters(apiObject.FindingProviderFieldsTypes),
		"first_observed_at":                                    flattenDateFilters(apiObject.FirstObservedAt),
		"generator_id":                                         flattenStringFilters(apiObject.GeneratorId),
		names.AttrID:                                           flattenStringFilters(apiObject.Id),
		"keyword":                                              flattenKeywordFilters(apiObject.Keyword),
		"last_observed_at":                                     flattenDateFilters(apiObject.LastObservedAt),
		"malware_name":                                         flattenStringFilters(apiObject.MalwareName),
		"malware_path":                                         flattenStringFilters(apiObject.MalwarePath),
		"malware_state":                                        flattenStringFilters(apiObject.MalwareState),
		"malware_type":                                         flattenStringFilters(apiObject.MalwareType),
		"network_destination_domain":                           flattenStringFilters(apiObject.NetworkDestinationDomain),
		"network_destination_ipv4":                             flattenIPFilters(apiObject.NetworkDestinationIpV4),
		"network_destination_ipv6":                             flattenIPFilters(apiObject.NetworkDestinationIpV6),
		"network_destination_port":                             flattenNumberFilters(apiObject.NetworkDestinationPort),
		"network_direction":                                    flattenStringFilters(apiObject.NetworkDirection),
		"network_protocol":                                     flattenStringFilters(apiObject.NetworkProtocol),
		"network_source_domain":                                flattenStringFilters(apiObject.NetworkSourceDomain),
		"network_source_ipv4":                                  flattenIPFilters(apiObject.NetworkSourceIpV4),
		"network_source_ipv6":                                  flattenIPFilters(apiObject.NetworkSourceIpV6),
		"network_source_mac":                                   flattenStringFilters(apiObject.NetworkSourceMac),
		"network_source_port":                                  flattenNumberFilters(apiObject.NetworkSourcePort),
		"note_text":                                            flattenStringFilters(apiObject.NoteText),
		"note_updated_at":                                      flattenDateFilters(apiObject.NoteUpdatedAt),
		"note_updated_by":                                      flattenStringFilters(apiObject.NoteUpdatedBy),
		"process_launched_at":                                  flattenDateFilters(apiObject.ProcessLaunchedAt),
		"process_name":                                         flattenStringFilters(apiObject.ProcessName),
		"process_parent_pid":                                   flattenNumberFilters(apiObject.ProcessParentPid),
		"process_path":                                         flattenStringFilters(apiObject.ProcessPath),
		"process_pid":                                          flattenNumberFilters(apiObject.ProcessPid),
		"process_terminated_at":                                flattenDateFilters(apiObject.ProcessTerminatedAt),
		"product_arn":                                          flattenStringFilters(apiObject.ProductArn),
		"product_fields":                                       flattenMapFilters(apiObject.ProductFields),
		"product_name":                                         flattenStringFilters(apiObject.ProductName),
		"recommendation_text":                                  flattenStringFilters(apiObject.RecommendationText),
		"record_state":                                         flattenStringFilters(apiObject.RecordState),
		"related_findings_id":                                  flattenStringFilters(apiObject.RelatedFindingsId),
		"related_findings_product_arn":                         flattenStringFilters(apiObject.RelatedFindingsProductArn),
		"resource_aws_ec2_instance_iam_instance_profile_arn": flattenStringFilters(apiObject.ResourceAwsEc2InstanceIamInstanceProfileArn),
		"resource_aws_ec2_instance_image_id":                 flattenStringFilters(apiObject.ResourceAwsEc2InstanceImageId),
		"resource_aws_ec2_instance_ipv4_addresses":           flattenIPFilters(apiObject.ResourceAwsEc2InstanceIpV4Addresses),
		"resource_aws_ec2_instance_ipv6_addresses":           flattenIPFilters(apiObject.ResourceAwsEc2InstanceIpV6Addresses),
		"resource_aws_ec2_instance_key_name":                 flattenStringFilters(apiObject.ResourceAwsEc2InstanceKeyName),
		"resource_aws_ec2_instance_launched_at":              flattenDateFilters(apiObject.ResourceAwsEc2InstanceLaunchedAt),
		"resource_aws_ec2_instance_subnet_id":                flattenStringFilters(apiObject.ResourceAwsEc2InstanceSubnetId),
		"resource_aws_ec2_instance_type":                     flattenStringFilters(apiObject.ResourceAwsEc2InstanceType),
		"resource_aws_ec2_instance_vpc_id":                   flattenStringFilters(apiObject.ResourceAwsEc2InstanceVpcId),
		"resource_aws_iam_access_key_created_at":             flattenDateFilters(apiObject.ResourceAwsIamAccessKeyCreatedAt),
		"resource_aws_iam_access_key_status":                 flattenStringFilters(apiObject.ResourceAwsIamAccessKeyStatus),
		"resource_aws_iam_access_key_user_name":              flattenStringFilters(apiObject.ResourceAwsIamAccessKeyUserName),
		"resource_aws_s3_bucket_owner_id":                    flattenStringFilters(apiObject.ResourceAwsS3BucketOwnerId),
		"resource_aws_s3_bucket_owner_name":                  flattenStringFilters(apiObject.ResourceAwsS3BucketOwnerName),
		"resource_container_image_id":                        flattenStringFilters(apiObject.ResourceContainerImageId),
		"resource_container_image_name":                      flattenStringFilters(apiObject.ResourceContainerImageName),
		"resource_container_launched_at":                     flattenDateFilters(apiObject.ResourceContainerLaunchedAt),
		"resource_container_name":                            flattenStringFilters(apiObject.ResourceContainerName),
		"resource_details_other":                             flattenMapFilters(apiObject.ResourceDetailsOther),
		names.AttrResourceID:                                 flattenStringFilters(apiObject.ResourceId),
		"resource_partition":                                 flattenStringFilters(apiObject.ResourcePartition),
		"resource_region":                                    flattenStringFilters(apiObject.ResourceRegion),
		names.AttrResourceTags:                               flattenMapFilters(apiObject.ResourceTags),
		names.AttrResourceType:                               flattenStringFilters(apiObject.ResourceType),
		"severity_label":                                     flattenStringFilters(apiObject.SeverityLabel),
		"source_url":                                         flattenStringFilters(apiObject.ThreatIntelIndicatorSourceUrl),
		"threat_intel_indicator_category":                    flattenStringFilters(apiObject.ThreatIntelIndicatorCategory),
		"threat_intel_indicator_last_observed_at":            flattenDateFilters(apiObject.ThreatIntelIndicatorLastObservedAt),
		"threat_intel_indicator_source":                      flattenStringFilters(apiObject.ThreatIntelIndicatorSource),
		"threat_intel_indicator_source_url":                  flattenStringFilters(apiObject.ThreatIntelIndicatorSourceUrl),
		"threat_intel_indicator_type":                        flattenStringFilters(apiObject.ThreatIntelIndicatorType),
		"threat_intel_indicator_value":                       flattenStringFilters(apiObject.ThreatIntelIndicatorValue),
		"title":                                              flattenStringFilters(apiObject.Title),
		names.AttrType:                                       flattenStringFilters(apiObject.Type),
		"updated_at":                                         flattenDateFilters(apiObject.UpdatedAt),
		"user_defined_values":                                flattenMapFilters(apiObject.UserDefinedFields),
		"verification_state":                                 flattenStringFilters(apiObject.VerificationState),
		"workflow_status":                                    flattenStringFilters(apiObject.WorkflowStatus),
	}

	return []any{tfMap}
}

func flattenDateRange(apiObject *types.DateRange) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{
		names.AttrUnit:  apiObject.Unit,
		names.AttrValue: aws.ToInt32((apiObject.Value)),
	}

	return []any{tfMap}
}

func flattenDateFilters(apiObjects []types.DateFilter) []any {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []any

	for _, apiObject := range apiObjects {
		tfMap := map[string]any{
			"date_range": flattenDateRange(apiObject.DateRange),
			"end":        aws.ToString(apiObject.End),
			"start":      aws.ToString(apiObject.Start),
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenIPFilters(apiObjects []types.IpFilter) []any {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []any

	for _, apiObject := range apiObjects {
		tfMap := map[string]any{
			"cidr": aws.ToString(apiObject.Cidr),
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenKeywordFilters(apiObjects []types.KeywordFilter) []any {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []any

	for _, apiObject := range apiObjects {
		tfMap := map[string]any{
			names.AttrValue: aws.ToString(apiObject.Value),
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenMapFilters(apiObjects []types.MapFilter) []any {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []any

	for _, apiObject := range apiObjects {
		tfMap := map[string]any{
			"comparison":    apiObject.Comparison,
			names.AttrKey:   aws.ToString(apiObject.Key),
			names.AttrValue: aws.ToString(apiObject.Value),
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenNumberFilters(apiObjects []types.NumberFilter) []any {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []any

	for _, apiObject := range apiObjects {
		tfMap := map[string]any{}

		if apiObject.Eq != nil {
			tfMap["eq"] = flex.Float64ToStringValue(apiObject.Eq)
		}

		if apiObject.Gte != nil {
			tfMap["gte"] = flex.Float64ToStringValue(apiObject.Gte)
		}

		if apiObject.Lte != nil {
			tfMap["lte"] = flex.Float64ToStringValue(apiObject.Lte)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenStringFilters(apiObjects []types.StringFilter) []any {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []any

	for _, apiObject := range apiObjects {
		tfMap := map[string]any{
			"comparison":    apiObject.Comparison,
			names.AttrValue: aws.ToString(apiObject.Value),
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func ExpandSecurityFindingFilters(tfList []any) *types.AwsSecurityFindingFilters {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &types.AwsSecurityFindingFilters{}

	if v, ok := tfMap[names.AttrAWSAccountID].(*schema.Set); ok && v.Len() > 0 {
		apiObject.AwsAccountId = expandStringFilters(v.List())
	}

	if v, ok := tfMap["aws_account_name"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.AwsAccountName = expandStringFilters(v.List())
	}

	if v, ok := tfMap["company_name"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.CompanyName = expandStringFilters(v.List())
	}

	if v, ok := tfMap["compliance_associated_standards_id"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.ComplianceAssociatedStandardsId = expandStringFilters(v.List())
	}

	if v, ok := tfMap["compliance_security_control_id"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.ComplianceSecurityControlId = expandStringFilters(v.List())
	}

	if v, ok := tfMap["compliance_security_control_parameters_name"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.ComplianceSecurityControlParametersName = expandStringFilters(v.List())
	}

	if v, ok := tfMap["compliance_security_control_parameters_value"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.ComplianceSecurityControlParametersValue = expandStringFilters(v.List())
	}

	if v, ok := tfMap["compliance_status"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.ComplianceStatus = expandStringFilters(v.List())
	}

	if v, ok := tfMap["confidence"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.Confidence = expandNumberFilters(v.List())
	}

	if v, ok := tfMap[names.AttrCreatedAt].(*schema.Set); ok && v.Len() > 0 {
		apiObject.CreatedAt = expandDateFilters(v.List())
	}

	if v, ok := tfMap["criticality"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.Criticality = expandNumberFilters(v.List())
	}

	if v, ok := tfMap[names.AttrDescription].(*schema.Set); ok && v.Len() > 0 {
		apiObject.Description = expandStringFilters(v.List())
	}

	if v, ok := tfMap["finding_provider_fields_confidence"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.FindingProviderFieldsConfidence = expandNumberFilters(v.List())
	}

	if v, ok := tfMap["finding_provider_fields_criticality"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.FindingProviderFieldsCriticality = expandNumberFilters(v.List())
	}

	if v, ok := tfMap["finding_provider_fields_related_findings_id"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.FindingProviderFieldsRelatedFindingsId = expandStringFilters(v.List())
	}

	if v, ok := tfMap["finding_provider_fields_related_findings_product_arn"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.FindingProviderFieldsRelatedFindingsProductArn = expandStringFilters(v.List())
	}

	if v, ok := tfMap["finding_provider_fields_severity_label"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.FindingProviderFieldsSeverityLabel = expandStringFilters(v.List())
	}

	if v, ok := tfMap["finding_provider_fields_severity_original"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.FindingProviderFieldsSeverityOriginal = expandStringFilters(v.List())
	}

	if v, ok := tfMap["finding_provider_fields_types"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.FindingProviderFieldsTypes = expandStringFilters(v.List())
	}

	if v, ok := tfMap["first_observed_at"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.FirstObservedAt = expandDateFilters(v.List())
	}

	if v, ok := tfMap["generator_id"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.GeneratorId = expandStringFilters(v.List())
	}

	if v, ok := tfMap[names.AttrID].(*schema.Set); ok && v.Len() > 0 {
		apiObject.Id = expandStringFilters(v.List())
	}

	if v, ok := tfMap["keyword"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.Keyword = expandKeywordFilters(v.List())
	}

	if v, ok := tfMap["last_observed_at"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.LastObservedAt = expandDateFilters(v.List())
	}

	if v, ok := tfMap["malware_name"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.MalwareName = expandStringFilters(v.List())
	}

	if v, ok := tfMap["malware_path"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.MalwarePath = expandStringFilters(v.List())
	}

	if v, ok := tfMap["malware_state"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.MalwareState = expandStringFilters(v.List())
	}

	if v, ok := tfMap["malware_type"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.MalwareType = expandStringFilters(v.List())
	}

	if v, ok := tfMap["network_destination_domain"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.NetworkDestinationDomain = expandStringFilters(v.List())
	}

	if v, ok := tfMap["network_destination_ipv4"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.NetworkDestinationIpV4 = expandIPFilters(v.List())
	}

	if v, ok := tfMap["network_destination_ipv6"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.NetworkDestinationIpV6 = expandIPFilters(v.List())
	}

	if v, ok := tfMap["network_destination_port"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.NetworkDestinationPort = expandNumberFilters(v.List())
	}

	if v, ok := tfMap["network_direction"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.NetworkDirection = expandStringFilters(v.List())
	}

	if v, ok := tfMap["network_protocol"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.NetworkProtocol = expandStringFilters(v.List())
	}

	if v, ok := tfMap["network_source_domain"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.NetworkSourceDomain = expandStringFilters(v.List())
	}

	if v, ok := tfMap["network_source_ipv4"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.NetworkSourceIpV4 = expandIPFilters(v.List())
	}

	if v, ok := tfMap["network_source_ipv6"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.NetworkSourceIpV6 = expandIPFilters(v.List())
	}

	if v, ok := tfMap["network_source_mac"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.NetworkSourceMac = expandStringFilters(v.List())
	}

	if v, ok := tfMap["network_source_port"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.NetworkSourcePort = expandNumberFilters(v.List())
	}

	if v, ok := tfMap["note_text"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.NoteText = expandStringFilters(v.List())
	}

	if v, ok := tfMap["note_updated_at"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.NoteUpdatedAt = expandDateFilters(v.List())
	}

	if v, ok := tfMap["note_updated_by"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.NoteUpdatedBy = expandStringFilters(v.List())
	}

	if v, ok := tfMap["process_launched_at"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.ProcessLaunchedAt = expandDateFilters(v.List())
	}

	if v, ok := tfMap["process_name"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.ProcessName = expandStringFilters(v.List())
	}

	if v, ok := tfMap["process_parent_pid"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.ProcessParentPid = expandNumberFilters(v.List())
	}

	if v, ok := tfMap["process_path"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.ProcessPath = expandStringFilters(v.List())
	}

	if v, ok := tfMap["process_pid"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.ProcessPid = expandNumberFilters(v.List())
	}

	if v, ok := tfMap["process_terminated_at"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.ProcessTerminatedAt = expandDateFilters(v.List())
	}

	if v, ok := tfMap["product_arn"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.ProductArn = expandStringFilters(v.List())
	}

	if v, ok := tfMap["product_fields"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.ProductFields = expandMapFilters(v.List())
	}

	if v, ok := tfMap["product_name"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.ProductName = expandStringFilters(v.List())
	}

	if v, ok := tfMap["recommendation_text"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.RecommendationText = expandStringFilters(v.List())
	}

	if v, ok := tfMap["record_state"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.RecordState = expandStringFilters(v.List())
	}

	if v, ok := tfMap["related_findings_id"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.RelatedFindingsId = expandStringFilters(v.List())
	}

	if v, ok := tfMap["related_findings_product_arn"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.RelatedFindingsProductArn = expandStringFilters(v.List())
	}

	if v, ok := tfMap["resource_aws_ec2_instance_iam_instance_profile_arn"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.ResourceAwsEc2InstanceIamInstanceProfileArn = expandStringFilters(v.List())
	}

	if v, ok := tfMap["resource_aws_ec2_instance_image_id"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.ResourceAwsEc2InstanceImageId = expandStringFilters(v.List())
	}

	if v, ok := tfMap["resource_aws_ec2_instance_ipv4_addresses"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.ResourceAwsEc2InstanceIpV4Addresses = expandIPFilters(v.List())
	}

	if v, ok := tfMap["resource_aws_ec2_instance_ipv6_addresses"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.ResourceAwsEc2InstanceIpV6Addresses = expandIPFilters(v.List())
	}

	if v, ok := tfMap["resource_aws_ec2_instance_key_name"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.ResourceAwsEc2InstanceKeyName = expandStringFilters(v.List())
	}

	if v, ok := tfMap["resource_aws_ec2_instance_launched_at"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.ResourceAwsEc2InstanceLaunchedAt = expandDateFilters(v.List())
	}

	if v, ok := tfMap["resource_aws_ec2_instance_subnet_id"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.ResourceAwsEc2InstanceSubnetId = expandStringFilters(v.List())
	}

	if v, ok := tfMap["resource_aws_ec2_instance_type"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.ResourceAwsEc2InstanceType = expandStringFilters(v.List())
	}

	if v, ok := tfMap["resource_aws_ec2_instance_vpc_id"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.ResourceAwsEc2InstanceVpcId = expandStringFilters(v.List())
	}

	if v, ok := tfMap["resource_aws_iam_access_key_created_at"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.ResourceAwsIamAccessKeyCreatedAt = expandDateFilters(v.List())
	}

	if v, ok := tfMap["resource_aws_iam_access_key_status"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.ResourceAwsIamAccessKeyStatus = expandStringFilters(v.List())
	}

	if v, ok := tfMap["resource_aws_iam_access_key_user_name"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.ResourceAwsIamAccessKeyUserName = expandStringFilters(v.List())
	}

	if v, ok := tfMap["resource_aws_s3_bucket_owner_id"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.ResourceAwsS3BucketOwnerId = expandStringFilters(v.List())
	}

	if v, ok := tfMap["resource_aws_s3_bucket_owner_name"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.ResourceAwsS3BucketOwnerName = expandStringFilters(v.List())
	}

	if v, ok := tfMap["resource_container_image_id"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.ResourceContainerImageId = expandStringFilters(v.List())
	}

	if v, ok := tfMap["resource_container_image_name"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.ResourceContainerImageName = expandStringFilters(v.List())
	}

	if v, ok := tfMap["resource_container_launched_at"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.ResourceContainerLaunchedAt = expandDateFilters(v.List())
	}

	if v, ok := tfMap["resource_container_name"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.ResourceContainerName = expandStringFilters(v.List())
	}

	if v, ok := tfMap["resource_details_other"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.ResourceDetailsOther = expandMapFilters(v.List())
	}

	if v, ok := tfMap[names.AttrResourceID].(*schema.Set); ok && v.Len() > 0 {
		apiObject.ResourceId = expandStringFilters(v.List())
	}

	if v, ok := tfMap["resource_partition"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.ResourcePartition = expandStringFilters(v.List())
	}

	if v, ok := tfMap["resource_region"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.ResourceRegion = expandStringFilters(v.List())
	}

	if v, ok := tfMap[names.AttrResourceTags].(*schema.Set); ok && v.Len() > 0 {
		apiObject.ResourceTags = expandMapFilters(v.List())
	}

	if v, ok := tfMap[names.AttrResourceType].(*schema.Set); ok && v.Len() > 0 {
		apiObject.ResourceType = expandStringFilters(v.List())
	}

	if v, ok := tfMap["severity_label"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.SeverityLabel = expandStringFilters(v.List())
	}

	if v, ok := tfMap["source_url"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.SourceUrl = expandStringFilters(v.List())
	}

	if v, ok := tfMap["threat_intel_indicator_category"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.ThreatIntelIndicatorCategory = expandStringFilters(v.List())
	}

	if v, ok := tfMap["threat_intel_indicator_last_observed_at"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.ThreatIntelIndicatorLastObservedAt = expandDateFilters(v.List())
	}

	if v, ok := tfMap["threat_intel_indicator_source"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.ThreatIntelIndicatorSource = expandStringFilters(v.List())
	}

	if v, ok := tfMap["threat_intel_indicator_source_url"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.ThreatIntelIndicatorSourceUrl = expandStringFilters(v.List())
	}

	if v, ok := tfMap["threat_intel_indicator_type"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.ThreatIntelIndicatorType = expandStringFilters(v.List())
	}

	if v, ok := tfMap["threat_intel_indicator_value"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.ThreatIntelIndicatorValue = expandStringFilters(v.List())
	}

	if v, ok := tfMap["title"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.Title = expandStringFilters(v.List())
	}

	if v, ok := tfMap[names.AttrType].(*schema.Set); ok && v.Len() > 0 {
		apiObject.Type = expandStringFilters(v.List())
	}

	if v, ok := tfMap["updated_at"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.UpdatedAt = expandDateFilters(v.List())
	}

	if v, ok := tfMap["user_defined_values"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.UserDefinedFields = expandMapFilters(v.List())
	}

	if v, ok := tfMap["verification_state"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.VerificationState = expandStringFilters(v.List())
	}

	if v, ok := tfMap["workflow_status"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.WorkflowStatus = expandStringFilters(v.List())
	}

	return apiObject
}

func expandDateRange(tfList []any) *types.DateRange {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]any)
	if !ok {
		return nil
	}

	apiObject := &types.DateRange{}

	if v, ok := tfMap[names.AttrUnit].(string); ok && v != "" {
		apiObject.Unit = types.DateRangeUnit(v)
	}

	if v, ok := tfMap[names.AttrValue].(int); ok {
		apiObject.Value = aws.Int32(int32(v))
	}

	return apiObject
}

func expandDateFilters(tfList []any) []types.DateFilter {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	var apiObjects []types.DateFilter

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		apiObject := types.DateFilter{}

		if v, ok := tfMap["date_range"].([]any); ok && len(v) > 0 && v[0] != nil {
			apiObject.DateRange = expandDateRange(v)
		}

		if v, ok := tfMap["end"].(string); ok && v != "" {
			apiObject.End = aws.String(v)
		}

		if v, ok := tfMap["start"].(string); ok && v != "" {
			apiObject.Start = aws.String(v)
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandIPFilters(tfList []any) []types.IpFilter {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	var apiObjects []types.IpFilter

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		apiObject := types.IpFilter{}

		if v, ok := tfMap["cidr"].(string); ok && v != "" {
			apiObject.Cidr = aws.String(v)
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandKeywordFilters(tfList []any) []types.KeywordFilter {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	var apiObjects []types.KeywordFilter

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		apiObject := types.KeywordFilter{}

		if v, ok := tfMap[names.AttrValue].(string); ok && v != "" {
			apiObject.Value = aws.String(v)
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandMapFilters(tfList []any) []types.MapFilter {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	var apiObjects []types.MapFilter

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		apiObject := types.MapFilter{}

		if v, ok := tfMap["comparison"].(string); ok && v != "" {
			apiObject.Comparison = types.MapFilterComparison(v)
		}

		if v, ok := tfMap[names.AttrKey].(string); ok && v != "" {
			apiObject.Key = aws.String(v)
		}

		if v, ok := tfMap[names.AttrValue].(string); ok && v != "" {
			apiObject.Value = aws.String(v)
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandNumberFilters(tfList []any) []types.NumberFilter {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	var apiObjects []types.NumberFilter

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		apiObject := types.NumberFilter{}

		if v, ok := tfMap["eq"].(string); ok {
			if v, null, _ := nullable.Float(v).ValueFloat64(); !null {
				apiObject.Eq = aws.Float64(v)
			}
		}

		if v, ok := tfMap["gte"].(string); ok {
			if v, null, _ := nullable.Float(v).ValueFloat64(); !null {
				apiObject.Gte = aws.Float64(v)
			}
		}

		if v, ok := tfMap["lte"].(string); ok {
			if v, null, _ := nullable.Float(v).ValueFloat64(); !null {
				apiObject.Lte = aws.Float64(v)
			}
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandStringFilters(tfList []any) []types.StringFilter {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	var apiObjects []types.StringFilter

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		apiObject := types.StringFilter{}

		if v, ok := tfMap["comparison"].(string); ok && v != "" {
			apiObject.Comparison = types.StringFilterComparison(v)
		}

		if v, ok := tfMap[names.AttrValue].(string); ok && v != "" {
			apiObject.Value = aws.String(v)
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}
