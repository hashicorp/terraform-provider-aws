package workmail

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/workmail"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func DataSourceOrganization() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceOrganizationRead,
		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"alias": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"completed_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"default_mail_domain": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"directory_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"directory_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"error_message": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"organization_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"state": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceOrganizationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).WorkMailConn
	organizationID := d.Get("organization_id").(string)

	input := &workmail.DescribeOrganizationInput{
		OrganizationId: aws.String(organizationID),
	}

	log.Printf("[DEBUG] Reading WorkMail Organization ID: %s", input)
	output, err := conn.DescribeOrganization(input)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, workmail.ErrCodeOrganizationNotFoundException) {
			return fmt.Errorf("WorkMail Organization ID %q  not found", organizationID)
		}
		if tfawserr.ErrMessageContains(err, workmail.ErrCodeOrganizationStateException, "You can’t perform this operation on this organization because of invalid state") {
			return fmt.Errorf("WorkMail Organization ID %q  not found", organizationID)
		}
		return fmt.Errorf("error reading Organization ID: %w", err)
	}

	d.SetId(organizationID)
	d.Set("arn", output.ARN)
	d.Set("alias", output.Alias)
	d.Set("state", output.State)
	d.Set("completed_data", output.CompletedDate)
	d.Set("default_mail_domain", output.DefaultMailDomain)
	d.Set("directory_id", output.DirectoryId)
	d.Set("directory_type", output.DirectoryType)

	return nil
}
