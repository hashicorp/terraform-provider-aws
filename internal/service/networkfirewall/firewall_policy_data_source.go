package networkfirewall

import (
	"context"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func DataSourceFirewallPolicy() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceFirewallPolicyRead,
		Schema: map[string]*schema.Schema{
			"arn": {
				Type:         schema.TypeString,
				AtLeastOneOf: []string{"arn", "name"},
				Optional:     true,
				ValidateFunc: verify.ValidARN,
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
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9-]{1,128}$`), "Must have 1-128 valid characters: a-z, A-Z, 0-9 and -(hyphen)"),
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
	conn := meta.(*conns.AWSClient).NetworkFirewallConn()
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	arn := d.Get("arn").(string)
	name := d.Get("name").(string)

	log.Printf("[DEBUG] Reading NetworkFirewall Firewall Policy %s %s", arn, name)

	output, err := FindFirewallPolicyByNameAndARN(ctx, conn, arn, name)

	if err != nil {
		return diag.Errorf("reading NetworkFirewall Firewall Policy (%s, %s): %s", arn, name, err)
	}

	if output == nil {
		return diag.Errorf("reading NetworkFirewall Firewall Policy (%s, %s): empty output", arn, name)
	}
	if output.FirewallPolicyResponse == nil {
		return diag.Errorf("reading NetworkFirewall Firewall Policy (%s, %s): empty output.FirewallPolicyResponse", arn, name)
	}

	resp := output.FirewallPolicyResponse
	policy := output.FirewallPolicy

	d.SetId(aws.StringValue(resp.FirewallPolicyArn))

	d.Set("arn", resp.FirewallPolicyArn)
	d.Set("description", resp.Description)
	d.Set("name", resp.FirewallPolicyName)
	d.Set("update_token", output.UpdateToken)

	if err := d.Set("firewall_policy", flattenFirewallPolicy(policy)); err != nil {
		return diag.Errorf("setting firewall_policy: %s", err)
	}

	tags := KeyValueTags(resp.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	if err := d.Set("tags", tags.Map()); err != nil {
		return diag.Errorf("setting tags: %s", err)
	}

	return nil
}
