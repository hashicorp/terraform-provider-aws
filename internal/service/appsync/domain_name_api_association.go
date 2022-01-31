package appsync

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appsync"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func ResourceDomainNameApiAssociation() *schema.Resource {

	return &schema.Resource{
		Create: resourceDomainNameApiAssociationCreate,
		Read:   resourceDomainNameApiAssociationRead,
		Update: resourceDomainNameApiAssociationUpdate,
		Delete: resourceDomainNameApiAssociationDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"api_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"domain_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceDomainNameApiAssociationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).AppSyncConn

	params := &appsync.AssociateApiInput{
		ApiId:      aws.String(d.Get("api_id").(string)),
		DomainName: aws.String(d.Get("domain_name").(string)),
	}

	resp, err := conn.AssociateApi(params)
	if err != nil {
		return fmt.Errorf("error creating Appsync Domain Name API Association: %w", err)
	}

	d.SetId(aws.StringValue(resp.ApiAssociation.DomainName))

	if err := waitDomainNameApiAssociation(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for Appsync Domain Name API (%s) Association: %w", d.Id(), err)
	}

	return resourceDomainNameApiAssociationRead(d, meta)
}

func resourceDomainNameApiAssociationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).AppSyncConn

	association, err := FindDomainNameApiAssociationByID(conn, d.Id())
	if association == nil && !d.IsNewResource() {
		log.Printf("[WARN] Appsync Domain Name API Association (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error getting Appsync Domain Name API Association %q: %w", d.Id(), err)
	}

	d.Set("domain_name", association.DomainName)
	d.Set("api_id", association.ApiId)

	return nil
}

func resourceDomainNameApiAssociationUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).AppSyncConn

	params := &appsync.AssociateApiInput{
		ApiId:      aws.String(d.Get("api_id").(string)),
		DomainName: aws.String(d.Get("domain_name").(string)),
	}

	_, err := conn.AssociateApi(params)
	if err != nil {
		return fmt.Errorf("error creating Appsync Domain Name API Association: %w", err)
	}

	if err := waitDomainNameApiAssociation(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for Appsync Domain Name API (%s) Association: %w", d.Id(), err)
	}

	return resourceDomainNameApiAssociationRead(d, meta)
}

func resourceDomainNameApiAssociationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).AppSyncConn

	input := &appsync.DisassociateApiInput{
		DomainName: aws.String(d.Id()),
	}
	_, err := conn.DisassociateApi(input)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, appsync.ErrCodeNotFoundException) {
			return nil
		}
		return fmt.Errorf("error deleting Appsync Domain Name API Association: %w", err)
	}

	if err := waitDomainNameApiDisassociation(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for Appsync Domain Name API (%s) Disassociation: %w", d.Id(), err)
	}

	return nil
}
