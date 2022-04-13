package ec2

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
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
					return nil, fmt.Errorf("Unexpected format of ID (%q), expected ACCOUNT-ID/IMAGE-ID, or 'ARN'/IMAGE-ID", d.Id())
				}
				imageId := idParts[1]
				d.Set("image_id", imageId)
				if strings.HasPrefix(idParts[0], "arn") {
					arn := idParts[0]
					d.Set("arn", arn)
					d.SetId(fmt.Sprintf("%s-%s", imageId, arn))
				} else {
					accountId := idParts[0]
					d.Set("account_id", accountId)
					d.SetId(fmt.Sprintf("%s-%s", imageId, accountId))
				}
				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"image_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"account_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"account_id", "arn"},
			},
			"arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"account_id", "arn"},
			},
			"arn_type": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				RequiredWith: []string{"arn"},
				ValidateFunc: validation.StringInSlice([]string{
					"OrganizationArn",
					"OrganizationalUnitArn",
				}, false),
			},
		},
	}
}

func resourceAMILaunchPermissionCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	image_id := d.Get("image_id").(string)

	launch_permission := BuildLaunchPermission(d)

	_, err := conn.ModifyImageAttribute(&ec2.ModifyImageAttributeInput{
		ImageId:   aws.String(image_id),
		Attribute: aws.String(ec2.ImageAttributeNameLaunchPermission),
		LaunchPermission: &ec2.LaunchPermissionModifications{
			Add: launch_permission,
		},
	})
	if err != nil {
		return fmt.Errorf("error creating AMI launch permission: %w", err)
	}

	account_id := d.Get("account_id").(string)
	d.SetId(fmt.Sprintf("%s-%s", image_id, account_id))
	return nil
}

func resourceAMILaunchPermissionRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	account_id := d.Get("account_id").(string)
	arn := d.Get("arn").(string)
	arn_type := d.Get("arn_type").(string)

	var read_value, read_type string

	if len(account_id) > 0 {
		read_value = account_id
		read_type = "UserId"
	} else {
		read_value = arn
		read_type = arn_type
	}

	exists, err := HasLaunchPermission(conn, d.Get("image_id").(string), read_value, read_type)
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

	image_id := d.Get("image_id").(string)

	launch_permission := BuildLaunchPermission(d)

	_, err := conn.ModifyImageAttribute(&ec2.ModifyImageAttributeInput{
		ImageId:   aws.String(image_id),
		Attribute: aws.String(ec2.ImageAttributeNameLaunchPermission),
		LaunchPermission: &ec2.LaunchPermissionModifications{
			Remove: launch_permission,
		},
	})
	if err != nil {
		return fmt.Errorf("error deleting AMI launch permission (%s): %w", d.Id(), err)
	}

	return nil
}

func HasLaunchPermission(conn *ec2.EC2, image_id string, read_value string, read_type string) (bool, error) {
	attrs, err := conn.DescribeImageAttribute(&ec2.DescribeImageAttributeInput{
		ImageId:   aws.String(image_id),
		Attribute: aws.String(ec2.ImageAttributeNameLaunchPermission),
	})
	if err != nil {
		// When an AMI disappears out from under a launch permission resource, we will
		// see either InvalidAMIID.NotFound or InvalidAMIID.Unavailable.
		if ec2err, ok := err.(awserr.Error); ok && strings.HasPrefix(ec2err.Code(), "InvalidAMIID") {
			log.Printf("[DEBUG] %s no longer exists, so we'll drop launch permission for %s from the state", image_id, read_value)
			return false, nil
		}
		return false, err
	}

	for _, lp := range attrs.LaunchPermissions {
		switch read_type {
		case "UserId":
			if aws.StringValue(lp.UserId) == read_value {
				return true, nil
			}
		case "OrganizationArn":
			if aws.StringValue(lp.OrganizationArn) == read_value {
				return true, nil
			}
		case "OrganizationalUnitArn":
			if aws.StringValue(lp.OrganizationalUnitArn) == read_value {
				return true, nil
			}
		}
	}
	return false, nil
}

func BuildLaunchPermission(d *schema.ResourceData) []*ec2.LaunchPermission {
	account_id := d.Get("account_id").(string)
	arn := d.Get("arn").(string)
	arn_type := d.Get("arn_type").(string)

	var launch_permission []*ec2.LaunchPermission

	if len(account_id) > 0 {
		log.Printf("[DEBUG] Building LaunchPermission of type UserId: %s", account_id)
		launch_permission = []*ec2.LaunchPermission{
			{UserId: aws.String(account_id)},
		}
	} else if arn_type == "OrganizationArn" {
		log.Printf("[DEBUG] Building LaunchPermission of type OrganizationArn: %s", arn)
		launch_permission = []*ec2.LaunchPermission{
			{OrganizationArn: aws.String(arn)},
		}
	} else if arn_type == "OrganizationalUnitArn" {
		log.Printf("[DEBUG] Building LaunchPermission of type OrganizationalUnitArn: %s", arn)
		launch_permission = []*ec2.LaunchPermission{
			{OrganizationalUnitArn: aws.String(arn)},
		}
	}

	return launch_permission
}
