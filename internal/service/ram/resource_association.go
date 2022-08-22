package ram

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ram"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourceResourceAssociation() *schema.Resource {
	return &schema.Resource{
		Create: resourceResourceAssociationCreate,
		Read:   resourceResourceAssociationRead,
		Delete: resourceResourceAssociationDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"resource_arn": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"resource_share_arn": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceResourceAssociationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RAMConn
	resourceARN := d.Get("resource_arn").(string)
	resourceShareARN := d.Get("resource_share_arn").(string)

	input := &ram.AssociateResourceShareInput{
		ClientToken:      aws.String(resource.UniqueId()),
		ResourceArns:     aws.StringSlice([]string{resourceARN}),
		ResourceShareArn: aws.String(resourceShareARN),
	}

	log.Printf("[DEBUG] Associating RAM Resource Share: %s", input)
	_, err := conn.AssociateResourceShare(input)
	if err != nil {
		return fmt.Errorf("error associating RAM Resource Share: %s", err)
	}

	d.SetId(fmt.Sprintf("%s,%s", resourceShareARN, resourceARN))

	if err := waitForResourceShareResourceAssociation(conn, resourceShareARN, resourceARN); err != nil {
		return fmt.Errorf("error waiting for RAM Resource Share (%s) Resource Association (%s): %s", resourceShareARN, resourceARN, err)
	}

	return resourceResourceAssociationRead(d, meta)
}

func resourceResourceAssociationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RAMConn

	resourceShareARN, resourceARN, err := DecodeResourceAssociationID(d.Id())
	if err != nil {
		return err
	}

	resourceShareAssociation, err := GetResourceShareAssociation(conn, resourceShareARN, resourceARN)
	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] RAM Resource Share (%s) Resource Association (%s) not found, removing from state", resourceShareARN, resourceARN)
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("error reading RAM Resource Share (%s) Resource Association (%s): %w", resourceShareARN, resourceARN, err)
	}

	if !d.IsNewResource() && aws.StringValue(resourceShareAssociation.Status) != ram.ResourceShareAssociationStatusAssociated {
		log.Printf("[WARN] RAM Resource Share (%s) Resource Association (%s) not associated, removing from state", resourceShareARN, resourceARN)
		d.SetId("")
		return nil
	}

	d.Set("resource_arn", resourceARN)
	d.Set("resource_share_arn", resourceShareARN)

	return nil
}

func resourceResourceAssociationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RAMConn

	resourceShareARN, resourceARN, err := DecodeResourceAssociationID(d.Id())
	if err != nil {
		return err
	}

	input := &ram.DisassociateResourceShareInput{
		ResourceArns:     aws.StringSlice([]string{resourceARN}),
		ResourceShareArn: aws.String(resourceShareARN),
	}

	log.Printf("[DEBUG] Disassociating RAM Resource Share: %s", input)
	_, err = conn.DisassociateResourceShare(input)

	if tfawserr.ErrCodeEquals(err, ram.ErrCodeUnknownResourceException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error disassociating RAM Resource Share (%s) Resource Association (%s): %s", resourceShareARN, resourceARN, err)
	}

	if err := WaitForResourceShareResourceDisassociation(conn, resourceShareARN, resourceARN); err != nil {
		return fmt.Errorf("error waiting for RAM Resource Share (%s) Resource Association (%s) disassociation: %s", resourceShareARN, resourceARN, err)
	}

	return nil
}

func DecodeResourceAssociationID(id string) (string, string, error) {
	idFormatErr := fmt.Errorf("unexpected format of ID (%s), expected SHARE,RESOURCE", id)

	parts := strings.SplitN(id, ",", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", idFormatErr
	}

	return parts[0], parts[1], nil
}

func GetResourceShareAssociation(conn *ram.RAM, resourceShareARN, resourceARN string) (*ram.ResourceShareAssociation, error) {
	input := &ram.GetResourceShareAssociationsInput{
		AssociationType:   aws.String(ram.ResourceShareAssociationTypeResource),
		ResourceArn:       aws.String(resourceARN),
		ResourceShareArns: aws.StringSlice([]string{resourceShareARN}),
	}

	output, err := conn.GetResourceShareAssociations(input)

	if tfawserr.ErrCodeEquals(err, ram.ErrCodeUnknownResourceException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	switch count := len(output.ResourceShareAssociations); count {
	case 0:
		return nil, tfresource.NewEmptyResultError(input)
	case 1:
		return output.ResourceShareAssociations[0], nil
	default:
		return nil, tfresource.NewTooManyResultsError(count, input)
	}
}

func resourceAssociationStateRefreshFunc(conn *ram.RAM, resourceShareARN, resourceARN string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		resourceShareAssociation, err := GetResourceShareAssociation(conn, resourceShareARN, resourceARN)
		if tfresource.NotFound(err) {
			return nil, ram.ResourceShareAssociationStatusDisassociated, nil
		}
		if err != nil {
			return nil, "", err
		}

		if aws.StringValue(resourceShareAssociation.Status) == ram.ResourceShareAssociationStatusFailed {
			extendedErr := fmt.Errorf("association status message: %s", aws.StringValue(resourceShareAssociation.StatusMessage))
			return resourceShareAssociation, aws.StringValue(resourceShareAssociation.Status), extendedErr
		}

		return resourceShareAssociation, aws.StringValue(resourceShareAssociation.Status), nil
	}
}

func waitForResourceShareResourceAssociation(conn *ram.RAM, resourceShareARN, resourceARN string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{ram.ResourceShareAssociationStatusAssociating},
		Target:  []string{ram.ResourceShareAssociationStatusAssociated},
		Refresh: resourceAssociationStateRefreshFunc(conn, resourceShareARN, resourceARN),
		Timeout: 5 * time.Minute,
	}

	_, err := stateConf.WaitForState()

	return err
}

func WaitForResourceShareResourceDisassociation(conn *ram.RAM, resourceShareARN, resourceARN string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{ram.ResourceShareAssociationStatusAssociated, ram.ResourceShareAssociationStatusDisassociating},
		Target:  []string{ram.ResourceShareAssociationStatusDisassociated},
		Refresh: resourceAssociationStateRefreshFunc(conn, resourceShareARN, resourceARN),
		Timeout: 5 * time.Minute,
	}

	_, err := stateConf.WaitForState()

	return err
}
