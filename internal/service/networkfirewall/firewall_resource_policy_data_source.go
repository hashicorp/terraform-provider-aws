package networkfirewall

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func DataSourceFirewallResourcePolicy() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceFirewallResourcePolicyRead,
		Schema: map[string]*schema.Schema{
			"resource_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"policy": {
				// No validation needed since policy JSON is returned from call
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceFirewallResourcePolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).NetworkFirewallConn

	resourceArn := d.Get("resource_arn").(string)

	log.Printf("[DEBUG] Reading NetworkFirewall Resource Policy for resource: %s", resourceArn)

	policy, err := FindResourcePolicy(ctx, conn, resourceArn)

	if err != nil {
		return diag.Errorf("reading NetworkFirewall Resource Policy (for resource: %s): %w", resourceArn, err)
	}

	if policy == nil {
		return diag.Errorf("reading NetworkFirewall Resource Policy (for resource: %s): empty output", resourceArn)
	}

	// Id is identical to the resource ARN
	d.SetId(resourceArn)
	d.Set("resource_arn", resourceArn)
	d.Set("policy", policy)

	return nil
}
