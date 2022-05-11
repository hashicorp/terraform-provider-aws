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
				Type: schema.TypeString,
				//Computed: true,
				// Assuming ARN is the only acceptable input (it's not)
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
				// Optional: true,
				// Attributes are being fetched by the API call
			},
			"firewall_policy": {
				Type:     schema.TypeList,
				Computed: true,
				// Required: true,
				// Not passing in items  - MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"stateful_default_actions": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"stateful_engine_options": {
							Type: schema.TypeList,
							//MaxItems: 1,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"rule_order": {
										Type: schema.TypeString,
										// Required:     true,
										Computed: true,
										// API is returning already valid data
										// ValidateFunc: validation.StringInSlice(networkfirewall.RuleOrder_Values(), false),
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
										//ValidateFunc: validation.IntAtLeast(1),
									},
									"resource_arn": {
										Type:     schema.TypeString,
										Computed: true,
										//ValidateFunc: verify.ValidARN,
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
										//ValidateFunc: validation.IntAtLeast(1),
									},
									"resource_arn": {
										Type:     schema.TypeString,
										Computed: true,
										//ValidateFunc: verify.ValidARN,
									},
								},
							},
						},
					},
				},
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
				// ForceNew: true,
			},
			// "tags":     tftags.TagsSchema(),
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
	// Don't need to exclude default tags in the data source read
	// defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	arn := d.Get("arn").(string)

	log.Printf("[DEBUG] Reading NetworkFirewall Firewall Policy %s", arn)

	output, err := FindFirewallPolicy(ctx, conn, arn)
	// if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, networkfirewall.ErrCodeResourceNotFoundException) {
	// 	log.Printf("[WARN] NetworkFirewall Firewall Policy (%s) not found, removing from state", arn)
	// 	d.SetId("")
	// 	return nil
	// }
	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading NetworkFirewall Firewall Policy (%s): %w", arn, err))
	}

	if output == nil {
		return diag.FromErr(fmt.Errorf("error reading NetworkFirewall Firewall Policy (%s): empty output", arn))
	}
	if output.FirewallPolicyResponse == nil {
		return diag.FromErr(fmt.Errorf("error reading NetworkFirewall Firewall Policy (%s): empty output.FirewallPolicyResponse", arn))
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

	//lintignore:AWSR002
	// if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
	// 	return diag.FromErr(fmt.Errorf("error setting tags: %w", err))
	// }
	//
	if err := d.Set("tags", tags.Map()); err != nil {
		return diag.FromErr(fmt.Errorf("error setting tags: %w", err))
	}

	return nil
}
