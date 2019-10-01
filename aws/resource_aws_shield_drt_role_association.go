package aws

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/shield"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsShieldDrtRoleAssociation() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsShieldDrtRoleAssociationCreate,
		Read:   resourceAwsShieldDrtRoleAssociationRead,
		Update: resourceAwsShieldDrtRoleAssociationCreate,
		Delete: resourceAwsShieldDrtRoleAssociationDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"role_arn": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceAwsShieldDrtRoleAssociationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).shieldconn

	input := &shield.AssociateDRTRoleInput{
		RoleArn: aws.String(d.Get("role_arn").(string)),
	}

	_, err := conn.AssociateDRTRole(input)
	if err != nil {
		return fmt.Errorf("error associating DRT Role: %v", err)
	}

	d.SetId(time.Now().UTC().String())
	return resourceAwsShieldDrtRoleAssociationRead(d, meta)
}

func resourceAwsShieldDrtRoleAssociationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).shieldconn

	input := &shield.DescribeDRTAccessInput{}

	resp, err := conn.DescribeDRTAccess(input)
	if err != nil {
		return fmt.Errorf("error reading DRT Access: %v", err)
	}

	d.Set("role_arn", resp.RoleArn)

	return nil
}

func resourceAwsShieldDrtRoleAssociationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).shieldconn

	input := &shield.DescribeDRTAccessInput{}

	_, err := conn.DescribeDRTAccess(input)
	if err != nil {
		return fmt.Errorf("error disassociating DRT Role: %v", err)
	}

	d.SetId("")

	return nil
}
