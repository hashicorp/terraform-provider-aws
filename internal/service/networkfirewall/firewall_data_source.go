// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package networkfirewall

import (
	"context"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/networkfirewall"
	awstypes "github.com/aws/aws-sdk-go-v2/service/networkfirewall/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_networkfirewall_firewall", name="Firewall")
// @Tags
func dataSourceFirewall() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceFirewallResourceRead,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: verify.ValidARN,
				AtLeastOneOf: []string{names.AttrARN, names.AttrName},
			},
			"delete_protection": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrEncryptionConfiguration: {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrKeyID: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrType: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"firewall_policy_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"firewall_policy_change_protection": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"firewall_status": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"capacity_usage_summary": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"cidrs": {
										Type:     schema.TypeSet,
										Computed: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"available_cidr_count": {
													Type:     schema.TypeInt,
													Computed: true,
												},
												"ip_set_references": {
													Type:     schema.TypeSet,
													Computed: true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"resolved_cidr_count": {
																Type:     schema.TypeInt,
																Computed: true,
															},
														},
													},
												},
												"utilized_cidr_count": {
													Type:     schema.TypeInt,
													Computed: true,
												},
											},
										},
									},
								},
							},
						},
						"configuration_sync_state_summary": {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrStatus: {
							Type:     schema.TypeString,
							Computed: true,
						},
						"sync_states": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"attachment": {
										Type:     schema.TypeList,
										Computed: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"endpoint_id": {
													Type:     schema.TypeString,
													Computed: true,
												},
												names.AttrStatus: {
													Type:     schema.TypeString,
													Computed: true,
												},
												names.AttrSubnetID: {
													Type:     schema.TypeString,
													Computed: true,
												},
											},
										},
									},
									names.AttrAvailabilityZone: {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
					},
				},
			},
			names.AttrName: {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				AtLeastOneOf: []string{names.AttrARN, names.AttrName},
				ValidateFunc: validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z-]{1,128}$`), "Must have 1-128 valid characters: a-z, A-Z, 0-9 and -(hyphen)"),
			},
			"subnet_change_protection": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"subnet_mapping": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrSubnetID: {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
			"update_token": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrVPCID: {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceFirewallResourceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).NetworkFirewallClient(ctx)

	input := &networkfirewall.DescribeFirewallInput{}
	if v, ok := d.GetOk(names.AttrARN); ok {
		input.FirewallArn = aws.String(v.(string))
	}
	if v, ok := d.GetOk(names.AttrName); ok {
		input.FirewallName = aws.String(v.(string))
	}

	output, err := findFirewall(ctx, conn, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading NetworkFirewall Firewall: %s", err)
	}

	firewall := output.Firewall
	d.SetId(aws.ToString(firewall.FirewallArn))
	d.Set(names.AttrARN, firewall.FirewallArn)
	d.Set("delete_protection", firewall.DeleteProtection)
	d.Set(names.AttrDescription, firewall.Description)
	if err := d.Set(names.AttrEncryptionConfiguration, flattenDataSourceEncryptionConfiguration(firewall.EncryptionConfiguration)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting encryption_configuration: %s", err)
	}
	d.Set("firewall_policy_arn", firewall.FirewallPolicyArn)
	d.Set("firewall_policy_change_protection", firewall.FirewallPolicyChangeProtection)
	if err := d.Set("firewall_status", flattenDataSourceFirewallStatus(output.FirewallStatus)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting firewall_status: %s", err)
	}
	d.Set(names.AttrName, firewall.FirewallName)
	d.Set("subnet_change_protection", firewall.SubnetChangeProtection)
	if err := d.Set("subnet_mapping", flattenDataSourceSubnetMappings(firewall.SubnetMappings)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting subnet_mappings: %s", err)
	}
	d.Set("update_token", output.UpdateToken)
	d.Set(names.AttrVPCID, firewall.VpcId)

	setTagsOut(ctx, firewall.Tags)

	return diags
}

func flattenDataSourceFirewallStatus(apiObject *awstypes.FirewallStatus) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"configuration_sync_state_summary": apiObject.ConfigurationSyncStateSummary,
		names.AttrStatus:                   apiObject.Status,
	}

	if apiObject.CapacityUsageSummary != nil {
		tfMap["capacity_usage_summary"] = flattenDataSourceCapacityUsageSummary(apiObject.CapacityUsageSummary)
	}
	if apiObject.SyncStates != nil {
		tfMap["sync_states"] = flattenDataSourceSyncStates(apiObject.SyncStates)
	}

	return []interface{}{tfMap}
}

func flattenDataSourceCapacityUsageSummary(apiObject *awstypes.CapacityUsageSummary) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"cidrs": flattenDataSourceCIDRSummary(apiObject.CIDRs),
	}

	return []interface{}{tfMap}
}

func flattenDataSourceCIDRSummary(apiObject *awstypes.CIDRSummary) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"available_cidr_count": aws.ToInt32(apiObject.AvailableCIDRCount),
		"ip_set_references":    flattenDataSourceIPSetReferences(apiObject.IPSetReferences),
		"utilized_cidr_count":  aws.ToInt32(apiObject.UtilizedCIDRCount),
	}

	return []interface{}{tfMap}
}

func flattenDataSourceIPSetReferences(apiObject map[string]awstypes.IPSetMetadata) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfList := make([]interface{}, 0, len(apiObject))

	for _, v := range apiObject {
		tfList = append(tfList, map[string]interface{}{
			"resolved_cidr_count": aws.ToInt32(v.ResolvedCIDRCount),
		})
	}

	return tfList
}

func flattenDataSourceSyncStates(apiObject map[string]awstypes.SyncState) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfList := make([]interface{}, 0, len(apiObject))

	for k, v := range apiObject {
		tfMap := map[string]interface{}{
			"attachment":               flattenDataSourceAttachment(v.Attachment),
			names.AttrAvailabilityZone: k,
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenDataSourceAttachment(apiObject *awstypes.Attachment) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"endpoint_id":      aws.ToString(apiObject.EndpointId),
		names.AttrStatus:   apiObject.Status,
		names.AttrSubnetID: aws.ToString(apiObject.SubnetId),
	}

	return []interface{}{tfMap}
}

func flattenDataSourceSubnetMappings(apiObjects []awstypes.SubnetMapping) []interface{} {
	tfList := make([]interface{}, 0, len(apiObjects))

	for _, s := range apiObjects {
		tfMap := map[string]interface{}{
			names.AttrSubnetID: aws.ToString(s.SubnetId),
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenDataSourceEncryptionConfiguration(apiObject *awstypes.EncryptionConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		names.AttrKeyID: aws.ToString(apiObject.KeyId),
		names.AttrType:  apiObject.Type,
	}

	return []interface{}{tfMap}
}
