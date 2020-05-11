package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/greengrass"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceAwsIotGreengrassServiceRole() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsIotGreengrassServiceRoleCreate,
		Read:   resourceAwsIotGreengrassServiceRoleRead,
		Update: resourceAwsIotGreengrassServiceRoleUpdate,
		Delete: resourceAwsIotGreengrassServiceRoleDelete,
		Schema: map[string]*schema.Schema{
			"associated_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"role_arn": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceAwsIotGreengrassServiceRoleCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).greengrassconn
	roleArn := d.Get("role_arn").(string)
	input := &greengrass.AssociateServiceRoleToAccountInput{
		RoleArn: aws.String(roleArn),
	}

	_, err := conn.AssociateServiceRoleToAccount(input)

	if err != nil {
		return fmt.Errorf("greengrass service role could not be associated with account: %v", err)
	}

	d.SetId(roleArn)

	return resourceAwsIotGreengrassServiceRoleRead(d, meta)
}

func resourceAwsIotGreengrassServiceRoleRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).greengrassconn
	input := &greengrass.GetServiceRoleForAccountInput{}

	output, err := conn.GetServiceRoleForAccount(input)

	if err != nil {
		if isAWSErrRequestFailureStatusCode(err, 404) {
			return fmt.Errorf("No greengrass service role is set for this account")
		}
		return fmt.Errorf("error while getting greengrass service role for account: %s", err)
	}

	roleArn := aws.StringValue(output.RoleArn)
	associatedAt := aws.StringValue(output.AssociatedAt)

	if err := d.Set("role_arn", roleArn); err != nil {
		return fmt.Errorf("error setting role_arn: %s", err)
	}
	if err := d.Set("associated_at", associatedAt); err != nil {
		return fmt.Errorf("error setting associated_at: %s", err)
	}

	return nil
}

func resourceAwsIotGreengrassServiceRoleUpdate(d *schema.ResourceData, meta interface{}) error {

	if d.HasChange("role_arn") {
		conn := meta.(*AWSClient).greengrassconn
		roleArn := d.Get("role_arn").(string)
		input := &greengrass.AssociateServiceRoleToAccountInput{
			RoleArn: aws.String(roleArn),
		}

		_, err := conn.AssociateServiceRoleToAccount(input)

		if err != nil {
			return fmt.Errorf("greengrass service role could not be associated with account: %v", err)
		}

		d.SetId(roleArn)
	}

	return resourceAwsIotGreengrassServiceRoleRead(d, meta)
}

func resourceAwsIotGreengrassServiceRoleDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).greengrassconn
	input := &greengrass.DisassociateServiceRoleFromAccountInput{}

	_, err := conn.DisassociateServiceRoleFromAccount(input)
	if err != nil {
		return fmt.Errorf("error disassociating greengrass service role from account: %v", err)
	}

	return nil
}
