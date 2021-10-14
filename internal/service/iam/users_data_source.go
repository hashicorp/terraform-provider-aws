package iam

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func DataSourceUsers() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceUsersRead,
		Schema: map[string]*schema.Schema{
			"arns": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"name_regex": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringIsValidRegExp,
			},
			"names": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"path_prefix": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func dataSourceUsersRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).IAMConn

	nameRegex := d.Get("name_regex").(string)
	pathPrefix := d.Get("path_prefix").(string)

	results, err := FindUsers(conn, nameRegex, pathPrefix)

	if err != nil {
		return fmt.Errorf("error reading IAM users: %w", err)
	}

	d.SetId(meta.(*conns.AWSClient).Region)

	var arns, names []string

	for _, r := range results {
		names = append(names, aws.StringValue(r.UserName))
		arns = append(arns, aws.StringValue(r.Arn))
	}

	if err := d.Set("names", names); err != nil {
		return fmt.Errorf("error setting names: %w", err)
	}

	if err := d.Set("arns", arns); err != nil {
		return fmt.Errorf("error setting arns: %w", err)
	}

	return nil
}
