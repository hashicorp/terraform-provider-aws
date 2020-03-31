package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/naming"
)

func resourceAwsSecurityGroupImportState(
	d *schema.ResourceData,
	meta interface{}) ([]*schema.ResourceData, error) {
	conn := meta.(*AWSClient).ec2conn

	// First query the security group
	sgRaw, _, err := SGStateRefreshFunc(conn, d.Id())()
	if err != nil {
		return nil, err
	}
	if sgRaw == nil {
		return nil, fmt.Errorf("security group not found")
	}
	sg := sgRaw.(*ec2.SecurityGroup)

	// Perform nil check to avoid ImportStateVerify difference when unconfigured
	if namePrefix := naming.NamePrefixFromName(aws.StringValue(sg.GroupName)); namePrefix != nil {
		d.Set("name_prefix", namePrefix)
	}

	// Start building our results
	results := make([]*schema.ResourceData, 1)
	results[0] = d

	return results, nil
}
