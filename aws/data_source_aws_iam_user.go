package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/keyvaluetags"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func dataSourceAwsIAMUser() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsIAMUserRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"path": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"permissions_boundary": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"user_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"user_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"tags": tagsSchemaComputed(),
		},
	}
}

func dataSourceAwsIAMUserRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).IAMConn
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	userName := d.Get("user_name").(string)
	req := &iam.GetUserInput{
		UserName: aws.String(userName),
	}

	log.Printf("[DEBUG] Reading IAM User: %s", req)
	resp, err := conn.GetUser(req)
	if err != nil {
		return fmt.Errorf("error getting user: %w", err)
	}

	user := resp.User
	d.SetId(aws.StringValue(user.UserId))
	d.Set("arn", user.Arn)
	d.Set("path", user.Path)
	d.Set("permissions_boundary", "")
	if user.PermissionsBoundary != nil {
		d.Set("permissions_boundary", user.PermissionsBoundary.PermissionsBoundaryArn)
	}
	d.Set("user_id", user.UserId)
	if err := d.Set("tags", keyvaluetags.IamKeyValueTags(user.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	return nil
}
