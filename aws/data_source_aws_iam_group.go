package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func dataSourceAwsIAMGroup() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsIAMGroupRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"path": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"group_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"group_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"users": {
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeMap,
				},
				Computed: true,
			},
		},
	}
}

func dataSourceAwsIAMGroupRead(d *schema.ResourceData, meta interface{}) error {
	iamconn := meta.(*AWSClient).iamconn

	groupName := d.Get("group_name").(string)

	req := &iam.GetGroupInput{
		GroupName: aws.String(groupName),
	}

	log.Printf("[DEBUG] Reading IAM Group: %s", req)
	resp, err := iamconn.GetGroup(req)
	if err != nil {
		return fmt.Errorf("Error getting group: %s", err)
	}
	if resp == nil {
		return fmt.Errorf("no IAM group found")
	}

	group := resp.Group
	users := resp.Users

	var usersList []map[string]*string

	d.SetId(*group.GroupId)
	d.Set("arn", group.Arn)
	d.Set("path", group.Path)
	d.Set("group_id", group.GroupId)

	for _, u := range users {
		usersList = append(usersList, map[string]*string{
			"Arn":      u.Arn,
			"UserId":   u.UserId,
			"UserName": u.UserName,
		})
	}

	if err := d.Set("users", usersList); err != nil {
		return fmt.Errorf("Error setting users for resource %s: %s", d.Id(), err)
	}

	return err
}
