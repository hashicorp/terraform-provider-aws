package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/backup"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	awsprovider "github.com/terraform-providers/terraform-provider-aws/provider"
)

func resourceAwsBackupRegionSettings() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsBackupRegionSettingsUpdate,
		Update: resourceAwsBackupRegionSettingsUpdate,
		Read:   resourceAwsBackupRegionSettingsRead,
		Delete: schema.Noop,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"resource_type_opt_in_preference": {
				Type:     schema.TypeMap,
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeBool},
			},
		},
	}
}

func resourceAwsBackupRegionSettingsUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*awsprovider.AWSClient).BackupConn

	prefrences := d.Get("resource_type_opt_in_preference").(map[string]interface{})
	list := make(map[string]*bool, len(prefrences))
	for i, v := range prefrences {
		list[i] = aws.Bool(v.(bool))
	}

	input := &backup.UpdateRegionSettingsInput{
		ResourceTypeOptInPreference: list,
	}

	_, err := conn.UpdateRegionSettings(input)
	if err != nil {
		return fmt.Errorf("error setting Backup Region Settings (%s): %w", d.Id(), err)
	}

	d.SetId(meta.(*awsprovider.AWSClient).Region)

	return resourceAwsBackupRegionSettingsRead(d, meta)
}

func resourceAwsBackupRegionSettingsRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*awsprovider.AWSClient).BackupConn

	resp, err := conn.DescribeRegionSettings(&backup.DescribeRegionSettingsInput{})
	if err != nil {
		return fmt.Errorf("error reading Backup Region Settings (%s): %w", d.Id(), err)
	}

	d.Set("resource_type_opt_in_preference", aws.BoolValueMap(resp.ResourceTypeOptInPreference))

	return nil
}
