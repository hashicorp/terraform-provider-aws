// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package securityhub

import (
	"context"
	"log"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/securityhub"
	"github.com/aws/aws-sdk-go-v2/service/securityhub/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_securityhub_insight", name="Insight")
func resourceInsight() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceInsightCreate,
		ReadWithoutTimeout:   resourceInsightRead,
		UpdateWithoutTimeout: resourceInsightUpdate,
		DeleteWithoutTimeout: resourceInsightDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				names.AttrARN: {
					Type:     schema.TypeString,
					Computed: true,
				},
				"filters": {
					Type:     schema.TypeList,
					Required: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							names.AttrAWSAccountID:                        stringFilterSchema(),
							"company_name":                                stringFilterSchema(),
							"compliance_status":                           stringFilterSchema(),
							"confidence":                                  numberFilterSchema(),
							names.AttrCreatedAt:                           dateFilterSchema(),
							"criticality":                                 numberFilterSchema(),
							names.AttrDescription:                         stringFilterSchema(),
							"finding_provider_fields_confidence":          numberFilterSchema(),
							"finding_provider_fields_criticality":         numberFilterSchema(),
							"finding_provider_fields_related_findings_id": stringFilterSchema(),
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
				},
				"group_by_attribute": {
					Type:     schema.TypeString,
					Required: true,
				},
				names.AttrName: {
					Type:     schema.TypeString,
					Required: true,
				},
			}
		},
	}
}

func resourceInsightCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SecurityHubClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &securityhub.CreateInsightInput{
		GroupByAttribute: aws.String(d.Get("group_by_attribute").(string)),
		Name:             aws.String(name),
	}

	if v, ok := d.GetOk("filters"); ok {
		input.Filters = expandSecurityFindingFilters(v.([]interface{}))
	}

	output, err := conn.CreateInsight(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Security Hub Insight (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.InsightArn))

	return append(diags, resourceInsightRead(ctx, d, meta)...)
}

func resourceInsightRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SecurityHubClient(ctx)

	insight, err := findInsightByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Security Hub Insight (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Security Hub Insight (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, insight.InsightArn)
	if err := d.Set("filters", flattenSecurityFindingFilters(insight.Filters)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting filters: %s", err)
	}
	d.Set("group_by_attribute", insight.GroupByAttribute)
	d.Set(names.AttrName, insight.Name)

	return diags
}

func resourceInsightUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SecurityHubClient(ctx)

	input := &securityhub.UpdateInsightInput{
		InsightArn: aws.String(d.Id()),
	}

	if d.HasChange("filters") {
		input.Filters = expandSecurityFindingFilters(d.Get("filters").([]interface{}))
	}

	if d.HasChange("group_by_attribute") {
		input.GroupByAttribute = aws.String(d.Get("group_by_attribute").(string))
	}

	if v, ok := d.GetOk(names.AttrName); ok {
		input.Name = aws.String(v.(string))
	}

	_, err := conn.UpdateInsight(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Security Hub Insight (%s): %s", d.Id(), err)
	}

	return append(diags, resourceInsightRead(ctx, d, meta)...)
}

func resourceInsightDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SecurityHubClient(ctx)

	log.Printf("[DEBUG] Deleting Security Hub Insight: %s", d.Id())
	_, err := conn.DeleteInsight(ctx, &securityhub.DeleteInsightInput{
		InsightArn: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, errCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Security Hub Insight (%s): %s", d.Id(), err)
	}

	return diags
}

func findInsightByARN(ctx context.Context, conn *securityhub.Client, arn string) (*types.Insight, error) {
	input := &securityhub.GetInsightsInput{
		InsightArns: []string{arn},
	}

	return findInsight(ctx, conn, input)
}

func findInsight(ctx context.Context, conn *securityhub.Client, input *securityhub.GetInsightsInput) (*types.Insight, error) {
	output, err := findInsights(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findInsights(ctx context.Context, conn *securityhub.Client, input *securityhub.GetInsightsInput) ([]types.Insight, error) {
	var output []types.Insight

	pages := securityhub.NewGetInsightsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if tfawserr.ErrCodeEquals(err, errCodeResourceNotFoundException) || tfawserr.ErrMessageContains(err, errCodeInvalidAccessException, "not subscribed to AWS Security Hub") {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.Insights...)
	}

	return output, nil
}

func dateFilterSchema() *schema.Schema {
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
							names.AttrUnit: {
								Type:             schema.TypeString,
								Required:         true,
								ValidateDiagFunc: enum.Validate[types.DateRangeUnit](),
							},
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
}

func ipFilterSchema() *schema.Schema {
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
}

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

func mapFilterSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		MaxItems: 20,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"comparison": {
					Type:             schema.TypeString,
					Required:         true,
					ValidateDiagFunc: enum.Validate[types.MapFilterComparison](),
				},
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
}

func numberFilterSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		MaxItems: 20,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"eq": {
					Type:         schema.TypeString,
					Optional:     true,
					ValidateFunc: verify.ValidTypeStringNullableFloat,
				},
				"gte": {
					Type:         schema.TypeString,
					Optional:     true,
					ValidateFunc: verify.ValidTypeStringNullableFloat,
				},
				"lte": {
					Type:         schema.TypeString,
					Optional:     true,
					ValidateFunc: verify.ValidTypeStringNullableFloat,
				},
			},
		},
	}
}

func stringFilterSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		MaxItems: 20,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"comparison": {
					Type:             schema.TypeString,
					Required:         true,
					ValidateDiagFunc: enum.Validate[types.StringFilterComparison](),
				},
				names.AttrValue: {
					Type:     schema.TypeString,
					Required: true,
				},
			},
		},
	}
}

func workflowStatusSchema() *schema.Schema {
	s := stringFilterSchema()

	s.Elem.(*schema.Resource).Schema[names.AttrValue].ValidateDiagFunc = enum.Validate[types.WorkflowStatus]()

	return s
}

func expandDateFilterDateRange(l []interface{}) *types.DateRange {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})
	if !ok {
		return nil
	}

	dr := &types.DateRange{}

	if v, ok := tfMap[names.AttrUnit].(string); ok && v != "" {
		dr.Unit = types.DateRangeUnit(v)
	}

	if v, ok := tfMap[names.AttrValue].(int); ok {
		dr.Value = aws.Int32(int32(v))
	}

	return dr
}

func expandDateFilters(l []interface{}) []types.DateFilter {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	var dateFilters []types.DateFilter

	for _, item := range l {
		tfMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		df := types.DateFilter{}

		if v, ok := tfMap["date_range"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
			df.DateRange = expandDateFilterDateRange(v)
		}

		if v, ok := tfMap["end"].(string); ok && v != "" {
			df.End = aws.String(v)
		}

		if v, ok := tfMap["start"].(string); ok && v != "" {
			df.Start = aws.String(v)
		}

		dateFilters = append(dateFilters, df)
	}

	return dateFilters
}

func expandSecurityFindingFilters(l []interface{}) *types.AwsSecurityFindingFilters {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})
	if !ok {
		return nil
	}

	filters := &types.AwsSecurityFindingFilters{}

	if v, ok := tfMap[names.AttrAWSAccountID].(*schema.Set); ok && v.Len() > 0 {
		filters.AwsAccountId = expandStringFilters(v.List())
	}

	if v, ok := tfMap["company_name"].(*schema.Set); ok && v.Len() > 0 {
		filters.CompanyName = expandStringFilters(v.List())
	}

	if v, ok := tfMap["compliance_status"].(*schema.Set); ok && v.Len() > 0 {
		filters.ComplianceStatus = expandStringFilters(v.List())
	}

	if v, ok := tfMap["confidence"].(*schema.Set); ok && v.Len() > 0 {
		filters.Confidence = expandNumberFilters(v.List())
	}

	if v, ok := tfMap[names.AttrCreatedAt].(*schema.Set); ok && v.Len() > 0 {
		filters.CreatedAt = expandDateFilters(v.List())
	}

	if v, ok := tfMap["criticality"].(*schema.Set); ok && v.Len() > 0 {
		filters.Criticality = expandNumberFilters(v.List())
	}

	if v, ok := tfMap[names.AttrDescription].(*schema.Set); ok && v.Len() > 0 {
		filters.Description = expandStringFilters(v.List())
	}

	if v, ok := tfMap["finding_provider_fields_confidence"].(*schema.Set); ok && v.Len() > 0 {
		filters.FindingProviderFieldsConfidence = expandNumberFilters(v.List())
	}

	if v, ok := tfMap["finding_provider_fields_criticality"].(*schema.Set); ok && v.Len() > 0 {
		filters.FindingProviderFieldsCriticality = expandNumberFilters(v.List())
	}

	if v, ok := tfMap["finding_provider_fields_related_findings_id"].(*schema.Set); ok && v.Len() > 0 {
		filters.FindingProviderFieldsRelatedFindingsId = expandStringFilters(v.List())
	}

	if v, ok := tfMap["finding_provider_fields_related_findings_product_arn"].(*schema.Set); ok && v.Len() > 0 {
		filters.FindingProviderFieldsRelatedFindingsProductArn = expandStringFilters(v.List())
	}

	if v, ok := tfMap["finding_provider_fields_severity_label"].(*schema.Set); ok && v.Len() > 0 {
		filters.FindingProviderFieldsSeverityLabel = expandStringFilters(v.List())
	}

	if v, ok := tfMap["finding_provider_fields_severity_original"].(*schema.Set); ok && v.Len() > 0 {
		filters.FindingProviderFieldsSeverityOriginal = expandStringFilters(v.List())
	}

	if v, ok := tfMap["finding_provider_fields_types"].(*schema.Set); ok && v.Len() > 0 {
		filters.FindingProviderFieldsTypes = expandStringFilters(v.List())
	}

	if v, ok := tfMap["first_observed_at"].(*schema.Set); ok && v.Len() > 0 {
		filters.FirstObservedAt = expandDateFilters(v.List())
	}

	if v, ok := tfMap["generator_id"].(*schema.Set); ok && v.Len() > 0 {
		filters.GeneratorId = expandStringFilters(v.List())
	}

	if v, ok := tfMap[names.AttrID].(*schema.Set); ok && v.Len() > 0 {
		filters.Id = expandStringFilters(v.List())
	}

	if v, ok := tfMap["keyword"].(*schema.Set); ok && v.Len() > 0 {
		filters.Keyword = expandKeywordFilters(v.List())
	}

	if v, ok := tfMap["last_observed_at"].(*schema.Set); ok && v.Len() > 0 {
		filters.LastObservedAt = expandDateFilters(v.List())
	}

	if v, ok := tfMap["malware_name"].(*schema.Set); ok && v.Len() > 0 {
		filters.MalwareName = expandStringFilters(v.List())
	}

	if v, ok := tfMap["malware_path"].(*schema.Set); ok && v.Len() > 0 {
		filters.MalwarePath = expandStringFilters(v.List())
	}

	if v, ok := tfMap["malware_state"].(*schema.Set); ok && v.Len() > 0 {
		filters.MalwareState = expandStringFilters(v.List())
	}

	if v, ok := tfMap["malware_type"].(*schema.Set); ok && v.Len() > 0 {
		filters.MalwareType = expandStringFilters(v.List())
	}

	if v, ok := tfMap["network_destination_domain"].(*schema.Set); ok && v.Len() > 0 {
		filters.NetworkDestinationDomain = expandStringFilters(v.List())
	}

	if v, ok := tfMap["network_destination_ipv4"].(*schema.Set); ok && v.Len() > 0 {
		filters.NetworkDestinationIpV4 = expandIPFilters(v.List())
	}

	if v, ok := tfMap["network_destination_ipv6"].(*schema.Set); ok && v.Len() > 0 {
		filters.NetworkDestinationIpV6 = expandIPFilters(v.List())
	}

	if v, ok := tfMap["network_destination_port"].(*schema.Set); ok && v.Len() > 0 {
		filters.NetworkDestinationPort = expandNumberFilters(v.List())
	}

	if v, ok := tfMap["network_direction"].(*schema.Set); ok && v.Len() > 0 {
		filters.NetworkDirection = expandStringFilters(v.List())
	}

	if v, ok := tfMap["network_protocol"].(*schema.Set); ok && v.Len() > 0 {
		filters.NetworkProtocol = expandStringFilters(v.List())
	}

	if v, ok := tfMap["network_source_domain"].(*schema.Set); ok && v.Len() > 0 {
		filters.NetworkSourceDomain = expandStringFilters(v.List())
	}

	if v, ok := tfMap["network_source_ipv4"].(*schema.Set); ok && v.Len() > 0 {
		filters.NetworkSourceIpV4 = expandIPFilters(v.List())
	}

	if v, ok := tfMap["network_source_ipv6"].(*schema.Set); ok && v.Len() > 0 {
		filters.NetworkSourceIpV6 = expandIPFilters(v.List())
	}

	if v, ok := tfMap["network_source_mac"].(*schema.Set); ok && v.Len() > 0 {
		filters.NetworkSourceMac = expandStringFilters(v.List())
	}

	if v, ok := tfMap["network_source_port"].(*schema.Set); ok && v.Len() > 0 {
		filters.NetworkSourcePort = expandNumberFilters(v.List())
	}

	if v, ok := tfMap["note_text"].(*schema.Set); ok && v.Len() > 0 {
		filters.NoteText = expandStringFilters(v.List())
	}

	if v, ok := tfMap["note_updated_at"].(*schema.Set); ok && v.Len() > 0 {
		filters.NoteUpdatedAt = expandDateFilters(v.List())
	}

	if v, ok := tfMap["note_updated_by"].(*schema.Set); ok && v.Len() > 0 {
		filters.NoteUpdatedBy = expandStringFilters(v.List())
	}

	if v, ok := tfMap["process_launched_at"].(*schema.Set); ok && v.Len() > 0 {
		filters.ProcessLaunchedAt = expandDateFilters(v.List())
	}

	if v, ok := tfMap["process_name"].(*schema.Set); ok && v.Len() > 0 {
		filters.ProcessName = expandStringFilters(v.List())
	}

	if v, ok := tfMap["process_parent_pid"].(*schema.Set); ok && v.Len() > 0 {
		filters.ProcessParentPid = expandNumberFilters(v.List())
	}

	if v, ok := tfMap["process_path"].(*schema.Set); ok && v.Len() > 0 {
		filters.ProcessPath = expandStringFilters(v.List())
	}

	if v, ok := tfMap["process_pid"].(*schema.Set); ok && v.Len() > 0 {
		filters.ProcessPid = expandNumberFilters(v.List())
	}

	if v, ok := tfMap["process_terminated_at"].(*schema.Set); ok && v.Len() > 0 {
		filters.ProcessTerminatedAt = expandDateFilters(v.List())
	}

	if v, ok := tfMap["product_arn"].(*schema.Set); ok && v.Len() > 0 {
		filters.ProductArn = expandStringFilters(v.List())
	}

	if v, ok := tfMap["product_fields"].(*schema.Set); ok && v.Len() > 0 {
		filters.ProductFields = expandMapFilters(v.List())
	}

	if v, ok := tfMap["product_name"].(*schema.Set); ok && v.Len() > 0 {
		filters.ProductName = expandStringFilters(v.List())
	}

	if v, ok := tfMap["recommendation_text"].(*schema.Set); ok && v.Len() > 0 {
		filters.RecommendationText = expandStringFilters(v.List())
	}

	if v, ok := tfMap["record_state"].(*schema.Set); ok && v.Len() > 0 {
		filters.RecordState = expandStringFilters(v.List())
	}

	if v, ok := tfMap["related_findings_id"].(*schema.Set); ok && v.Len() > 0 {
		filters.RelatedFindingsId = expandStringFilters(v.List())
	}

	if v, ok := tfMap["related_findings_product_arn"].(*schema.Set); ok && v.Len() > 0 {
		filters.RelatedFindingsProductArn = expandStringFilters(v.List())
	}

	if v, ok := tfMap["resource_aws_ec2_instance_iam_instance_profile_arn"].(*schema.Set); ok && v.Len() > 0 {
		filters.ResourceAwsEc2InstanceIamInstanceProfileArn = expandStringFilters(v.List())
	}

	if v, ok := tfMap["resource_aws_ec2_instance_image_id"].(*schema.Set); ok && v.Len() > 0 {
		filters.ResourceAwsEc2InstanceImageId = expandStringFilters(v.List())
	}

	if v, ok := tfMap["resource_aws_ec2_instance_ipv4_addresses"].(*schema.Set); ok && v.Len() > 0 {
		filters.ResourceAwsEc2InstanceIpV4Addresses = expandIPFilters(v.List())
	}

	if v, ok := tfMap["resource_aws_ec2_instance_ipv6_addresses"].(*schema.Set); ok && v.Len() > 0 {
		filters.ResourceAwsEc2InstanceIpV6Addresses = expandIPFilters(v.List())
	}

	if v, ok := tfMap["resource_aws_ec2_instance_key_name"].(*schema.Set); ok && v.Len() > 0 {
		filters.ResourceAwsEc2InstanceKeyName = expandStringFilters(v.List())
	}

	if v, ok := tfMap["resource_aws_ec2_instance_launched_at"].(*schema.Set); ok && v.Len() > 0 {
		filters.ResourceAwsEc2InstanceLaunchedAt = expandDateFilters(v.List())
	}

	if v, ok := tfMap["resource_aws_ec2_instance_subnet_id"].(*schema.Set); ok && v.Len() > 0 {
		filters.ResourceAwsEc2InstanceSubnetId = expandStringFilters(v.List())
	}

	if v, ok := tfMap["resource_aws_ec2_instance_type"].(*schema.Set); ok && v.Len() > 0 {
		filters.ResourceAwsEc2InstanceType = expandStringFilters(v.List())
	}

	if v, ok := tfMap["resource_aws_ec2_instance_vpc_id"].(*schema.Set); ok && v.Len() > 0 {
		filters.ResourceAwsEc2InstanceVpcId = expandStringFilters(v.List())
	}

	if v, ok := tfMap["resource_aws_iam_access_key_created_at"].(*schema.Set); ok && v.Len() > 0 {
		filters.ResourceAwsIamAccessKeyCreatedAt = expandDateFilters(v.List())
	}

	if v, ok := tfMap["resource_aws_iam_access_key_status"].(*schema.Set); ok && v.Len() > 0 {
		filters.ResourceAwsIamAccessKeyStatus = expandStringFilters(v.List())
	}

	if v, ok := tfMap["resource_aws_iam_access_key_user_name"].(*schema.Set); ok && v.Len() > 0 {
		filters.ResourceAwsIamAccessKeyUserName = expandStringFilters(v.List())
	}

	if v, ok := tfMap["resource_aws_s3_bucket_owner_id"].(*schema.Set); ok && v.Len() > 0 {
		filters.ResourceAwsS3BucketOwnerId = expandStringFilters(v.List())
	}

	if v, ok := tfMap["resource_aws_s3_bucket_owner_name"].(*schema.Set); ok && v.Len() > 0 {
		filters.ResourceAwsS3BucketOwnerName = expandStringFilters(v.List())
	}

	if v, ok := tfMap["resource_container_image_id"].(*schema.Set); ok && v.Len() > 0 {
		filters.ResourceContainerImageId = expandStringFilters(v.List())
	}

	if v, ok := tfMap["resource_container_image_name"].(*schema.Set); ok && v.Len() > 0 {
		filters.ResourceContainerImageName = expandStringFilters(v.List())
	}

	if v, ok := tfMap["resource_container_launched_at"].(*schema.Set); ok && v.Len() > 0 {
		filters.ResourceContainerLaunchedAt = expandDateFilters(v.List())
	}

	if v, ok := tfMap["resource_container_name"].(*schema.Set); ok && v.Len() > 0 {
		filters.ResourceContainerName = expandStringFilters(v.List())
	}

	if v, ok := tfMap["resource_details_other"].(*schema.Set); ok && v.Len() > 0 {
		filters.ResourceDetailsOther = expandMapFilters(v.List())
	}

	if v, ok := tfMap[names.AttrResourceID].(*schema.Set); ok && v.Len() > 0 {
		filters.ResourceId = expandStringFilters(v.List())
	}

	if v, ok := tfMap["resource_partition"].(*schema.Set); ok && v.Len() > 0 {
		filters.ResourcePartition = expandStringFilters(v.List())
	}

	if v, ok := tfMap["resource_region"].(*schema.Set); ok && v.Len() > 0 {
		filters.ResourceRegion = expandStringFilters(v.List())
	}

	if v, ok := tfMap[names.AttrResourceTags].(*schema.Set); ok && v.Len() > 0 {
		filters.ResourceTags = expandMapFilters(v.List())
	}

	if v, ok := tfMap[names.AttrResourceType].(*schema.Set); ok && v.Len() > 0 {
		filters.ResourceType = expandStringFilters(v.List())
	}

	if v, ok := tfMap["severity_label"].(*schema.Set); ok && v.Len() > 0 {
		filters.SeverityLabel = expandStringFilters(v.List())
	}

	if v, ok := tfMap["source_url"].(*schema.Set); ok && v.Len() > 0 {
		filters.SourceUrl = expandStringFilters(v.List())
	}

	if v, ok := tfMap["threat_intel_indicator_category"].(*schema.Set); ok && v.Len() > 0 {
		filters.ThreatIntelIndicatorCategory = expandStringFilters(v.List())
	}

	if v, ok := tfMap["threat_intel_indicator_last_observed_at"].(*schema.Set); ok && v.Len() > 0 {
		filters.ThreatIntelIndicatorLastObservedAt = expandDateFilters(v.List())
	}

	if v, ok := tfMap["threat_intel_indicator_source"].(*schema.Set); ok && v.Len() > 0 {
		filters.ThreatIntelIndicatorSource = expandStringFilters(v.List())
	}

	if v, ok := tfMap["threat_intel_indicator_source_url"].(*schema.Set); ok && v.Len() > 0 {
		filters.ThreatIntelIndicatorSourceUrl = expandStringFilters(v.List())
	}

	if v, ok := tfMap["threat_intel_indicator_type"].(*schema.Set); ok && v.Len() > 0 {
		filters.ThreatIntelIndicatorType = expandStringFilters(v.List())
	}

	if v, ok := tfMap["threat_intel_indicator_value"].(*schema.Set); ok && v.Len() > 0 {
		filters.ThreatIntelIndicatorValue = expandStringFilters(v.List())
	}

	if v, ok := tfMap["title"].(*schema.Set); ok && v.Len() > 0 {
		filters.Title = expandStringFilters(v.List())
	}

	if v, ok := tfMap[names.AttrType].(*schema.Set); ok && v.Len() > 0 {
		filters.Type = expandStringFilters(v.List())
	}

	if v, ok := tfMap["updated_at"].(*schema.Set); ok && v.Len() > 0 {
		filters.UpdatedAt = expandDateFilters(v.List())
	}

	if v, ok := tfMap["user_defined_values"].(*schema.Set); ok && v.Len() > 0 {
		filters.UserDefinedFields = expandMapFilters(v.List())
	}

	if v, ok := tfMap["verification_state"].(*schema.Set); ok && v.Len() > 0 {
		filters.VerificationState = expandStringFilters(v.List())
	}

	if v, ok := tfMap["workflow_status"].(*schema.Set); ok && v.Len() > 0 {
		filters.WorkflowStatus = expandStringFilters(v.List())
	}

	return filters
}

func expandIPFilters(l []interface{}) []types.IpFilter {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	var ipFilters []types.IpFilter

	for _, item := range l {
		tfMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		ipFilter := types.IpFilter{}

		if v, ok := tfMap["cidr"].(string); ok && v != "" {
			ipFilter.Cidr = aws.String(v)
		}

		ipFilters = append(ipFilters, ipFilter)
	}

	return ipFilters
}

func expandKeywordFilters(l []interface{}) []types.KeywordFilter {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	var keywordFilters []types.KeywordFilter

	for _, item := range l {
		tfMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		kf := types.KeywordFilter{}

		if v, ok := tfMap[names.AttrValue].(string); ok && v != "" {
			kf.Value = aws.String(v)
		}

		keywordFilters = append(keywordFilters, kf)
	}

	return keywordFilters
}

func expandMapFilters(l []interface{}) []types.MapFilter {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	var mapFilters []types.MapFilter

	for _, item := range l {
		tfMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		mf := types.MapFilter{}

		if v, ok := tfMap["comparison"].(string); ok && v != "" {
			mf.Comparison = types.MapFilterComparison(v)
		}

		if v, ok := tfMap[names.AttrKey].(string); ok && v != "" {
			mf.Key = aws.String(v)
		}

		if v, ok := tfMap[names.AttrValue].(string); ok && v != "" {
			mf.Value = aws.String(v)
		}

		mapFilters = append(mapFilters, mf)
	}

	return mapFilters
}

func expandNumberFilters(l []interface{}) []types.NumberFilter {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	var numFilters []types.NumberFilter

	for _, item := range l {
		tfMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		nf := types.NumberFilter{}

		if v, ok := tfMap["eq"].(string); ok && v != "" {
			val, err := strconv.ParseFloat(v, 64)
			if err == nil {
				nf.Eq = aws.Float64(val)
			}
		}

		if v, ok := tfMap["gte"].(string); ok && v != "" {
			val, err := strconv.ParseFloat(v, 64)
			if err == nil {
				nf.Gte = aws.Float64(val)
			}
		}

		if v, ok := tfMap["lte"].(string); ok && v != "" {
			val, err := strconv.ParseFloat(v, 64)
			if err == nil {
				nf.Lte = aws.Float64(val)
			}
		}

		numFilters = append(numFilters, nf)
	}

	return numFilters
}

func expandStringFilters(l []interface{}) []types.StringFilter {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	var stringFilters []types.StringFilter

	for _, item := range l {
		tfMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		sf := types.StringFilter{}

		if v, ok := tfMap["comparison"].(string); ok && v != "" {
			sf.Comparison = types.StringFilterComparison(v)
		}

		if v, ok := tfMap[names.AttrValue].(string); ok && v != "" {
			sf.Value = aws.String(v)
		}

		stringFilters = append(stringFilters, sf)
	}

	return stringFilters
}

func flattenDateFilterDateRange(dateRange *types.DateRange) []interface{} {
	if dateRange == nil {
		return nil
	}

	m := map[string]interface{}{
		names.AttrUnit:  string(dateRange.Unit),
		names.AttrValue: aws.ToInt32((dateRange.Value)),
	}

	return []interface{}{m}
}

func flattenDateFilters(filters []types.DateFilter) []interface{} {
	if len(filters) == 0 {
		return nil
	}

	var dateFilters []interface{}

	for _, filter := range filters {
		m := map[string]interface{}{
			"date_range": flattenDateFilterDateRange(filter.DateRange),
			"end":        aws.ToString(filter.End),
			"start":      aws.ToString(filter.Start),
		}

		dateFilters = append(dateFilters, m)
	}

	return dateFilters
}

func flattenIPFilters(filters []types.IpFilter) []interface{} {
	if len(filters) == 0 {
		return nil
	}

	var ipFilters []interface{}

	for _, filter := range filters {
		m := map[string]interface{}{
			"cidr": aws.ToString(filter.Cidr),
		}

		ipFilters = append(ipFilters, m)
	}

	return ipFilters
}

func flattenKeywordFilters(filters []types.KeywordFilter) []interface{} {
	if len(filters) == 0 {
		return nil
	}

	var keywordFilters []interface{}

	for _, filter := range filters {
		m := map[string]interface{}{
			names.AttrValue: aws.ToString(filter.Value),
		}

		keywordFilters = append(keywordFilters, m)
	}

	return keywordFilters
}

func flattenMapFilters(filters []types.MapFilter) []interface{} {
	if len(filters) == 0 {
		return nil
	}

	var mapFilters []interface{}

	for _, filter := range filters {
		m := map[string]interface{}{
			"comparison":    string(filter.Comparison),
			names.AttrKey:   aws.ToString(filter.Key),
			names.AttrValue: aws.ToString(filter.Value),
		}

		mapFilters = append(mapFilters, m)
	}

	return mapFilters
}

func flattenNumberFilters(filters []types.NumberFilter) []interface{} {
	if len(filters) == 0 {
		return nil
	}

	var numFilters []interface{}

	for _, filter := range filters {
		m := map[string]interface{}{}

		if filter.Eq != nil {
			m["eq"] = strconv.FormatFloat(aws.ToFloat64(filter.Eq), 'f', -1, 64)
		}

		if filter.Gte != nil {
			m["gte"] = strconv.FormatFloat(aws.ToFloat64(filter.Gte), 'f', -1, 64)
		}

		if filter.Lte != nil {
			m["lte"] = strconv.FormatFloat(aws.ToFloat64(filter.Lte), 'f', -1, 64)
		}

		numFilters = append(numFilters, m)
	}

	return numFilters
}

func flattenSecurityFindingFilters(filters *types.AwsSecurityFindingFilters) []interface{} {
	if filters == nil {
		return nil
	}

	m := map[string]interface{}{
		names.AttrAWSAccountID:                        flattenStringFilters(filters.AwsAccountId),
		"company_name":                                flattenStringFilters(filters.CompanyName),
		"compliance_status":                           flattenStringFilters(filters.ComplianceStatus),
		"confidence":                                  flattenNumberFilters(filters.Confidence),
		names.AttrCreatedAt:                           flattenDateFilters(filters.CreatedAt),
		"criticality":                                 flattenNumberFilters(filters.Criticality),
		names.AttrDescription:                         flattenStringFilters(filters.Description),
		"finding_provider_fields_confidence":          flattenNumberFilters(filters.FindingProviderFieldsConfidence),
		"finding_provider_fields_criticality":         flattenNumberFilters(filters.FindingProviderFieldsCriticality),
		"finding_provider_fields_related_findings_id": flattenStringFilters(filters.FindingProviderFieldsRelatedFindingsId),
		"finding_provider_fields_related_findings_product_arn": flattenStringFilters(filters.FindingProviderFieldsRelatedFindingsProductArn),
		"finding_provider_fields_severity_label":               flattenStringFilters(filters.FindingProviderFieldsSeverityLabel),
		"finding_provider_fields_severity_original":            flattenStringFilters(filters.FindingProviderFieldsSeverityOriginal),
		"finding_provider_fields_types":                        flattenStringFilters(filters.FindingProviderFieldsTypes),
		"first_observed_at":                                    flattenDateFilters(filters.FirstObservedAt),
		"generator_id":                                         flattenStringFilters(filters.GeneratorId),
		names.AttrID:                                           flattenStringFilters(filters.Id),
		"keyword":                                              flattenKeywordFilters(filters.Keyword),
		"last_observed_at":                                     flattenDateFilters(filters.LastObservedAt),
		"malware_name":                                         flattenStringFilters(filters.MalwareName),
		"malware_path":                                         flattenStringFilters(filters.MalwarePath),
		"malware_state":                                        flattenStringFilters(filters.MalwareState),
		"malware_type":                                         flattenStringFilters(filters.MalwareType),
		"network_destination_domain":                           flattenStringFilters(filters.NetworkDestinationDomain),
		"network_destination_ipv4":                             flattenIPFilters(filters.NetworkDestinationIpV4),
		"network_destination_ipv6":                             flattenIPFilters(filters.NetworkDestinationIpV6),
		"network_destination_port":                             flattenNumberFilters(filters.NetworkDestinationPort),
		"network_direction":                                    flattenStringFilters(filters.NetworkDirection),
		"network_protocol":                                     flattenStringFilters(filters.NetworkProtocol),
		"network_source_domain":                                flattenStringFilters(filters.NetworkSourceDomain),
		"network_source_ipv4":                                  flattenIPFilters(filters.NetworkSourceIpV4),
		"network_source_ipv6":                                  flattenIPFilters(filters.NetworkSourceIpV6),
		"network_source_mac":                                   flattenStringFilters(filters.NetworkSourceMac),
		"network_source_port":                                  flattenNumberFilters(filters.NetworkSourcePort),
		"note_text":                                            flattenStringFilters(filters.NoteText),
		"note_updated_at":                                      flattenDateFilters(filters.NoteUpdatedAt),
		"note_updated_by":                                      flattenStringFilters(filters.NoteUpdatedBy),
		"process_launched_at":                                  flattenDateFilters(filters.ProcessLaunchedAt),
		"process_name":                                         flattenStringFilters(filters.ProcessName),
		"process_parent_pid":                                   flattenNumberFilters(filters.ProcessParentPid),
		"process_path":                                         flattenStringFilters(filters.ProcessPath),
		"process_pid":                                          flattenNumberFilters(filters.ProcessPid),
		"process_terminated_at":                                flattenDateFilters(filters.ProcessTerminatedAt),
		"product_arn":                                          flattenStringFilters(filters.ProductArn),
		"product_fields":                                       flattenMapFilters(filters.ProductFields),
		"product_name":                                         flattenStringFilters(filters.ProductName),
		"recommendation_text":                                  flattenStringFilters(filters.RecommendationText),
		"record_state":                                         flattenStringFilters(filters.RecordState),
		"related_findings_id":                                  flattenStringFilters(filters.RelatedFindingsId),
		"related_findings_product_arn":                         flattenStringFilters(filters.RelatedFindingsProductArn),
		"resource_aws_ec2_instance_iam_instance_profile_arn": flattenStringFilters(filters.ResourceAwsEc2InstanceIamInstanceProfileArn),
		"resource_aws_ec2_instance_image_id":                 flattenStringFilters(filters.ResourceAwsEc2InstanceImageId),
		"resource_aws_ec2_instance_ipv4_addresses":           flattenIPFilters(filters.ResourceAwsEc2InstanceIpV4Addresses),
		"resource_aws_ec2_instance_ipv6_addresses":           flattenIPFilters(filters.ResourceAwsEc2InstanceIpV6Addresses),
		"resource_aws_ec2_instance_key_name":                 flattenStringFilters(filters.ResourceAwsEc2InstanceKeyName),
		"resource_aws_ec2_instance_launched_at":              flattenDateFilters(filters.ResourceAwsEc2InstanceLaunchedAt),
		"resource_aws_ec2_instance_subnet_id":                flattenStringFilters(filters.ResourceAwsEc2InstanceSubnetId),
		"resource_aws_ec2_instance_type":                     flattenStringFilters(filters.ResourceAwsEc2InstanceType),
		"resource_aws_ec2_instance_vpc_id":                   flattenStringFilters(filters.ResourceAwsEc2InstanceVpcId),
		"resource_aws_iam_access_key_created_at":             flattenDateFilters(filters.ResourceAwsIamAccessKeyCreatedAt),
		"resource_aws_iam_access_key_status":                 flattenStringFilters(filters.ResourceAwsIamAccessKeyStatus),
		"resource_aws_iam_access_key_user_name":              flattenStringFilters(filters.ResourceAwsIamAccessKeyUserName),
		"resource_aws_s3_bucket_owner_id":                    flattenStringFilters(filters.ResourceAwsS3BucketOwnerId),
		"resource_aws_s3_bucket_owner_name":                  flattenStringFilters(filters.ResourceAwsS3BucketOwnerName),
		"resource_container_image_id":                        flattenStringFilters(filters.ResourceContainerImageId),
		"resource_container_image_name":                      flattenStringFilters(filters.ResourceContainerImageName),
		"resource_container_launched_at":                     flattenDateFilters(filters.ResourceContainerLaunchedAt),
		"resource_container_name":                            flattenStringFilters(filters.ResourceContainerName),
		"resource_details_other":                             flattenMapFilters(filters.ResourceDetailsOther),
		names.AttrResourceID:                                 flattenStringFilters(filters.ResourceId),
		"resource_partition":                                 flattenStringFilters(filters.ResourcePartition),
		"resource_region":                                    flattenStringFilters(filters.ResourceRegion),
		names.AttrResourceTags:                               flattenMapFilters(filters.ResourceTags),
		names.AttrResourceType:                               flattenStringFilters(filters.ResourceType),
		"severity_label":                                     flattenStringFilters(filters.SeverityLabel),
		"source_url":                                         flattenStringFilters(filters.ThreatIntelIndicatorSourceUrl),
		"threat_intel_indicator_category":                    flattenStringFilters(filters.ThreatIntelIndicatorCategory),
		"threat_intel_indicator_last_observed_at":            flattenDateFilters(filters.ThreatIntelIndicatorLastObservedAt),
		"threat_intel_indicator_source":                      flattenStringFilters(filters.ThreatIntelIndicatorSource),
		"threat_intel_indicator_source_url":                  flattenStringFilters(filters.ThreatIntelIndicatorSourceUrl),
		"threat_intel_indicator_type":                        flattenStringFilters(filters.ThreatIntelIndicatorType),
		"threat_intel_indicator_value":                       flattenStringFilters(filters.ThreatIntelIndicatorValue),
		"title":                                              flattenStringFilters(filters.Title),
		names.AttrType:                                       flattenStringFilters(filters.Type),
		"updated_at":                                         flattenDateFilters(filters.UpdatedAt),
		"user_defined_values":                                flattenMapFilters(filters.UserDefinedFields),
		"verification_state":                                 flattenStringFilters(filters.VerificationState),
		"workflow_status":                                    flattenStringFilters(filters.WorkflowStatus),
	}

	return []interface{}{m}
}

func flattenStringFilters(filters []types.StringFilter) []interface{} {
	if len(filters) == 0 {
		return nil
	}

	var stringFilters []interface{}

	for _, filter := range filters {
		m := map[string]interface{}{
			"comparison":    string(filter.Comparison),
			names.AttrValue: aws.ToString(filter.Value),
		}

		stringFilters = append(stringFilters, m)
	}

	return stringFilters
}
