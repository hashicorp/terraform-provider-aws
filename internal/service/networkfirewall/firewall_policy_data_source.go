package networkfirewall

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

func DataSourceFirewallPolicy() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceFirewallPolicyRead,
		Schema: map[string]*schema.Schema{
			"arn": {
<<<<<<< HEAD
				Type:         schema.TypeString,
=======
				Type: schema.TypeString,
				//Computed: true,
				// Assuming ARN is the only acceptable input (it's not)
>>>>>>> 7505fe3b1a (Add changelog, WIP support for name input)
				AtLeastOneOf: []string{"arn", "name"},
				Optional:     true,
			},
			"description": {
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
								},
							},
						},
						"stateful_rule_group_reference": {
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
					},
				},
			},
			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				AtLeastOneOf: []string{"arn", "name"},
			},
			"tags": tftags.TagsSchemaComputed(),
			"update_token": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceFirewallPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).NetworkFirewallConn
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	arn := d.Get("arn").(string)
	name := d.Get("name").(string)

	log.Printf("[DEBUG] Reading NetworkFirewall Firewall Policy %s %s", arn, name)

	output, err := FindFirewallPolicyByNameAndArn(ctx, conn, arn, name)

	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading NetworkFirewall Firewall Policy (%s, %s): %w", arn, name, err))
	}

	if output == nil {
		return diag.FromErr(fmt.Errorf("error reading NetworkFirewall Firewall Policy (%s, %s): empty output", arn, name))
	}
	if output.FirewallPolicyResponse == nil {
		return diag.FromErr(fmt.Errorf("error reading NetworkFirewall Firewall Policy (%s, %s): empty output.FirewallPolicyResponse", arn, name))
	}

	resp := output.FirewallPolicyResponse
	policy := output.FirewallPolicy

	d.SetId(aws.StringValue(resp.FirewallPolicyArn))

	d.Set("arn", resp.FirewallPolicyArn)
	d.Set("description", resp.Description)
	d.Set("name", resp.FirewallPolicyName)
	d.Set("update_token", output.UpdateToken)

	if err := d.Set("firewall_policy", flattenFirewallPolicy(policy)); err != nil {
		return diag.FromErr(fmt.Errorf("error setting firewall_policy: %w", err))
	}

	tags := KeyValueTags(resp.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	if err := d.Set("tags", tags.Map()); err != nil {
		return diag.FromErr(fmt.Errorf("error setting tags: %w", err))
	}

	return nil
}
