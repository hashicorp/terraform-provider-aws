package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/securityhub"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func ResourceOrganizationConfiguration() *schema.Resource {
	return &schema.Resource{
		Create: resourceOrganizationConfigurationUpdate,
		Read:   resourceOrganizationConfigurationRead,
		Update: resourceOrganizationConfigurationUpdate,
		Delete: schema.Noop,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"auto_enable": {
				Type:     schema.TypeBool,
				Required: true,
			},
		},
	}
}

func resourceOrganizationConfigurationUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SecurityHubConn

	input := &securityhub.UpdateOrganizationConfigurationInput{
		AutoEnable: aws.Bool(d.Get("auto_enable").(bool)),
	}

	_, err := conn.UpdateOrganizationConfiguration(input)

	if err != nil {
		return fmt.Errorf("error updating Security Hub Organization Configuration (%s): %w", d.Id(), err)
	}

	d.SetId(meta.(*conns.AWSClient).AccountID)

	return resourceOrganizationConfigurationRead(d, meta)
}

func resourceOrganizationConfigurationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SecurityHubConn

	output, err := conn.DescribeOrganizationConfiguration(&securityhub.DescribeOrganizationConfigurationInput{})

	if err != nil {
		return fmt.Errorf("error reading Security Hub Organization Configuration: %w", err)
	}

	d.Set("auto_enable", output.AutoEnable)

	return nil
}
