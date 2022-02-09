package route53domains

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53domains"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourceRegisteredDomain() *schema.Resource {
	return &schema.Resource{
		Create: resourceRegisteredDomainCreate,
		Read:   resourceRegisteredDomainRead,
		Update: resourceRegisteredDomainUpdate,
		Delete: resourceRegisteredDomainDelete,

		Schema: map[string]*schema.Schema{
			"domain_name": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceRegisteredDomainCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).Route53DomainsConn

	domainName := d.Get("domain_name").(string)
	domainDetail, err := FindDomainDetailByName(conn, domainName)

	if err != nil {
		return fmt.Errorf("error reading Route 53 Domains Domain (%s): %w", domainName, err)
	}

	d.SetId(aws.StringValue(domainDetail.DomainName))

	return resourceRegisteredDomainRead(d, meta)
}

func resourceRegisteredDomainRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).Route53DomainsConn

	domainDetail, err := FindDomainDetailByName(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Route 53 Domains Domain %s not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Route 53 Domains Domain (%s): %w", d.Id(), err)
	}

	d.Set("domain_name", domainDetail.DomainName)

	return nil
}

func resourceRegisteredDomainUpdate(d *schema.ResourceData, meta interface{}) error {
	//conn := meta.(*conns.AWSClient).Route53DomainsConn

	return resourceRegisteredDomainRead(d, meta)
}

func resourceRegisteredDomainDelete(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[WARN] Route 53 Domains Registered Domain (%s) not deleted, removing from state", d.Id())

	return nil
}

func FindDomainDetailByName(conn *route53domains.Route53Domains, name string) (*route53domains.GetDomainDetailOutput, error) {
	input := &route53domains.GetDomainDetailInput{
		DomainName: aws.String(name),
	}

	output, err := conn.GetDomainDetail(input)

	if tfawserr.ErrMessageContains(err, route53domains.ErrCodeInvalidInput, "not found") {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}
