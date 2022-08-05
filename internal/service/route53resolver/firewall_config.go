package route53resolver

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53resolver"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourceFirewallConfig() *schema.Resource {
	return &schema.Resource{
		Create: resourceFirewallConfigCreate,
		Read:   resourceFirewallConfigRead,
		Update: resourceFirewallConfigUpdate,
		Delete: resourceFirewallConfigDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"owner_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"resource_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"firewall_fail_open": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice(route53resolver.FirewallFailOpenStatus_Values(), false),
			},
		},
	}
}

func resourceFirewallConfigCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).Route53ResolverConn

	input := &route53resolver.UpdateFirewallConfigInput{
		ResourceId: aws.String(d.Get("resource_id").(string)),
	}

	if v, ok := d.GetOk("firewall_fail_open"); ok {
		input.FirewallFailOpen = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating Route 53 Resolver DNS Firewall config: %#v", input)
	output, err := conn.UpdateFirewallConfig(input)
	if err != nil {
		return fmt.Errorf("error creating Route 53 Resolver DNS Firewall config: %w", err)
	}

	d.SetId(aws.StringValue(output.FirewallConfig.Id))

	return resourceFirewallConfigRead(d, meta)
}

func resourceFirewallConfigRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).Route53ResolverConn

	config, err := FindFirewallConfigByID(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Route 53 Resolver DNS Firewall config (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error getting Route 53 Resolver DNS Firewall config (%s): %w", d.Id(), err)
	}

	d.Set("owner_id", config.OwnerId)
	d.Set("resource_id", config.ResourceId)
	d.Set("firewall_fail_open", config.FirewallFailOpen)

	return nil
}

func resourceFirewallConfigUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).Route53ResolverConn

	input := &route53resolver.UpdateFirewallConfigInput{
		ResourceId: aws.String(d.Get("resource_id").(string)),
	}

	if v, ok := d.GetOk("firewall_fail_open"); ok {
		input.FirewallFailOpen = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Updating Route 53 Resolver DNS Firewall config: %#v", input)
	_, err := conn.UpdateFirewallConfig(input)
	if err != nil {
		return fmt.Errorf("error creating Route 53 Resolver DNS Firewall config: %w", err)
	}

	return resourceFirewallConfigRead(d, meta)
}

func resourceFirewallConfigDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).Route53ResolverConn

	log.Printf("[DEBUG] Deleting Route 53 Resolver DNS Firewall config")
	_, err := conn.UpdateFirewallConfig(&route53resolver.UpdateFirewallConfigInput{
		ResourceId:       aws.String(d.Get("resource_id").(string)),
		FirewallFailOpen: aws.String(route53resolver.FirewallFailOpenStatusDisabled),
	})

	if err != nil {
		return fmt.Errorf("error deleting Route 53 Resolver DNS Firewall config (%s): %w", d.Id(), err)
	}

	return nil
}
