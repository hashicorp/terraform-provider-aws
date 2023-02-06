package route53resolver

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53resolver"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourceRuleAssociation() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceRuleAssociationCreate,
		ReadWithoutTimeout:   resourceRuleAssociationRead,
		DeleteWithoutTimeout: resourceRuleAssociationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
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

func resourceRuleAssociationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).Route53ResolverConn()

	input := &route53resolver.AssociateResolverRuleInput{
		ResolverRuleId: aws.String(d.Get("resolver_rule_id").(string)),
		VPCId:          aws.String(d.Get("vpc_id").(string)),
	}

	if v, ok := d.GetOk("name"); ok {
		input.Name = aws.String(v.(string))
	}

	output, err := conn.AssociateResolverRuleWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("creating Route53 Resolver Rule Association: %s", err)
	}

	d.SetId(aws.StringValue(output.ResolverRuleAssociation.Id))

	if _, err := waitRuleAssociationCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return diag.Errorf("waiting for Route53 Resolver Rule Association (%s) create: %s", d.Id(), err)
	}

	return resourceRuleAssociationRead(ctx, d, meta)
}

func resourceRuleAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).Route53ResolverConn()

	ruleAssociation, err := FindResolverRuleAssociationByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Route53 Resolver Rule Association (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading Route53 Resolver Rule Association (%s): %s", d.Id(), err)
	}

	d.Set("name", ruleAssociation.Name)
	d.Set("resolver_rule_id", ruleAssociation.ResolverRuleId)
	d.Set("vpc_id", ruleAssociation.VPCId)

	return nil
}

func resourceRuleAssociationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).Route53ResolverConn()

	log.Printf("[DEBUG] Deleting Route53 Resolver Rule Association: %s", d.Id())
	_, err := conn.DisassociateResolverRuleWithContext(ctx, &route53resolver.DisassociateResolverRuleInput{
		ResolverRuleId: aws.String(d.Get("resolver_rule_id").(string)),
		VPCId:          aws.String(d.Get("vpc_id").(string)),
	})

	if tfawserr.ErrCodeEquals(err, route53resolver.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return diag.Errorf("deleting Route53 Resolver Rule Association (%s): %s", d.Id(), err)
	}

	if _, err := waitRuleAssociationDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return diag.Errorf("waiting for Route53 Resolver Rule Association (%s) delete: %s", d.Id(), err)
	}

	return nil
}

func FindResolverRuleAssociationByID(ctx context.Context, conn *route53resolver.Route53Resolver, id string) (*route53resolver.ResolverRuleAssociation, error) {
	input := &route53resolver.GetResolverRuleAssociationInput{
		ResolverRuleAssociationId: aws.String(id),
	}

	output, err := conn.GetResolverRuleAssociationWithContext(ctx, input)

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

func statusRuleAssociation(ctx context.Context, conn *route53resolver.Route53Resolver, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindResolverRuleAssociationByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status), nil
	}
}

func waitRuleAssociationCreated(ctx context.Context, conn *route53resolver.Route53Resolver, id string, timeout time.Duration) (*route53resolver.ResolverRuleAssociation, error) {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{route53resolver.ResolverRuleAssociationStatusCreating},
		Target:     []string{route53resolver.ResolverRuleAssociationStatusComplete},
		Refresh:    statusRuleAssociation(ctx, conn, id),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 5 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*route53resolver.ResolverRuleAssociation); ok {
		tfresource.SetLastError(err, errors.New(aws.StringValue(output.StatusMessage)))

		return output, err
	}

	return nil, err
}

func waitRuleAssociationDeleted(ctx context.Context, conn *route53resolver.Route53Resolver, id string, timeout time.Duration) (*route53resolver.ResolverRuleAssociation, error) {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{route53resolver.ResolverRuleAssociationStatusDeleting},
		Target:     []string{},
		Refresh:    statusRuleAssociation(ctx, conn, id),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 5 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*route53resolver.ResolverRuleAssociation); ok {
		tfresource.SetLastError(err, errors.New(aws.StringValue(output.StatusMessage)))

		return output, err
	}

	return nil, err
}
