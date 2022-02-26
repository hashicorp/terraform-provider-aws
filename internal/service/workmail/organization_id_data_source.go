package workmail

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/workmail"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
)

func DataSourceOrganization() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceOrganizationRead,

		Schema: map[string]*schema.Schema{
			"organization_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"alias": {
				Type:     schema.TypeString,
				Required: true,
			},
			"state": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"directory_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"directory_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"completed_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceOrganizationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).WorkMailConn
	organizationID := d.Get("organization_id").(string)
	var version string

	input := &workmail.DescribeOrganizationInput{
		organizationID: aws.String(organizationID),
	}

	if v, ok := d.GetOk("version_id"); ok {
		versionID := v.(string)
		input.VersionId = aws.String(versionID)
		version = versionID
	} else {
		versionStage := d.Get("version_stage").(string)
		input.VersionStage = aws.String(versionStage)
		version = versionStage
	}

	log.Printf("[DEBUG] Reading WorkMail Organization ID: %s", input)
	output, err := conn.DescribeOrganization(input)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, workmail.ErrCodeOrganizationNotFoundException) {
			return fmt.Errorf("WorkMail Organization ID %q  not found", organizationID, version)
		}
		if tfawserr.ErrMessageContains(err, secretsmanager.ErrCodeOrganizationStateException, "You can’t perform this operation on this organization because of invalid state") {
			return fmt.Errorf("WorkMail Organization ID %q  not found", organizationID, version)
		}
		return fmt.Errorf("error reading Organization ID: %w", err)
	}

	d.SetId(fmt.Sprintf("%s", organizationID))
	d.Set("arn", output.ARN)
	d.Set("state", output.State)
	d.Set("alias", output.Alias)
	d.Set("completed_data", output.CompletedDate)
	d.Set("directory_id", output.DirectoryId)
	d.Set("directory_type", output.DirectoryType)

	return nil
}
