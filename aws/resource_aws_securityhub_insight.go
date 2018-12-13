package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/securityhub"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
)

const (
	SecurityHubFilterComparisonEquals   = "EQUALS"
	SecurityHubFilterComparisonContains = "CONTAINS"
	SecurityHubFilterComparisonPrefix   = "PREFIX"
)

const (
	SecurityHubFilterDateRangeUnitDays = "DAYS"
)

func resourceAwsSecurityHubInsight() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsSecurityHubInsightCreate,
		Read:   resourceAwsSecurityHubInsightRead,
		Update: resourceAwsSecurityHubInsightUpdate,
		Delete: resourceAwsSecurityHubInsightDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"group_by_attribute": {
				Type:     schema.TypeString,
				Required: true,
			},
			"filter": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"product_arn":                               securityHubStringFilterSchema(),
						"aws_account_id":                            securityHubStringFilterSchema(),
						"id":                                        securityHubStringFilterSchema(),
						"generator_id":                              securityHubStringFilterSchema(),
						"type":                                      securityHubStringFilterSchema(),
						"first_observed_at":                         securityHubDateFilterSchema(),
						"last_observed_at":                          securityHubDateFilterSchema(),
						"created_at":                                securityHubDateFilterSchema(),
						"updated_at":                                securityHubDateFilterSchema(),
						"severity_product":                          securityHubNumberFilterSchema(),
						"severity_normalized":                       securityHubNumberFilterSchema(),
						"severity_label":                            securityHubStringFilterSchema(),
						"confidence":                                securityHubNumberFilterSchema(),
						"criticality":                               securityHubNumberFilterSchema(),
						"title":                                     securityHubStringFilterSchema(),
						"description":                               securityHubStringFilterSchema(),
						"recommendation_text":                       securityHubStringFilterSchema(),
						"source_url":                                securityHubStringFilterSchema(),
						"product_fields":                            securityHubMapFilterSchema(),
						"product_name":                              securityHubStringFilterSchema(),
						"company_name":                              securityHubStringFilterSchema(),
						"user_defined_fields":                       securityHubMapFilterSchema(),
						"malware_name":                              securityHubStringFilterSchema(),
						"malware_type":                              securityHubStringFilterSchema(),
						"malware_path":                              securityHubStringFilterSchema(),
						"malware_state":                             securityHubStringFilterSchema(),
						"network_direction":                         securityHubStringFilterSchema(),
						"network_protocol":                          securityHubStringFilterSchema(),
						"network_source_ip_v4":                      securityHubIpFilterSchema(),
						"network_source_ip_v6":                      securityHubIpFilterSchema(),
						"network_source_port":                       securityHubNumberFilterSchema(),
						"network_source_domain":                     securityHubStringFilterSchema(),
						"network_source_mac":                        securityHubStringFilterSchema(),
						"network_destination_ip_v4":                 securityHubIpFilterSchema(),
						"network_destination_ip_v6":                 securityHubIpFilterSchema(),
						"network_destination_port":                  securityHubNumberFilterSchema(),
						"network_destination_domain":                securityHubStringFilterSchema(),
						"process_name":                              securityHubStringFilterSchema(),
						"process_path":                              securityHubStringFilterSchema(),
						"process_pid":                               securityHubNumberFilterSchema(),
						"process_parent_pid":                        securityHubNumberFilterSchema(),
						"process_launched_at":                       securityHubDateFilterSchema(),
						"process_terminated_at":                     securityHubDateFilterSchema(),
						"threat_intel_indicator_type":               securityHubStringFilterSchema(),
						"threat_intel_indicator_value":              securityHubStringFilterSchema(),
						"threat_intel_indicator_category":           securityHubStringFilterSchema(),
						"threat_intel_indicator_last_observed_at":   securityHubDateFilterSchema(),
						"threat_intel_indicator_source":             securityHubStringFilterSchema(),
						"threat_intel_indicator_source_url":         securityHubStringFilterSchema(),
						"resource_type":                             securityHubStringFilterSchema(),
						"resource_id":                               securityHubStringFilterSchema(),
						"resource_partition":                        securityHubStringFilterSchema(),
						"resource_region":                           securityHubStringFilterSchema(),
						"resource_tags":                             securityHubMapFilterSchema(),
						"resource_aws_ec2_instance_type":            securityHubStringFilterSchema(),
						"resource_aws_ec2_instance_image_id":        securityHubStringFilterSchema(),
						"resource_aws_ec2_instance_ip_v4_addresses": securityHubIpFilterSchema(),
						"resource_aws_ec2_instance_ip_v6_addresses": securityHubIpFilterSchema(),
						"resource_aws_ec2_instance_key_name":        securityHubStringFilterSchema(),
						"resource_aws_ec2_instance_iam_instance_profile_arn": securityHubStringFilterSchema(),
						"resource_aws_ec2_instance_vpc_id":                   securityHubStringFilterSchema(),
						"resource_aws_ec2_instance_subnet_id":                securityHubStringFilterSchema(),
						"resource_aws_ec2_instance_launched_at":              securityHubDateFilterSchema(),
						"resource_aws_s3_bucket_owner_id":                    securityHubStringFilterSchema(),
						"resource_aws_s3_bucket_owner_name":                  securityHubStringFilterSchema(),
						"resource_aws_iam_access_key_user_name":              securityHubStringFilterSchema(),
						"resource_aws_iam_access_key_status":                 securityHubStringFilterSchema(),
						"resource_aws_iam_access_key_created_at":             securityHubDateFilterSchema(),
						"resource_container_name":                            securityHubStringFilterSchema(),
						"resource_container_image_id":                        securityHubStringFilterSchema(),
						"resource_container_image_name":                      securityHubStringFilterSchema(),
						"resource_container_launched_at":                     securityHubDateFilterSchema(),
						"resource_details_other":                             securityHubMapFilterSchema(),
						"compliance_status":                                  securityHubStringFilterSchema(),
						"verification_state":                                 securityHubStringFilterSchema(),
						"workflow_state":                                     securityHubStringFilterSchema(),
						"record_state":                                       securityHubStringFilterSchema(),
						"related_findings_product_arn":                       securityHubStringFilterSchema(),
						"related_findings_id":                                securityHubStringFilterSchema(),
						"note_text":                                          securityHubStringFilterSchema(),
						"note_updated_at":                                    securityHubDateFilterSchema(),
						"note_updated_by":                                    securityHubStringFilterSchema(),
						"keyword":                                            securityHubKeywordFilterSchema(),
					},
				},
			},
		},
	}
}

func expandSecurityHubAwsSecurityFindingFilters(d *schema.ResourceData) *securityhub.AwsSecurityFindingFilters {
	configs := d.Get("filter").([]interface{})
	data := configs[0].(map[string]interface{})

	return expandSecurityHubAwsSecurityFindingFiltersData(data)
}

func expandSecurityHubAwsSecurityFindingFiltersData(in map[string]interface{}) *securityhub.AwsSecurityFindingFilters {
	filters := &securityhub.AwsSecurityFindingFilters{}

	filters.ProductArn = expandSecurityHubStringFilters(in["product_arn"].(*schema.Set).List())
	filters.AwsAccountId = expandSecurityHubStringFilters(in["aws_account_id"].(*schema.Set).List())
	filters.Id = expandSecurityHubStringFilters(in["id"].(*schema.Set).List())
	filters.GeneratorId = expandSecurityHubStringFilters(in["generator_id"].(*schema.Set).List())
	filters.Type = expandSecurityHubStringFilters(in["type"].(*schema.Set).List())
	filters.FirstObservedAt = expandSecurityHubDateFilters(in["first_observed_at"].(*schema.Set).List())
	filters.LastObservedAt = expandSecurityHubDateFilters(in["last_observed_at"].(*schema.Set).List())
	filters.CreatedAt = expandSecurityHubDateFilters(in["created_at"].(*schema.Set).List())
	filters.UpdatedAt = expandSecurityHubDateFilters(in["updated_at"].(*schema.Set).List())
	filters.SeverityProduct = expandSecurityHubNumberFilters(in["severity_product"].(*schema.Set).List())
	filters.SeverityNormalized = expandSecurityHubNumberFilters(in["severity_normalized"].(*schema.Set).List())
	filters.SeverityLabel = expandSecurityHubStringFilters(in["severity_label"].(*schema.Set).List())
	filters.Confidence = expandSecurityHubNumberFilters(in["confidence"].(*schema.Set).List())
	filters.Criticality = expandSecurityHubNumberFilters(in["criticality"].(*schema.Set).List())
	filters.Title = expandSecurityHubStringFilters(in["title"].(*schema.Set).List())
	filters.Description = expandSecurityHubStringFilters(in["description"].(*schema.Set).List())
	filters.RecommendationText = expandSecurityHubStringFilters(in["recommendation_text"].(*schema.Set).List())
	filters.SourceUrl = expandSecurityHubStringFilters(in["source_url"].(*schema.Set).List())
	filters.ProductFields = expandSecurityHubMapFilters(in["product_fields"].(*schema.Set).List())
	filters.ProductName = expandSecurityHubStringFilters(in["product_name"].(*schema.Set).List())
	filters.CompanyName = expandSecurityHubStringFilters(in["company_name"].(*schema.Set).List())
	filters.UserDefinedFields = expandSecurityHubMapFilters(in["user_defined_fields"].(*schema.Set).List())
	filters.MalwareName = expandSecurityHubStringFilters(in["malware_name"].(*schema.Set).List())
	filters.MalwareType = expandSecurityHubStringFilters(in["malware_type"].(*schema.Set).List())
	filters.MalwarePath = expandSecurityHubStringFilters(in["malware_path"].(*schema.Set).List())
	filters.MalwareState = expandSecurityHubStringFilters(in["malware_state"].(*schema.Set).List())
	filters.NetworkDirection = expandSecurityHubStringFilters(in["network_direction"].(*schema.Set).List())
	filters.NetworkProtocol = expandSecurityHubStringFilters(in["network_protocol"].(*schema.Set).List())
	filters.NetworkSourceIpV4 = expandSecurityHubIpFilters(in["network_source_ip_v4"].(*schema.Set).List())
	filters.NetworkSourceIpV6 = expandSecurityHubIpFilters(in["network_source_ip_v6"].(*schema.Set).List())
	filters.NetworkSourcePort = expandSecurityHubNumberFilters(in["network_source_port"].(*schema.Set).List())
	filters.NetworkSourceDomain = expandSecurityHubStringFilters(in["network_source_domain"].(*schema.Set).List())
	filters.NetworkSourceMac = expandSecurityHubStringFilters(in["network_source_mac"].(*schema.Set).List())
	filters.NetworkDestinationIpV4 = expandSecurityHubIpFilters(in["network_destination_ip_v4"].(*schema.Set).List())
	filters.NetworkDestinationIpV6 = expandSecurityHubIpFilters(in["network_destination_ip_v6"].(*schema.Set).List())
	filters.NetworkDestinationPort = expandSecurityHubNumberFilters(in["network_destination_port"].(*schema.Set).List())
	filters.NetworkDestinationDomain = expandSecurityHubStringFilters(in["network_destination_domain"].(*schema.Set).List())
	filters.ProcessName = expandSecurityHubStringFilters(in["process_name"].(*schema.Set).List())
	filters.ProcessPath = expandSecurityHubStringFilters(in["process_path"].(*schema.Set).List())
	filters.ProcessPid = expandSecurityHubNumberFilters(in["process_pid"].(*schema.Set).List())
	filters.ProcessParentPid = expandSecurityHubNumberFilters(in["process_parent_pid"].(*schema.Set).List())
	filters.ProcessLaunchedAt = expandSecurityHubDateFilters(in["process_launched_at"].(*schema.Set).List())
	filters.ProcessTerminatedAt = expandSecurityHubDateFilters(in["process_terminated_at"].(*schema.Set).List())
	filters.ThreatIntelIndicatorType = expandSecurityHubStringFilters(in["threat_intel_indicator_type"].(*schema.Set).List())
	filters.ThreatIntelIndicatorValue = expandSecurityHubStringFilters(in["threat_intel_indicator_value"].(*schema.Set).List())
	filters.ThreatIntelIndicatorCategory = expandSecurityHubStringFilters(in["threat_intel_indicator_category"].(*schema.Set).List())
	filters.ThreatIntelIndicatorLastObservedAt = expandSecurityHubDateFilters(in["threat_intel_indicator_last_observed_at"].(*schema.Set).List())
	filters.ThreatIntelIndicatorSource = expandSecurityHubStringFilters(in["threat_intel_indicator_source"].(*schema.Set).List())
	filters.ThreatIntelIndicatorSourceUrl = expandSecurityHubStringFilters(in["threat_intel_indicator_source_url"].(*schema.Set).List())
	filters.ResourceType = expandSecurityHubStringFilters(in["resource_type"].(*schema.Set).List())
	filters.ResourceId = expandSecurityHubStringFilters(in["resource_id"].(*schema.Set).List())
	filters.ResourcePartition = expandSecurityHubStringFilters(in["resource_partition"].(*schema.Set).List())
	filters.ResourceRegion = expandSecurityHubStringFilters(in["resource_region"].(*schema.Set).List())
	filters.ResourceTags = expandSecurityHubMapFilters(in["resource_tags"].(*schema.Set).List())
	filters.ResourceAwsEc2InstanceType = expandSecurityHubStringFilters(in["resource_aws_ec2_instance_type"].(*schema.Set).List())
	filters.ResourceAwsEc2InstanceImageId = expandSecurityHubStringFilters(in["resource_aws_ec2_instance_image_id"].(*schema.Set).List())
	filters.ResourceAwsEc2InstanceIpV4Addresses = expandSecurityHubIpFilters(in["resource_aws_ec2_instance_ip_v4_addresses"].(*schema.Set).List())
	filters.ResourceAwsEc2InstanceIpV6Addresses = expandSecurityHubIpFilters(in["resource_aws_ec2_instance_ip_v6_addresses"].(*schema.Set).List())
	filters.ResourceAwsEc2InstanceKeyName = expandSecurityHubStringFilters(in["resource_aws_ec2_instance_key_name"].(*schema.Set).List())
	filters.ResourceAwsEc2InstanceIamInstanceProfileArn = expandSecurityHubStringFilters(in["resource_aws_ec2_instance_iam_instance_profile_arn"].(*schema.Set).List())
	filters.ResourceAwsEc2InstanceVpcId = expandSecurityHubStringFilters(in["resource_aws_ec2_instance_vpc_id"].(*schema.Set).List())
	filters.ResourceAwsEc2InstanceSubnetId = expandSecurityHubStringFilters(in["resource_aws_ec2_instance_subnet_id"].(*schema.Set).List())
	filters.ResourceAwsEc2InstanceLaunchedAt = expandSecurityHubDateFilters(in["resource_aws_ec2_instance_launched_at"].(*schema.Set).List())
	filters.ResourceAwsS3BucketOwnerId = expandSecurityHubStringFilters(in["resource_aws_s3_bucket_owner_id"].(*schema.Set).List())
	filters.ResourceAwsS3BucketOwnerName = expandSecurityHubStringFilters(in["resource_aws_s3_bucket_owner_name"].(*schema.Set).List())
	filters.ResourceAwsIamAccessKeyUserName = expandSecurityHubStringFilters(in["resource_aws_iam_access_key_user_name"].(*schema.Set).List())
	filters.ResourceAwsIamAccessKeyStatus = expandSecurityHubStringFilters(in["resource_aws_iam_access_key_status"].(*schema.Set).List())
	filters.ResourceAwsIamAccessKeyCreatedAt = expandSecurityHubDateFilters(in["resource_aws_iam_access_key_created_at"].(*schema.Set).List())
	filters.ResourceContainerName = expandSecurityHubStringFilters(in["resource_container_name"].(*schema.Set).List())
	filters.ResourceContainerImageId = expandSecurityHubStringFilters(in["resource_container_image_id"].(*schema.Set).List())
	filters.ResourceContainerImageName = expandSecurityHubStringFilters(in["resource_container_image_name"].(*schema.Set).List())
	filters.ResourceContainerLaunchedAt = expandSecurityHubDateFilters(in["resource_container_launched_at"].(*schema.Set).List())
	filters.ResourceDetailsOther = expandSecurityHubMapFilters(in["resource_details_other"].(*schema.Set).List())
	filters.ComplianceStatus = expandSecurityHubStringFilters(in["compliance_status"].(*schema.Set).List())
	filters.VerificationState = expandSecurityHubStringFilters(in["verification_state"].(*schema.Set).List())
	filters.WorkflowState = expandSecurityHubStringFilters(in["workflow_state"].(*schema.Set).List())
	filters.RecordState = expandSecurityHubStringFilters(in["record_state"].(*schema.Set).List())
	filters.RelatedFindingsProductArn = expandSecurityHubStringFilters(in["related_findings_product_arn"].(*schema.Set).List())
	filters.RelatedFindingsId = expandSecurityHubStringFilters(in["related_findings_id"].(*schema.Set).List())
	filters.NoteText = expandSecurityHubStringFilters(in["note_text"].(*schema.Set).List())
	filters.NoteUpdatedAt = expandSecurityHubDateFilters(in["note_updated_at"].(*schema.Set).List())
	filters.NoteUpdatedBy = expandSecurityHubStringFilters(in["note_updated_by"].(*schema.Set).List())
	filters.Keyword = expandSecurityHubKeywordFilters(in["keyword"].(*schema.Set).List())

	return filters
}

func flattenSecurityHubAwsSecurityFindingFilters(filters *securityhub.AwsSecurityFindingFilters) []map[string]interface{} {
	m := make(map[string]interface{}, 1)

	if filters.ResourceAwsEc2InstanceIpV4Addresses != nil {
		m["resource_aws_ec2_instance_ip_v4_addresses"] = flattenSecurityHubIpFilters(filters.ResourceAwsEc2InstanceIpV4Addresses)
	}

	var out = make([]map[string]interface{}, 1, 1)
	out[0] = m

	log.Printf("flattenSecurityHubAwsSecurityFindingFilters: %v", out)
	return out
}

func securityHubStringFilterSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"comparison": {
					Type:     schema.TypeString,
					Optional: true,
					ValidateFunc: validation.StringInSlice([]string{
						SecurityHubFilterComparisonEquals,
						SecurityHubFilterComparisonContains,
						SecurityHubFilterComparisonPrefix,
					}, false),
				},
				"value": {
					Type:     schema.TypeString,
					Optional: true,
				},
			},
		},
	}
}

func expandSecurityHubStringFilters(in []interface{}) []*securityhub.StringFilter {
	if in == nil || len(in) == 0 {
		return nil
	}

	var out = make([]*securityhub.StringFilter, 0)
	for _, mRaw := range in {
		if mRaw == nil {
			continue
		}
		m := mRaw.(map[string]interface{})
		filter := &securityhub.StringFilter{
			Comparison: aws.String(m["comparison"].(string)),
			Value:      aws.String(m["value"].(string)),
		}
		out = append(out, filter)
	}
	return out
}

func securityHubNumberFilterSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"eq": {
					Type:     schema.TypeFloat,
					Optional: true,
				},
				"gte": {
					Type:     schema.TypeFloat,
					Optional: true,
				},
				"lte": {
					Type:     schema.TypeFloat,
					Optional: true,
				},
			},
		},
	}
}

func expandSecurityHubNumberFilters(in []interface{}) []*securityhub.NumberFilter {
	if in == nil || len(in) == 0 {
		return nil
	}

	var out = make([]*securityhub.NumberFilter, 0)
	for _, mRaw := range in {
		if mRaw == nil {
			continue
		}
		m := mRaw.(map[string]interface{})
		filter := &securityhub.NumberFilter{}
		if m["eq"] != nil {
			filter.Eq = aws.Float64(m["eq"].(float64))
		}
		if m["gte"] != nil {
			filter.Gte = aws.Float64(m["gte"].(float64))
		}
		if m["lte"] != nil {
			filter.Lte = aws.Float64(m["lte"].(float64))
		}
		out = append(out, filter)
	}
	return out
}

func securityHubDateFilterSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"date_range": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"unit": {
								Type:     schema.TypeString,
								Optional: true,
								ValidateFunc: validation.StringInSlice([]string{
									SecurityHubFilterDateRangeUnitDays,
								}, false),
							},
							"value": {
								Type:     schema.TypeInt,
								Optional: true,
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

func expandSecurityHubDateFilters(in []interface{}) []*securityhub.DateFilter {
	if in == nil || len(in) == 0 {
		return nil
	}

	var out = make([]*securityhub.DateFilter, 0)
	for _, mRaw := range in {
		if mRaw == nil {
			continue
		}
		m := mRaw.(map[string]interface{})
		filter := &securityhub.DateFilter{}
		if m["date_range"] != nil {
			filter.DateRange = expandSecurityHubDateRange(m["date_range"].([]interface{}))
		}
		if m["end"] != nil {
			filter.End = aws.String(m["end"].(string))
		}
		if m["start"] != nil {
			filter.Start = aws.String(m["start"].(string))
		}
		out = append(out, filter)
	}
	return out
}

func expandSecurityHubDateRange(in []interface{}) *securityhub.DateRange {
	if in == nil || len(in) == 0 {
		return nil
	}

	m := in[0].(map[string]interface{})
	dateRange := &securityhub.DateRange{}
	if m["unit"] != nil {
		dateRange.Unit = aws.String(m["unit"].(string))
	}
	if m["value"] != nil {
		dateRange.Value = aws.Int64(int64(m["value"].(int)))
	}

	return dateRange
}

func securityHubKeywordFilterSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"value": {
					Type:     schema.TypeString,
					Optional: true,
				},
			},
		},
	}
}

func expandSecurityHubKeywordFilters(in []interface{}) []*securityhub.KeywordFilter {
	if in == nil || len(in) == 0 {
		return nil
	}

	var out = make([]*securityhub.KeywordFilter, 0)
	for _, mRaw := range in {
		if mRaw == nil {
			continue
		}
		m := mRaw.(map[string]interface{})
		filter := &securityhub.KeywordFilter{}
		if m["value"] != nil {
			filter.Value = aws.String(m["value"].(string))
		}
		out = append(out, filter)
	}
	return out
}

func securityHubIpFilterSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"cidr": {
					Type:     schema.TypeString,
					Optional: true,
				},
			},
		},
	}
}

func expandSecurityHubIpFilters(in []interface{}) []*securityhub.IpFilter {
	if in == nil || len(in) == 0 {
		return nil
	}

	var out = make([]*securityhub.IpFilter, 0)
	for _, mRaw := range in {
		if mRaw == nil {
			continue
		}
		m := mRaw.(map[string]interface{})
		filter := &securityhub.IpFilter{}
		if m["cidr"] != nil {
			filter.Cidr = aws.String(m["cidr"].(string))
		}
		out = append(out, filter)
	}
	return out
}

func flattenSecurityHubIpFilters(in []*securityhub.IpFilter) []map[string]interface{} {
	var out = make([]map[string]interface{}, len(in), len(in))
	for i, v := range in {
		m := make(map[string]interface{})
		if *v.Cidr != "" {
			m["cidr"] = *v.Cidr
		}
		out[i] = m
	}
	log.Printf("flattenSecurityHubIpFilters: %v", out)
	return out
}

func securityHubMapFilterSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"comparison": {
					Type:     schema.TypeString,
					Optional: true,
					ValidateFunc: validation.StringInSlice([]string{
						SecurityHubFilterComparisonContains,
					}, false),
				},
				"key": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"value": {
					Type:     schema.TypeString,
					Optional: true,
				},
			},
		},
	}
}

func expandSecurityHubMapFilters(in []interface{}) []*securityhub.MapFilter {
	if in == nil || len(in) == 0 {
		return nil
	}

	var out = make([]*securityhub.MapFilter, 0)
	for _, mRaw := range in {
		if mRaw == nil {
			continue
		}
		m := mRaw.(map[string]interface{})
		filter := &securityhub.MapFilter{}
		if m["comparison"] != nil {
			filter.Comparison = aws.String(m["comparison"].(string))
		}
		if m["key"] != nil {
			filter.Key = aws.String(m["key"].(string))
		}
		if m["value"] != nil {
			filter.Value = aws.String(m["value"].(string))
		}
		out = append(out, filter)
	}
	return out
}

func resourceAwsSecurityHubInsightCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).securityhubconn

	log.Printf("expandSecurityHubAwsSecurityFindingFilters: %v", d.Get("filter"))

	log.Print("[DEBUG] Enabling Security Hub insight")

	resp, err := conn.CreateInsight(&securityhub.CreateInsightInput{
		Name:             aws.String(d.Get("name").(string)),
		GroupByAttribute: aws.String(d.Get("group_by_attribute").(string)),
		Filters:          expandSecurityHubAwsSecurityFindingFilters(d),
	})

	if err != nil {
		return fmt.Errorf("Error creating Security Hub insight: %s", err)
	}

	d.SetId(*resp.InsightArn)

	return resourceAwsSecurityHubInsightRead(d, meta)
}

func resourceAwsSecurityHubInsightRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).securityhubconn

	log.Printf("[DEBUG] Reading Security Hub insight %s", d.Id())
	resp, err := conn.GetInsights(&securityhub.GetInsightsInput{
		InsightArns: []*string{aws.String(d.Id())},
	})

	if err != nil {
		if isAWSErr(err, securityhub.ErrCodeResourceNotFoundException, "") {
			log.Printf("[WARN] Security Hub insight (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error reading Security Hub insight %s: %s", d.Id(), err)
	}

	insight := resp.Insights[0]

	if insight == nil {
		log.Printf("[WARN] Security Hub insight (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("name", insight.Name)
	d.Set("group_by_attribute", insight.GroupByAttribute)
	d.Set("filters", flattenSecurityHubAwsSecurityFindingFilters(insight.Filters))

	return nil
}

func resourceAwsSecurityHubInsightUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).securityhubconn

	log.Printf("expandSecurityHubAwsSecurityFindingFilters: %v", d.Get("filter"))

	log.Printf("[DEBUG] Updating Security Hub insight %s", d.Id())
	_, err := conn.UpdateInsight(&securityhub.UpdateInsightInput{
		InsightArn:       aws.String(d.Id()),
		Name:             aws.String(d.Get("name").(string)),
		GroupByAttribute: aws.String(d.Get("group_by_attribute").(string)),
		Filters:          expandSecurityHubAwsSecurityFindingFilters(d),
	})

	if err != nil {
		return fmt.Errorf("Error updating Security Hub insight %s: %s", d.Id(), err)
	}

	return nil
}

func resourceAwsSecurityHubInsightDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).securityhubconn
	log.Printf("[DEBUG] Deleting Security Hub insight %s", d.Id())

	_, err := conn.DeleteInsight(&securityhub.DeleteInsightInput{
		InsightArn: aws.String(d.Id()),
	})

	if err != nil {
		if !isAWSErr(err, securityhub.ErrCodeResourceNotFoundException, "") {
			log.Printf("[WARN] Security Hub insight (%s) not found", d.Id())
			return nil
		}
		return fmt.Errorf("Error deleting Security Hub insight %s: %s", d.Id(), err)
	}

	return nil
}
