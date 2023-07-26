// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package networkfirewall

import (
	"context"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/networkfirewall"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKDataSource("aws_networkfirewall_firewall")
func DataSourceFirewall() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceFirewallResourceRead,
		Schema: map[string]*schema.Schema{
			"arn": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: verify.ValidARN,
				AtLeastOneOf: []string{"arn", "name"},
			},
			"delete_protection": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"encryption_configuration": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"key_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"type": {
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
						"status": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"sync_states": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"availability_zone": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"attachment": {
										Type:     schema.TypeList,
										Computed: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"endpoint_id": {
													Type:     schema.TypeString,
													Computed: true,
												},
												"status": {
													Type:     schema.TypeString,
													Computed: true,
												},
												"subnet_id": {
													Type:     schema.TypeString,
													Computed: true,
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				AtLeastOneOf: []string{"arn", "name"},
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9-]{1,128}$`), "Must have 1-128 valid characters: a-z, A-Z, 0-9 and -(hyphen)"),
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
						"subnet_id": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"tags": tftags.TagsSchema(),
			"update_token": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"vpc_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceFirewallResourceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).NetworkFirewallConn(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	input := &networkfirewall.DescribeFirewallInput{}

	if v, ok := d.GetOk("arn"); ok {
		input.FirewallArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("name"); ok {
		input.FirewallName = aws.String(v.(string))
	}

	if input.FirewallArn == nil && input.FirewallName == nil {
		return diag.Errorf("must specify either arn, name, or both")
	}

	output, err := conn.DescribeFirewallWithContext(ctx, input)
	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, networkfirewall.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] NetworkFirewall Firewall (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading NetworkFirewall Firewall (%s): %s", d.Id(), err)
	}

	if output == nil || output.Firewall == nil {
		return diag.Errorf("reading NetworkFirewall Firewall (%s): empty output", d.Id())
	}

	firewall := output.Firewall

	d.Set("arn", firewall.FirewallArn)
	d.Set("delete_protection", firewall.DeleteProtection)
	d.Set("description", firewall.Description)
	d.Set("name", firewall.FirewallName)
	d.Set("encryption_configuration", flattenDataSourceEncryptionConfiguration(firewall.EncryptionConfiguration))
	d.Set("firewall_policy_arn", firewall.FirewallPolicyArn)
	d.Set("firewall_policy_change_protection", firewall.FirewallPolicyChangeProtection)
	d.Set("firewall_status", flattenDataSourceFirewallStatus(output.FirewallStatus))
	d.Set("subnet_change_protection", firewall.SubnetChangeProtection)
	d.Set("update_token", output.UpdateToken)
	d.Set("vpc_id", firewall.VpcId)

	if err := d.Set("subnet_mapping", flattenDataSourceSubnetMappings(firewall.SubnetMappings)); err != nil {
		return diag.Errorf("setting subnet_mappings: %s", err)
	}

	if err := d.Set("tags", KeyValueTags(ctx, firewall.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return diag.Errorf("setting tags: %s", err)
	}

	d.SetId(aws.StringValue(firewall.FirewallArn))

	return nil
}

func flattenDataSourceFirewallStatus(status *networkfirewall.FirewallStatus) []interface{} {
	if status == nil {
		return nil
	}
	m := map[string]interface{}{}
	if status.CapacityUsageSummary != nil {
		m["capacity_usage_summary"] = flattenDataSourceCapacityUsageSummary(status.CapacityUsageSummary)
	}
	if status.ConfigurationSyncStateSummary != nil {
		m["configuration_sync_state_summary"] = aws.StringValue(status.ConfigurationSyncStateSummary)
	}
	if status.Status != nil {
		m["status"] = aws.StringValue(status.Status)
	}
	if status.SyncStates != nil {
		m["sync_states"] = flattenDataSourceSyncStates(status.SyncStates)
	}

	return []interface{}{m}
}

func flattenDataSourceCapacityUsageSummary(state *networkfirewall.CapacityUsageSummary) []interface{} {
	if state == nil {
		return nil
	}

	m := map[string]interface{}{
		"cidrs": flattenDataSourceCIDRSummary(state.CIDRs),
	}

	return []interface{}{m}
}

func flattenDataSourceCIDRSummary(state *networkfirewall.CIDRSummary) []interface{} {
	if state == nil {
		return nil
	}

	m := map[string]interface{}{
		"available_cidr_count": int(aws.Int64Value(state.AvailableCIDRCount)),
		"ip_set_references":    flattenDataSourceIPSetReferences(state.IPSetReferences),
		"utilized_cidr_count":  int(aws.Int64Value(state.UtilizedCIDRCount)),
	}

	return []interface{}{m}
}

func flattenDataSourceIPSetReferences(state map[string]*networkfirewall.IPSetMetadata) []interface{} {
	if state == nil {
		return nil
	}

	ipSetReferences := make([]interface{}, 0, len(state))
	for _, v := range state {
		m := map[string]interface{}{
			"resolved_cidr_count": int(aws.Int64Value(v.ResolvedCIDRCount)),
		}
		ipSetReferences = append(ipSetReferences, m)
	}

	return ipSetReferences
}

func flattenDataSourceSyncStates(state map[string]*networkfirewall.SyncState) []interface{} {
	if state == nil {
		return nil
	}

	syncStates := make([]interface{}, 0, len(state))
	for k, v := range state {
		m := map[string]interface{}{
			"availability_zone": k,
			"attachment":        flattenDataSourceSyncStateAttachment(v.Attachment),
		}
		syncStates = append(syncStates, m)
	}

	return syncStates
}

func flattenDataSourceSyncStateAttachment(attach *networkfirewall.Attachment) []interface{} {
	if attach == nil {
		return nil
	}

	m := map[string]interface{}{
		"endpoint_id": aws.StringValue(attach.EndpointId),
		"status":      aws.StringValue(attach.Status),
		"subnet_id":   aws.StringValue(attach.SubnetId),
	}

	return []interface{}{m}
}

func flattenDataSourceSubnetMappings(subnet []*networkfirewall.SubnetMapping) []interface{} {
	mappings := make([]interface{}, 0, len(subnet))
	for _, s := range subnet {
		m := map[string]interface{}{
			"subnet_id": aws.StringValue(s.SubnetId),
		}
		mappings = append(mappings, m)
	}

	return mappings
}

func flattenDataSourceEncryptionConfiguration(encrypt *networkfirewall.EncryptionConfiguration) []interface{} {
	if encrypt == nil {
		return nil
	}

	m := map[string]interface{}{
		"key_id": aws.StringValue(encrypt.KeyId),
		"type":   aws.StringValue(encrypt.Type),
	}

	return []interface{}{m}
}
