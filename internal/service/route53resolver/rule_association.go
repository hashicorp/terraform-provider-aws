package route53resolver

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53resolver"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

const (
	RuleAssociationStatusDeleted = "DELETED"
)

func ResourceRuleAssociation() *schema.Resource {
	return &schema.Resource{
		Create: resourceRuleAssociationCreate,
		Read:   resourceRuleAssociationRead,
		Delete: resourceRuleAssociationDelete,

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
				ValidateFunc: validResolverName,
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

func resourceRuleAssociationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).Route53ResolverConn

	req := &route53resolver.AssociateResolverRuleInput{
		ResolverRuleId: aws.String(d.Get("resolver_rule_id").(string)),
		VPCId:          aws.String(d.Get("vpc_id").(string)),
	}
	if v, ok := d.GetOk("name"); ok {
		req.Name = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating Route53 Resolver rule association: %s", req)
	resp, err := conn.AssociateResolverRule(req)
	if err != nil {
		return fmt.Errorf("error creating Route53 Resolver rule association: %s", err)
	}

	d.SetId(aws.StringValue(resp.ResolverRuleAssociation.Id))

	err = RuleAssociationWaitUntilTargetState(conn, d.Id(), d.Timeout(schema.TimeoutCreate),
		[]string{route53resolver.ResolverRuleAssociationStatusCreating},
		[]string{route53resolver.ResolverRuleAssociationStatusComplete})
	if err != nil {
		return err
	}

	return resourceRuleAssociationRead(d, meta)
}

func resourceRuleAssociationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).Route53ResolverConn

	ruleAssociation, err := FindResolverRuleAssociationByID(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Route53 Resolver Rule Association (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("reading Route53 Resolver Rule Association (%s): %w", d.Id(), err)
	}

	d.Set("name", ruleAssociation.Name)
	d.Set("resolver_rule_id", ruleAssociation.ResolverRuleId)
	d.Set("vpc_id", ruleAssociation.VPCId)

	return nil
}

func resourceRuleAssociationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).Route53ResolverConn

	log.Printf("[DEBUG] Deleting Route53 Resolver Rule Association: %s", d.Id())
	_, err := conn.DisassociateResolverRule(&route53resolver.DisassociateResolverRuleInput{
		ResolverRuleId: aws.String(d.Get("resolver_rule_id").(string)),
		VPCId:          aws.String(d.Get("vpc_id").(string)),
	})
	if tfawserr.ErrCodeEquals(err, route53resolver.ErrCodeResourceNotFoundException) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("error deleting Route53 Resolver rule association (%s): %s", d.Id(), err)
	}

	return RuleAssociationWaitUntilTargetState(conn, d.Id(), d.Timeout(schema.TimeoutDelete),
		[]string{route53resolver.ResolverRuleAssociationStatusDeleting},
		[]string{RuleAssociationStatusDeleted})
}

func FindResolverRuleAssociationByID(conn *route53resolver.Route53Resolver, id string) (*route53resolver.ResolverRuleAssociation, error) {
	input := &route53resolver.GetResolverRuleAssociationInput{
		ResolverRuleAssociationId: aws.String(id),
	}

	output, err := conn.GetResolverRuleAssociation(input)

	if tfawserr.ErrCodeEquals(err, route53resolver.ErrCodeResourceNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.ResolverRuleAssociation == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.ResolverRuleAssociation, nil
}

func ruleAssociationRefresh(conn *route53resolver.Route53Resolver, assocId string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		resp, err := conn.GetResolverRuleAssociation(&route53resolver.GetResolverRuleAssociationInput{
			ResolverRuleAssociationId: aws.String(assocId),
		})
		if tfawserr.ErrCodeEquals(err, route53resolver.ErrCodeResourceNotFoundException) {
			return "", RuleAssociationStatusDeleted, nil
		}
		if err != nil {
			return nil, "", err
		}

		if statusMessage := aws.StringValue(resp.ResolverRuleAssociation.StatusMessage); statusMessage != "" {
			log.Printf("[INFO] Route53 Resolver rule association (%s) status message: %s", assocId, statusMessage)
		}

		return resp.ResolverRuleAssociation, aws.StringValue(resp.ResolverRuleAssociation.Status), nil
	}
}

func RuleAssociationWaitUntilTargetState(conn *route53resolver.Route53Resolver, assocId string, timeout time.Duration, pending, target []string) error {
	stateConf := &resource.StateChangeConf{
		Pending:    pending,
		Target:     target,
		Refresh:    ruleAssociationRefresh(conn, assocId),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 5 * time.Second,
	}
	if _, err := stateConf.WaitForState(); err != nil {
		return fmt.Errorf("error waiting for Route53 Resolver rule association (%s) to reach target state: %s", assocId, err)
	}

	return nil
}
