package route53resolver

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53resolver"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func ResourceQueryLogConfigAssociation() *schema.Resource {
	return &schema.Resource{
		Create: resourceQueryLogConfigAssociationCreate,
		Read:   resourceQueryLogConfigAssociationRead,
		Delete: resourceQueryLogConfigAssociationDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"resolver_query_log_config_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"resource_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceQueryLogConfigAssociationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).Route53ResolverConn

	input := &route53resolver.AssociateResolverQueryLogConfigInput{
		ResolverQueryLogConfigId: aws.String(d.Get("resolver_query_log_config_id").(string)),
		ResourceId:               aws.String(d.Get("resource_id").(string)),
	}

	log.Printf("[DEBUG] Creating Route53 Resolver Query Log Config Association: %s", input)
	output, err := conn.AssociateResolverQueryLogConfig(input)

	if err != nil {
		return fmt.Errorf("error creating Route53 Resolver Query Log Config Association: %w", err)
	}

	d.SetId(aws.StringValue(output.ResolverQueryLogConfigAssociation.Id))

	_, err = waitQueryLogConfigAssociationCreated(conn, d.Id())

	if err != nil {
		return fmt.Errorf("error waiting for Route53 Resolver Query Log Config Association (%s) to become available: %w", d.Id(), err)
	}

	return resourceQueryLogConfigAssociationRead(d, meta)
}

func resourceQueryLogConfigAssociationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).Route53ResolverConn

	queryLogConfigAssociation, err := FindResolverQueryLogConfigAssociationByID(conn, d.Id())

	if tfawserr.ErrMessageContains(err, route53resolver.ErrCodeResourceNotFoundException, "") {
		log.Printf("[WARN] Route53 Resolver Query Log Config Association (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Route53 Resolver Query Log Config Association (%s): %w", d.Id(), err)
	}

	if queryLogConfigAssociation == nil {
		log.Printf("[WARN] Route53 Resolver Query Log Config Association (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("resolver_query_log_config_id", queryLogConfigAssociation.ResolverQueryLogConfigId)
	d.Set("resource_id", queryLogConfigAssociation.ResourceId)

	return nil
}

func resourceQueryLogConfigAssociationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).Route53ResolverConn

	log.Printf("[DEBUG] Deleting Route53 Resolver Query Log Config Association (%s)", d.Id())
	_, err := conn.DisassociateResolverQueryLogConfig(&route53resolver.DisassociateResolverQueryLogConfigInput{
		ResolverQueryLogConfigId: aws.String(d.Get("resolver_query_log_config_id").(string)),
		ResourceId:               aws.String(d.Get("resource_id").(string)),
	})

	if tfawserr.ErrMessageContains(err, route53resolver.ErrCodeResourceNotFoundException, "") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting Route53 Resolver Query Log Config Association (%s): %w", d.Id(), err)
	}

	_, err = waitQueryLogConfigAssociationDeleted(conn, d.Id())

	if err != nil {
		return fmt.Errorf("error waiting for Route53 Resolver Query Log Config Association (%s) to be deleted: %w", d.Id(), err)
	}

	return nil
}
