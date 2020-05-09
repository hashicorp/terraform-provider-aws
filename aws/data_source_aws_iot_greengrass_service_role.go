package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/greengrass"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func dataSourceAwsIotGreengrassServiceRole() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsIotGreengrassServiceRoleRead,
		Schema: map[string]*schema.Schema{
			"associated_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"role_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceAwsIotGreengrassServiceRoleRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).greengrassconn
	input := &greengrass.GetServiceRoleForAccountInput{}

	output, err := conn.GetServiceRoleForAccount(input)

	if err != nil {
		return fmt.Errorf("error while getting greengrass service role for account: %s", err)
	}

	roleArn := aws.StringValue(output.RoleArn)
	associatedAt := aws.StringValue(output.AssociatedAt)

	d.SetId(roleArn)
	if err := d.Set("role_arn", roleArn); err != nil {
		return fmt.Errorf("error setting role_arn: %s", err)
	}
	if err := d.Set("associated_at", associatedAt); err != nil {
		return fmt.Errorf("error setting associated_at: %s", err)
	}

	return nil
}
