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

func ResourceApiAssociation() *schema.Resource {

	return &schema.Resource{
		Create: resourceApiAssociationCreate,
		Read:   resourceApiAssociationRead,
		Update: resourceApiAssociationUpdate,
		Delete: resourceApiAssociationDelete,
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

func resourceApiAssociationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).AppSyncConn

	params := &appsync.AssociateApiInput{
		ApiId:      aws.String(d.Get("description").(string)),
		DomainName: aws.String(d.Get("domain_name").(string)),
	}

	resp, err := conn.AssociateApi(params)
	if err != nil {
		return fmt.Errorf("error creating Appsync API Association: %w", err)
	}

	d.SetId(aws.StringValue(resp.ApiAssociation.DomainName))

	return resourceApiAssociationRead(d, meta)
}

func resourceApiAssociationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).AppSyncConn

	association, err := FindApiAssociationByID(conn, d.Id())
	if association == nil && !d.IsNewResource() {
		log.Printf("[WARN] AppSync API Association (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error getting Appsync API Association %q: %w", d.Id(), err)
	}

	d.Set("domain_name", association.DomainName)
	d.Set("api_id", association.ApiId)

	return nil
}

func resourceApiAssociationUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).AppSyncConn

	params := &appsync.AssociateApiInput{
		ApiId:      aws.String(d.Get("description").(string)),
		DomainName: aws.String(d.Get("domain_name").(string)),
	}

	_, err := conn.AssociateApi(params)
	if err != nil {
		return fmt.Errorf("error creating Appsync API Association: %w", err)
	}

	return resourceApiAssociationRead(d, meta)
}

func resourceApiAssociationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).AppSyncConn

	input := &appsync.DisassociateApiInput{
		DomainName: aws.String(d.Id()),
	}
	_, err := conn.DisassociateApi(input)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, appsync.ErrCodeNotFoundException) {
			return nil
		}
		return fmt.Errorf("error deleting Appsync API Association: %w", err)
	}

	return nil
}
