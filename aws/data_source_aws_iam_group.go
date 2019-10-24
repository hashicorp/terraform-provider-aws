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
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"path": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"user_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"user_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"arn": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"create_date": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"password_last_used": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"permissions_boundary": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"permissions_boundary_type": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"permissions_boundary_arn": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
						"tags": tagsSchemaComputed(),
					},
				},
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

	d.SetId(*group.GroupId)
	d.Set("arn", group.Arn)
	d.Set("path", group.Path)
	d.Set("group_id", group.GroupId)

	if err := d.Set("users", getUsers(resp.Users)); err != nil {
		return err
	}

	return nil
}

func getUsers(u []*iam.User) interface{} {
	s := []interface{}{}
	for _, v := range u {
		user := map[string]interface{}{
			"path":        aws.StringValue(v.Path),
			"user_name":   aws.StringValue(v.UserName),
			"user_id":     aws.StringValue(v.UserId),
			"arn":         aws.StringValue(v.Arn),
			"create_date": aws.TimeValue(v.CreateDate).String(),
			"tags":        tagsToMapIAM(v.Tags),
		}

		if v.PasswordLastUsed != nil {
			user["password_last_used"] = aws.TimeValue(v.PasswordLastUsed).String()
		}

		if v.PermissionsBoundary != nil {
			pb := map[string]interface{}{
				"permissions_boundary_type": aws.StringValue(v.PermissionsBoundary.PermissionsBoundaryType),
				"permissions_boundary_arn":  aws.StringValue(v.PermissionsBoundary.PermissionsBoundaryArn),
			}
			user["permissions_boundary"] = pb
		}
		s = append(s, user)
	}
	return s
}
