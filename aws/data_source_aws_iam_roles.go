package aws

import (
	"fmt"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

func DataSourceRoles() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceRolesRead,

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

func dataSourceRolesRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).IAMConn

	input := &iam.ListRolesInput{}

	if v, ok := d.GetOk("path_prefix"); ok {
		input.PathPrefix = aws.String(v.(string))
	}

	var results []*iam.Role

	err := conn.ListRolesPages(input, func(page *iam.ListRolesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, role := range page.Roles {
			if role == nil {
				continue
			}

			if v, ok := d.GetOk("name_regex"); ok && !regexp.MustCompile(v.(string)).MatchString(aws.StringValue(role.RoleName)) {
				continue
			}

			results = append(results, role)
		}

		return !lastPage
	})

	if err != nil {
		return fmt.Errorf("error reading IAM roles: %w", err)
	}

	d.SetId(meta.(*conns.AWSClient).Region)

	var arns, names []string

	for _, r := range results {
		arns = append(arns, aws.StringValue(r.Arn))
		names = append(names, aws.StringValue(r.RoleName))
	}

	if err := d.Set("arns", arns); err != nil {
		return fmt.Errorf("error setting arns: %w", err)
	}

	if err := d.Set("names", names); err != nil {
		return fmt.Errorf("error setting names: %w", err)
	}

	return nil
}
