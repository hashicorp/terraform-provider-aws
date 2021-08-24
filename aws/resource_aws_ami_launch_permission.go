package aws

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceAwsAmiLaunchPermission() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsAmiLaunchPermissionCreate,
		Read:   resourceAwsAmiLaunchPermissionRead,
		Delete: resourceAwsAmiLaunchPermissionDelete,
		Importer: &schema.ResourceImporter{
			State: func(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				idParts := strings.Split(d.Id(), "/")

				parseError := fmt.Errorf("Unexpected format of ID (%q), expected ACCOUNT-ID/IMAGE-ID or group/GROUP-NAME/ACCOUNT-ID", d.Id())
				if len(idParts) == 2 {
					// Parsing the ACCOUNT-ID/IMAGE-ID branch
					if idParts[0] == "" || idParts[1] == "" {
						return nil, parseError
					}
					accountId := idParts[0]
					imageId := idParts[1]
					d.Set("account_id", accountId)
					d.Set("image_id", imageId)
					d.SetId(fmt.Sprintf("%s-account-%s", imageId, accountId))
				} else if len(idParts) == 3 && idParts[0] == "group" {
					// Parsing the group/GROUP-NAME/ACCOUNT-ID branch
					if idParts[1] == "" || idParts[2] == "" {
						return nil, parseError
					}
					groupName := idParts[1]
					imageId := idParts[2]
					d.Set("group_name", groupName)
					d.Set("image_id", imageId)
					d.SetId(fmt.Sprintf("%s-group-%s", imageId, groupName))
				} else {
					return nil, parseError
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
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				ExactlyOneOf: []string{
					"account_id",
					"group_name",
				},
			},
			"group_name": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				ExactlyOneOf: []string{
					"account_id",
					"group_name",
				},
			},
		},
	}
}

func resourceAwsAmiLaunchPermissionCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	image_id := d.Get("image_id").(string)
	account_id := d.Get("account_id").(string)
	group_name := d.Get("group_name").(string)

	var launch_permission *ec2.LaunchPermission

	if account_id != "" {
		launch_permission = &ec2.LaunchPermission{UserId: aws.String(account_id)}
	} else {
		launch_permission = &ec2.LaunchPermission{Group: aws.String(group_name)}
	}

	_, err := conn.ModifyImageAttribute(&ec2.ModifyImageAttributeInput{
		ImageId:   aws.String(image_id),
		Attribute: aws.String(ec2.ImageAttributeNameLaunchPermission),
		LaunchPermission: &ec2.LaunchPermissionModifications{
			Add: []*ec2.LaunchPermission{
				launch_permission,
			},
		},
	})
	if err != nil {
		return fmt.Errorf("error creating AMI launch permission: %w", err)
	}

	if account_id != "" {
		d.SetId(fmt.Sprintf("%s-account-%s", image_id, account_id))
	} else {
		d.SetId(fmt.Sprintf("%s-group-%s", image_id, group_name))
	}

	return nil
}

func resourceAwsAmiLaunchPermissionRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	exists, err := hasLaunchPermission(conn, d.Get("image_id").(string), d.Get("account_id").(string), d.Get("group_name").(string))
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

func resourceAwsAmiLaunchPermissionDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	image_id := d.Get("image_id").(string)
	account_id := d.Get("account_id").(string)
	group_name := d.Get("group_name").(string)

	var launch_permission *ec2.LaunchPermission

	if account_id != "" {
		launch_permission = &ec2.LaunchPermission{UserId: aws.String(account_id)}
	} else {
		launch_permission = &ec2.LaunchPermission{Group: aws.String(group_name)}
	}
	_, err := conn.ModifyImageAttribute(&ec2.ModifyImageAttributeInput{
		ImageId:   aws.String(image_id),
		Attribute: aws.String(ec2.ImageAttributeNameLaunchPermission),
		LaunchPermission: &ec2.LaunchPermissionModifications{
			Remove: []*ec2.LaunchPermission{
				launch_permission,
			},
		},
	})
	if err != nil {
		return fmt.Errorf("error deleting AMI launch permission (%s): %w", d.Id(), err)
	}

	return nil
}

func hasLaunchPermission(conn *ec2.EC2, image_id string, account_id string, group_name string) (bool, error) {
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
		if account_id != "" && aws.StringValue(lp.UserId) == account_id {
			return true, nil
		} else if group_name != "" && aws.StringValue(lp.Group) == group_name {
			return true, nil
		}
	}
	return false, nil
}
