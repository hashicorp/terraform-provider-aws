// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package networkfirewall

import (
	"context"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/networkfirewall"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/sdkv2"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_networkfirewall_firewall_policy", name="Firewall Policy")
// @Tags
func dataSourceFirewallPolicy() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceFirewallPolicyRead,

		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				names.AttrARN: {
					Type:         schema.TypeString,
					AtLeastOneOf: []string{names.AttrARN, names.AttrName},
					Optional:     true,
					ValidateFunc: verify.ValidARN,
				},
				names.AttrDescription: {
					Type:     schema.TypeString,
					Computed: true,
				},
				"firewall_policy": {
					Type:     schema.TypeList,
					Computed: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"stateful_default_actions": {
								Type:     schema.TypeSet,
								Computed: true,
								Elem:     &schema.Schema{Type: schema.TypeString},
							},
							"stateful_engine_options": {
								Type:     schema.TypeList,
								Computed: true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"rule_order": {
											Type:     schema.TypeString,
											Computed: true,
										},
										"stream_exception_policy": {
											Type:     schema.TypeString,
											Computed: true,
										},
									},
								},
							},
							"stateful_rule_group_reference": {
								Type:     schema.TypeSet,
								Computed: true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"override": {
											Type:     schema.TypeList,
											Optional: true,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													names.AttrAction: {
														Type:     schema.TypeString,
														Optional: true,
													},
												},
											},
										},
										names.AttrPriority: {
											Type:     schema.TypeInt,
											Computed: true,
										},
										names.AttrResourceARN: {
											Type:     schema.TypeString,
											Computed: true,
										},
									},
								},
							},
							"stateless_custom_action": sdkv2.DataSourcePropertyFromResourceProperty(customActionSchema()),
							"stateless_default_actions": {
								Type:     schema.TypeSet,
								Computed: true,
								Elem:     &schema.Schema{Type: schema.TypeString},
							},
							"stateless_fragment_default_actions": {
								Type:     schema.TypeSet,
								Computed: true,
								Elem:     &schema.Schema{Type: schema.TypeString},
							},
							"stateless_rule_group_reference": {
								Type:     schema.TypeSet,
								Computed: true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										names.AttrPriority: {
											Type:     schema.TypeInt,
											Computed: true,
										},
										names.AttrResourceARN: {
											Type:     schema.TypeString,
											Computed: true,
										},
									},
								},
							},
							"tls_inspection_configuration_arn": {
								Type:     schema.TypeString,
								Computed: true,
							},
						},
					},
				},
				names.AttrName: {
					Type:         schema.TypeString,
					Optional:     true,
					AtLeastOneOf: []string{names.AttrARN, names.AttrName},
					ValidateFunc: validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z-]{1,128}$`), "Must have 1-128 valid characters: a-z, A-Z, 0-9 and -(hyphen)"),
				},
				names.AttrTags: tftags.TagsSchemaComputed(),
				"update_token": {
					Type:     schema.TypeString,
					Computed: true,
				},
			}
		},
	}
}

func dataSourceFirewallPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).NetworkFirewallClient(ctx)

	input := &networkfirewall.DescribeFirewallPolicyInput{}
	if v := d.Get(names.AttrARN).(string); v != "" {
		input.FirewallPolicyArn = aws.String(v)
	}
	if v := d.Get(names.AttrName).(string); v != "" {
		input.FirewallPolicyName = aws.String(v)
	}

	output, err := findFirewallPolicy(ctx, conn, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading NetworkFirewall Firewall Policy: %s", err)
	}

	resp := output.FirewallPolicyResponse

	d.SetId(aws.ToString(resp.FirewallPolicyArn))
	d.Set(names.AttrARN, resp.FirewallPolicyArn)
	d.Set(names.AttrDescription, resp.Description)
	if err := d.Set("firewall_policy", flattenFirewallPolicy(output.FirewallPolicy)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting firewall_policy: %s", err)
	}
	d.Set(names.AttrName, resp.FirewallPolicyName)
	d.Set("update_token", output.UpdateToken)

	setTagsOut(ctx, resp.Tags)

	return diags
}
