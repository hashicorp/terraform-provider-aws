package aws

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ram"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsRamPrincipalAssociation() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsRamPrincipalAssociationCreate,
		Read:   resourceAwsRamPrincipalAssociationRead,
		Delete: resourceAwsRamPrincipalAssociationDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"resource_share_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateArn,
			},

			"principal": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceAwsRamPrincipalAssociationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ramconn

	resourceShareArn := d.Get("resource_share_arn").(string)
	principal := d.Get("principal").(string)

	request := &ram.AssociateResourceShareInput{
		ResourceShareArn: aws.String(resourceShareArn),
		Principals:       []*string{aws.String(principal)},
	}

	log.Println("[DEBUG] Create RAM principal association request:", request)
	_, err := conn.AssociateResourceShare(request)
	if err != nil {
		return fmt.Errorf("Error associating principal with RAM resource share: %s", err)
	}

	d.SetId(fmt.Sprintf("%s,%s", resourceShareArn, principal))

	stateConf := &resource.StateChangeConf{
		Pending: []string{ram.ResourceShareAssociationStatusAssociating},
		Target:  []string{ram.ResourceShareAssociationStatusAssociated},
		Refresh: resourceAwsRamPrincipalAssociationStateRefreshFunc(conn, resourceShareArn, principal),
		Timeout: d.Timeout(schema.TimeoutCreate),
	}

	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("Error waiting for RAM principal association (%s) to become ready: %s", d.Id(), err)
	}

	return resourceAwsRamPrincipalAssociationRead(d, meta)
}

func resourceAwsRamPrincipalAssociationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ramconn

	resourceShareArn, principal, err := resourceAwsRamPrincipalAssociationParseId(d.Id())
	if err != nil {
		return err
	}

	request := &ram.GetResourceShareAssociationsInput{
		ResourceShareArns: []*string{aws.String(resourceShareArn)},
		AssociationType:   aws.String(ram.ResourceShareAssociationTypePrincipal),
		Principal:         aws.String(principal),
	}

	output, err := conn.GetResourceShareAssociations(request)

	if err != nil {
		return fmt.Errorf("Error reading RAM principal association %s: %s", d.Id(), err)
	}

	if len(output.ResourceShareAssociations) == 0 {
		log.Printf("[WARN] RAM principal (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	association := output.ResourceShareAssociations[0]

	if aws.StringValue(association.Status) == ram.ResourceShareAssociationStatusAssociated {
		log.Printf("[WARN] RAM principal (%s) disassociat(ing|ed), removing from state", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("resource_share_arn", resourceShareArn)
	d.Set("principal", principal)

	return nil
}

func resourceAwsRamPrincipalAssociationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ramconn

	resourceShareArn, principal, err := resourceAwsRamPrincipalAssociationParseId(d.Id())
	if err != nil {
		return err
	}

	request := &ram.DisassociateResourceShareInput{
		ResourceShareArn: aws.String(resourceShareArn),
		Principals:       []*string{aws.String(principal)},
	}

	log.Println("[DEBUG] Delete RAM principal association request:", request)
	_, err = conn.DisassociateResourceShare(request)
	if err != nil {
		return fmt.Errorf("Error disassociating principals from RAM Resource Share: %s", err)
	}

	stateConf := &resource.StateChangeConf{
		Pending: []string{ram.ResourceShareAssociationStatusDisassociating},
		Target:  []string{ram.ResourceShareAssociationStatusDisassociated},
		Refresh: resourceAwsRamPrincipalAssociationStateRefreshFunc(conn, resourceShareArn, principal),
		Timeout: d.Timeout(schema.TimeoutDelete),
	}

	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("Error waiting for RAM principal association (%s) to become ready: %s", d.Id(), err)
	}

	return nil
}

func resourceAwsRamPrincipalAssociationParseId(id string) (string, string, error) {
	parts := strings.SplitN(id, ",", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("Expected RAM principal association ID in the form resource_share_arn,principal - received: %s", id)
	}
	return parts[0], parts[1], nil
}

func resourceAwsRamPrincipalAssociationStateRefreshFunc(conn *ram.RAM, resourceShareArn, principal string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		request := &ram.GetResourceShareAssociationsInput{
			ResourceShareArns: []*string{aws.String(resourceShareArn)},
			AssociationType:   aws.String(ram.ResourceShareAssociationTypePrincipal),
			Principal:         aws.String(principal),
		}

		output, err := conn.GetResourceShareAssociations(request)

		if err != nil {
			return nil, ram.ResourceShareAssociationStatusFailed, err
		}

		if len(output.ResourceShareAssociations) == 0 {
			return nil, ram.ResourceShareAssociationStatusDisassociated, nil
		}

		association := output.ResourceShareAssociations[0]

		return association, aws.StringValue(association.Status), nil
	}
}
