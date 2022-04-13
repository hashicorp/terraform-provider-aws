package ec2

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func ResourceAMILaunchPermission() *schema.Resource {
	return &schema.Resource{
		Create: resourceAMILaunchPermissionCreate,
		Read:   resourceAMILaunchPermissionRead,
		Delete: resourceAMILaunchPermissionDelete,

		Importer: &schema.ResourceImporter{
			State: func(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				idParts := strings.Split(d.Id(), "/")
				if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
					return nil, fmt.Errorf("Unexpected format of ID (%q), expected ACCOUNT-ID/IMAGE-ID", d.Id())
				}
				accountId := idParts[0]
				imageId := idParts[1]
				d.Set("account_id", accountId)
				d.Set("image_id", imageId)
				d.SetId(fmt.Sprintf("%s-%s", imageId, accountId))
				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"account_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"image_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceAMILaunchPermissionCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	imageID := d.Get("image_id").(string)
	accountID := d.Get("account_id").(string)
	id := AMILaunchPermissionCreateResourceID(imageID, accountID)
	input := &ec2.ModifyImageAttributeInput{
		Attribute: aws.String(ec2.ImageAttributeNameLaunchPermission),
		ImageId:   aws.String(imageID),
		LaunchPermission: &ec2.LaunchPermissionModifications{
			Add: expandLaunchPermissions(accountID),
		},
	}

	log.Printf("[DEBUG] Creating AMI Launch Permission: %s", input)
	_, err := conn.ModifyImageAttribute(input)

	if err != nil {
		return fmt.Errorf("creating AMI Launch Permission (%s): %w", id, err)
	}

	d.SetId(id)

	return nil
}

func resourceAMILaunchPermissionRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	exists, err := HasLaunchPermission(conn, d.Get("image_id").(string), d.Get("account_id").(string))
	if err != nil {
		return fmt.Errorf("error reading AMI launch permission (%s): %w", d.Id(), err)
	}
	if !exists {
		if d.IsNewResource() {
			return fmt.Errorf("error reading EC2 AMI Launch Permission (%s): not found", d.Id())
		}

		log.Printf("[WARN] AMI launch permission (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	return nil
}

func resourceAMILaunchPermissionDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	imageID, accountID, err := AMILaunchPermissionParseResourceID(d.Id())

	if err != nil {
		return err
	}

	input := &ec2.ModifyImageAttributeInput{
		Attribute: aws.String(ec2.ImageAttributeNameLaunchPermission),
		ImageId:   aws.String(imageID),
		LaunchPermission: &ec2.LaunchPermissionModifications{
			Remove: expandLaunchPermissions(accountID),
		},
	}

	log.Printf("[INFO] Deleting AMI Launch Permission: %s", d.Id())
	_, err = conn.ModifyImageAttribute(input)

	if tfawserr.ErrCodeEquals(err, ErrCodeInvalidAMIIDNotFound, ErrCodeInvalidAMIIDUnavailable) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("deleting AMI Launch Permission (%s): %w", d.Id(), err)
	}

	return nil
}

func HasLaunchPermission(conn *ec2.EC2, image_id string, account_id string) (bool, error) {
	attrs, err := conn.DescribeImageAttribute(&ec2.DescribeImageAttributeInput{
		ImageId:   aws.String(image_id),
		Attribute: aws.String(ec2.ImageAttributeNameLaunchPermission),
	})
	if err != nil {
		// When an AMI disappears out from under a launch permission resource, we will
		// see either InvalidAMIID.NotFound or InvalidAMIID.Unavailable.
		if ec2err, ok := err.(awserr.Error); ok && strings.HasPrefix(ec2err.Code(), "InvalidAMIID") {
			log.Printf("[DEBUG] %s no longer exists, so we'll drop launch permission for %s from the state", image_id, account_id)
			return false, nil
		}
		return false, err
	}

	for _, lp := range attrs.LaunchPermissions {
		if aws.StringValue(lp.UserId) == account_id {
			return true, nil
		}
	}
	return false, nil
}

func expandLaunchPermissions(accountID string) []*ec2.LaunchPermission {
	apiObject := &ec2.LaunchPermission{}

	if accountID != "" {
		apiObject.UserId = aws.String(accountID)
	}

	return []*ec2.LaunchPermission{apiObject}
}

const amiLaunchPermissionIDSeparator = "-"

func AMILaunchPermissionCreateResourceID(imageID, accountID string) string {
	parts := []string{imageID, accountID}
	id := strings.Join(parts, amiLaunchPermissionIDSeparator)

	return id
}

func AMILaunchPermissionParseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, amiLaunchPermissionIDSeparator)

	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		return parts[0], parts[1], nil
	}

	return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected IMAGE-ID%[2]sACCOUNT-ID", id, amiLaunchPermissionIDSeparator)
}
