package ram

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ram"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourcePrincipalAssociation() *schema.Resource {
	return &schema.Resource{
		Create: resourcePrincipalAssociationCreate,
		Read:   resourcePrincipalAssociationRead,
		Delete: resourcePrincipalAssociationDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"resource_share_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},

			"principal": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.Any(
					verify.ValidAccountID,
					verify.ValidARN,
				),
			},
		},
	}
}

func resourcePrincipalAssociationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RAMConn

	resourceShareArn := d.Get("resource_share_arn").(string)
	principal := d.Get("principal").(string)

	request := &ram.AssociateResourceShareInput{
		ClientToken:      aws.String(resource.UniqueId()),
		ResourceShareArn: aws.String(resourceShareArn),
		Principals:       []*string{aws.String(principal)},
	}

	log.Println("[DEBUG] Create RAM principal association request:", request)
	_, err := conn.AssociateResourceShare(request)
	if err != nil {
		return fmt.Errorf("error associating principal with RAM resource share: %w", err)
	}

	d.SetId(fmt.Sprintf("%s,%s", resourceShareArn, principal))

	// AWS Account ID Principals need to be accepted to become ASSOCIATED
	if ok, _ := regexp.MatchString(`^\d{12}$`, principal); ok {
		return resourcePrincipalAssociationRead(d, meta)
	}

	if _, err := WaitResourceSharePrincipalAssociated(conn, resourceShareArn, principal); err != nil {
		return fmt.Errorf("error waiting for RAM principal association (%s) to become ready: %w", d.Id(), err)
	}

	return resourcePrincipalAssociationRead(d, meta)
}

func resourcePrincipalAssociationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RAMConn

	resourceShareArn, principal, err := PrincipalAssociationParseID(d.Id())
	if err != nil {
		return fmt.Errorf("error reading RAM Principal Association, parsing ID (%s): %w", d.Id(), err)
	}

	var association *ram.ResourceShareAssociation

	if ok, _ := regexp.MatchString(`^\d{12}$`, principal); ok {
		// AWS Account ID Principals need to be accepted to become ASSOCIATED
		association, err = FindResourceSharePrincipalAssociationByShareARNPrincipal(conn, resourceShareArn, principal)
	} else {
		association, err = WaitResourceSharePrincipalAssociated(conn, resourceShareArn, principal)
	}

	if !d.IsNewResource() && (tfawserr.ErrCodeEquals(err, ram.ErrCodeResourceArnNotFoundException) || tfawserr.ErrCodeEquals(err, ram.ErrCodeUnknownResourceException)) {
		log.Printf("[WARN] No RAM resource share principal association with ARN (%s) found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading RAM Resource Share (%s) Principal Association (%s): %s", resourceShareArn, principal, err)
	}

	if !d.IsNewResource() && (association == nil || aws.StringValue(association.Status) == ram.ResourceShareAssociationStatusDisassociated) {
		log.Printf("[WARN] RAM resource share principal association with ARN (%s) found, but empty or disassociated - removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if aws.StringValue(association.Status) != ram.ResourceShareAssociationStatusAssociated && aws.StringValue(association.Status) != ram.ResourceShareAssociationStatusAssociating {
		return fmt.Errorf("error reading RAM Resource Share (%s) Principal Association (%s), status not associating or associated: %s", resourceShareArn, principal, aws.StringValue(association.Status))
	}

	d.Set("resource_share_arn", resourceShareArn)
	d.Set("principal", principal)

	return nil
}

func resourcePrincipalAssociationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RAMConn

	resourceShareArn, principal, err := PrincipalAssociationParseID(d.Id())
	if err != nil {
		return err
	}

	request := &ram.DisassociateResourceShareInput{
		ResourceShareArn: aws.String(resourceShareArn),
		Principals:       []*string{aws.String(principal)},
	}

	log.Println("[DEBUG] Delete RAM principal association request:", request)
	_, err = conn.DisassociateResourceShare(request)

	if tfawserr.ErrCodeEquals(err, ram.ErrCodeUnknownResourceException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error disassociating RAM Resource Share (%s) Principal Association (%s): %s", resourceShareArn, principal, err)
	}

	if _, err := WaitResourceSharePrincipalDisassociated(conn, resourceShareArn, principal); err != nil {
		return fmt.Errorf("error waiting for RAM Resource Share (%s) Principal Association (%s) disassociation: %s", resourceShareArn, principal, err)
	}

	return nil
}

func PrincipalAssociationParseID(id string) (string, string, error) {
	idFormatErr := fmt.Errorf("unexpected format of ID (%s), expected SHARE,PRINCIPAL", id)

	parts := strings.SplitN(id, ",", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", idFormatErr
	}

	return parts[0], parts[1], nil
}
