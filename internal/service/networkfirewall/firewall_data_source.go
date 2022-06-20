package networkfirewall

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/networkfirewall"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func DataSourceFirewall() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceFirewallResourceRead,
		Schema: map[string]*schema.Schema{
			"resource_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			"delete_protection": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"firewall_policy_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"firewall_policy_change_protection": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"firewall_status": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
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
				Type:     schema.TypeString,
				Computed: true,
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
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
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
	conn := meta.(*conns.AWSClient).NetworkFirewallConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	resourceArn := d.Get("resource_arn").(string)

	log.Printf("[DEBUG] Reading NetworkFirewall Firewall %s", resourceArn)

	input := &networkfirewall.DescribeFirewallInput{
		FirewallArn: aws.String(resourceArn),
	}
	output, err := conn.DescribeFirewallWithContext(ctx, input)
	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, networkfirewall.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] NetworkFirewall Firewall (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading NetworkFirewall Firewall (%s): %w", d.Id(), err))
	}

	if output == nil || output.Firewall == nil {
		return diag.FromErr(fmt.Errorf("error reading NetworkFirewall Firewall (%s): empty output", d.Id()))
	}

	firewall := output.Firewall

	d.SetId(resourceArn)
	d.Set("delete_protection", firewall.DeleteProtection)
	d.Set("description", firewall.Description)
	d.Set("name", firewall.FirewallName)
	d.Set("firewall_policy_arn", firewall.FirewallPolicyArn)
	d.Set("firewall_policy_change_protection", firewall.FirewallPolicyChangeProtection)
	d.Set("firewall_status", flattenDataSourceFirewallStatus(output.FirewallStatus))
	d.Set("subnet_change_protection", firewall.SubnetChangeProtection)
	d.Set("update_token", output.UpdateToken)
	d.Set("vpc_id", firewall.VpcId)

	if err := d.Set("subnet_mapping", flattenDataSourceSubnetMappings(firewall.SubnetMappings)); err != nil {
		return diag.FromErr(fmt.Errorf("error setting subnet_mappings: %w", err))
	}

	tags := KeyValueTags(firewall.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return diag.FromErr(fmt.Errorf("error setting tags: %w", err))
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return diag.FromErr(fmt.Errorf("error setting tags_all: %w", err))
	}

	return nil
}

func flattenDataSourceFirewallStatus(status *networkfirewall.FirewallStatus) []interface{} {
	if status == nil {
		return nil
	}

	m := map[string]interface{}{
		"sync_states": flattenDataSourceSyncStates(status.SyncStates),
	}

	return []interface{}{m}
}

func flattenDataSourceSyncStates(s map[string]*networkfirewall.SyncState) []interface{} {
	if s == nil {
		return nil
	}

	syncStates := make([]interface{}, 0, len(s))
	for k, v := range s {
		m := map[string]interface{}{
			"availability_zone": k,
			"attachment":        flattenDataSourceSyncStateAttachment(v.Attachment),
		}
		syncStates = append(syncStates, m)
	}

	return syncStates
}

func flattenDataSourceSyncStateAttachment(a *networkfirewall.Attachment) []interface{} {
	if a == nil {
		return nil
	}

	m := map[string]interface{}{
		"endpoint_id": aws.StringValue(a.EndpointId),
		"subnet_id":   aws.StringValue(a.SubnetId),
	}

	return []interface{}{m}
}

func flattenDataSourceSubnetMappings(sm []*networkfirewall.SubnetMapping) []interface{} {
	mappings := make([]interface{}, 0, len(sm))
	for _, s := range sm {
		m := map[string]interface{}{
			"subnet_id": aws.StringValue(s.SubnetId),
		}
		mappings = append(mappings, m)
	}

	return mappings
}
