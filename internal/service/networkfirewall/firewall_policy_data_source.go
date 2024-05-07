// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package networkfirewall

import (
	"context"
	"log"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_networkfirewall_firewall_policy")
func DataSourceFirewallPolicy() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceFirewallPolicyRead,
		Schema: map[string]*schema.Schema{
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
												"action": {
													Type:     schema.TypeString,
													Optional: true,
												},
											},
										},
									},
									"priority": {
										Type:     schema.TypeInt,
										Computed: true,
									},
									"resource_arn": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
						"stateless_custom_action": customActionSchemaDataSource(),
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
									"priority": {
										Type:     schema.TypeInt,
										Computed: true,
									},
									"resource_arn": {
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
		},
	}
}

func dataSourceFirewallPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).NetworkFirewallConn(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	arn := d.Get(names.AttrARN).(string)
	name := d.Get(names.AttrName).(string)

	log.Printf("[DEBUG] Reading NetworkFirewall Firewall Policy %s %s", arn, name)

	output, err := FindFirewallPolicyByNameAndARN(ctx, conn, arn, name)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading NetworkFirewall Firewall Policy (%s, %s): %s", arn, name, err)
	}

	if output == nil {
		return sdkdiag.AppendErrorf(diags, "reading NetworkFirewall Firewall Policy (%s, %s): empty output", arn, name)
	}
	if output.FirewallPolicyResponse == nil {
		return sdkdiag.AppendErrorf(diags, "reading NetworkFirewall Firewall Policy (%s, %s): empty output.FirewallPolicyResponse", arn, name)
	}

	resp := output.FirewallPolicyResponse
	policy := output.FirewallPolicy

	d.SetId(aws.StringValue(resp.FirewallPolicyArn))

	d.Set(names.AttrARN, resp.FirewallPolicyArn)
	d.Set(names.AttrDescription, resp.Description)
	d.Set(names.AttrName, resp.FirewallPolicyName)
	d.Set("update_token", output.UpdateToken)

	if err := d.Set("firewall_policy", flattenFirewallPolicy(policy)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting firewall_policy: %s", err)
	}

	tags := KeyValueTags(ctx, resp.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	if err := d.Set(names.AttrTags, tags.Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	return diags
}
