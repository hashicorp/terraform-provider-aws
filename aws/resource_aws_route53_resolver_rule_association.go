package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53resolver"
)

const (
	Route53ResolverRuleAssociationStatusDeleted = "DELETED"
)

func resourceAwsRoute53ResolverRuleAssociation() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsRoute53ResolverRuleAssociationCreate,
		Read:   resourceAwsRoute53ResolverRuleAssociationRead,
		Delete: resourceAwsRoute53ResolverRuleAssociationDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 64),
			},

			"resolver_rule_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 64),
			},

			"vpc_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 64),
			},
		},
	}
}

func resourceAwsRoute53ResolverRuleAssociationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).r53resolverconn

	req := &route53resolver.AssociateResolverRuleInput{
		ResolverRuleId: aws.String(d.Get("resolver_rule_id").(string)),
		VPCId:          aws.String(d.Get("vpc_id").(string)),
	}

	if v, ok := d.GetOk("name"); ok {
		req.Name = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating Route 53 Resolver rule association: %s", req)

	res, err := conn.AssociateResolverRule(req)
	if err != nil {
		return fmt.Errorf("Error creating Route 53 Resolver rule association: %s", err)
	}

	d.SetId(*res.ResolverRuleAssociation.Id)

	stateConf := &resource.StateChangeConf{
		Pending: []string{route53resolver.ResolverRuleAssociationStatusCreating},
		Target:  []string{route53resolver.ResolverRuleAssociationStatusComplete},
		Refresh: resourceAwsRoute53ResolverRuleAssociationStateRefresh(conn, d.Id()),
		Timeout: d.Timeout(schema.TimeoutDelete),
	}
	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("Error waiting for Route 53 Resolver rule association (%s) to be created: %s", d.Id(), err)
	}

	return resourceAwsRoute53ResolverRuleAssociationRead(d, meta)
}

func resourceAwsRoute53ResolverRuleAssociationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).r53resolverconn

	req := &route53resolver.GetResolverRuleAssociationInput{
		ResolverRuleAssociationId: aws.String(d.Id()),
	}

	res, err := conn.GetResolverRuleAssociation(req)
	if err != nil {
		if isAWSErr(err, route53resolver.ErrCodeResourceNotFoundException, "") {
			log.Printf("[WARN] No Route 53 Resolver rule association by Id (%s) found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error reading Route 53 Resolver rule association %s: %s", d.Id(), err)
	}

	assn := res.ResolverRuleAssociation

	d.Set("resolver_rule_id", assn.ResolverRuleId)
	d.Set("vpc_id", assn.VPCId)

	return nil
}

func resourceAwsRoute53ResolverRuleAssociationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).r53resolverconn

	req := &route53resolver.DisassociateResolverRuleInput{
		ResolverRuleId: aws.String(d.Get("resolver_rule_id").(string)),
		VPCId:          aws.String(d.Get("vpc_id").(string)),
	}

	_, err := conn.DisassociateResolverRule(req)
	if err != nil {
		if isAWSErr(err, route53resolver.ErrCodeResourceNotFoundException, "") {
			return nil
		}
		return fmt.Errorf("Error deleting Route 53 Resolver rule association %s: %s", d.Id(), err)
	}

	stateConf := &resource.StateChangeConf{
		Pending: []string{route53resolver.ResolverRuleAssociationStatusDeleting},
		Target:  []string{Route53ResolverRuleAssociationStatusDeleted},
		Refresh: resourceAwsRoute53ResolverRuleAssociationStateRefresh(conn, d.Id()),
		Timeout: d.Timeout(schema.TimeoutDelete),
	}
	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("Error waiting for Route 53 Resolver rule association (%s) to be deleted: %s", d.Id(), err)
	}

	return nil
}

func resourceAwsRoute53ResolverRuleAssociationStateRefresh(conn *route53resolver.Route53Resolver, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {

		req := &route53resolver.GetResolverRuleAssociationInput{
			ResolverRuleAssociationId: aws.String(id),
		}

		res, err := conn.GetResolverRuleAssociation(req)
		if err != nil {
			if isAWSErr(err, route53resolver.ErrCodeResourceNotFoundException, "") {
				return "", Route53ResolverRuleAssociationStatusDeleted, nil
			}
			return nil, "", err
		}

		return res.ResolverRuleAssociation, aws.StringValue(res.ResolverRuleAssociation.Status), nil
	}
}
