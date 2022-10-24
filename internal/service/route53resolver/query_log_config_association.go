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
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
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

	output, err := conn.AssociateResolverQueryLogConfig(input)

	if err != nil {
		return fmt.Errorf("creating Route53 Resolver Query Log Config Association: %w", err)
	}

	d.SetId(aws.StringValue(output.ResolverQueryLogConfigAssociation.Id))

	if _, err := waitQueryLogConfigAssociationCreated(conn, d.Id()); err != nil {
		return fmt.Errorf("waiting for Route53 Resolver Query Log Config Association (%s) create: %w", d.Id(), err)
	}

	return resourceQueryLogConfigAssociationRead(d, meta)
}

func resourceQueryLogConfigAssociationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).Route53ResolverConn

	queryLogConfigAssociation, err := FindResolverQueryLogConfigAssociationByID(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Route53 Resolver Query Log Config Association (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("reading Route53 Resolver Query Log Config Association (%s): %w", d.Id(), err)
	}

	d.Set("resolver_query_log_config_id", queryLogConfigAssociation.ResolverQueryLogConfigId)
	d.Set("resource_id", queryLogConfigAssociation.ResourceId)

	return nil
}

func resourceQueryLogConfigAssociationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).Route53ResolverConn

	log.Printf("[DEBUG] Deleting Route53 Resolver Query Log Config Association: %s", d.Id())
	_, err := conn.DisassociateResolverQueryLogConfig(&route53resolver.DisassociateResolverQueryLogConfigInput{
		ResolverQueryLogConfigId: aws.String(d.Get("resolver_query_log_config_id").(string)),
		ResourceId:               aws.String(d.Get("resource_id").(string)),
	})

	if tfawserr.ErrCodeEquals(err, route53resolver.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("deleting Route53 Resolver Query Log Config Association (%s): %w", d.Id(), err)
	}

	if _, err := waitQueryLogConfigAssociationDeleted(conn, d.Id()); err != nil {
		return fmt.Errorf("waiting for Route53 Resolver Query Log Config Association (%s) delete: %w", d.Id(), err)
	}

	return nil
}

func FindResolverQueryLogConfigAssociationByID(conn *route53resolver.Route53Resolver, id string) (*route53resolver.ResolverQueryLogConfigAssociation, error) {
	input := &route53resolver.GetResolverQueryLogConfigAssociationInput{
		ResolverQueryLogConfigAssociationId: aws.String(id),
	}

	output, err := conn.GetResolverQueryLogConfigAssociation(input)

	if tfawserr.ErrCodeEquals(err, route53resolver.ErrCodeResourceNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.ResolverQueryLogConfigAssociation == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.ResolverQueryLogConfigAssociation, nil
}

func statusQueryLogConfigAssociation(conn *route53resolver.Route53Resolver, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindResolverQueryLogConfigAssociationByID(conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status), nil
	}
}

const (
	queryLogConfigAssociationCreatedTimeout = 5 * time.Minute
	queryLogConfigAssociationDeletedTimeout = 5 * time.Minute
)

func waitQueryLogConfigAssociationCreated(conn *route53resolver.Route53Resolver, id string) (*route53resolver.ResolverQueryLogConfigAssociation, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{route53resolver.ResolverQueryLogConfigAssociationStatusCreating},
		Target:  []string{route53resolver.ResolverQueryLogConfigAssociationStatusActive},
		Refresh: statusQueryLogConfigAssociation(conn, id),
		Timeout: queryLogConfigAssociationCreatedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*route53resolver.ResolverQueryLogConfigAssociation); ok {
		if status := aws.StringValue(output.Status); status == route53resolver.ResolverQueryLogConfigAssociationStatusFailed {
			tfresource.SetLastError(err, fmt.Errorf("%s: %s", aws.StringValue(output.Error), aws.StringValue(output.ErrorMessage)))
		}

		return output, err
	}

	return nil, err
}

func waitQueryLogConfigAssociationDeleted(conn *route53resolver.Route53Resolver, id string) (*route53resolver.ResolverQueryLogConfigAssociation, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{route53resolver.ResolverQueryLogConfigAssociationStatusDeleting},
		Target:  []string{},
		Refresh: statusQueryLogConfigAssociation(conn, id),
		Timeout: queryLogConfigAssociationDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*route53resolver.ResolverQueryLogConfigAssociation); ok {
		if status := aws.StringValue(output.Status); status == route53resolver.ResolverQueryLogConfigAssociationStatusFailed {
			tfresource.SetLastError(err, fmt.Errorf("%s: %s", aws.StringValue(output.Error), aws.StringValue(output.ErrorMessage)))
		}

		return output, err
	}

	return nil, err
}
