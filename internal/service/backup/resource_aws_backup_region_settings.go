package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/backup"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceRegionSettings() *schema.Resource {
	return &schema.Resource{
		Create: resourceRegionSettingsUpdate,
		Update: resourceRegionSettingsUpdate,
		Read:   resourceRegionSettingsRead,
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

func resourceRegionSettingsUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).BackupConn

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

	d.SetId(meta.(*conns.AWSClient).Region)

	return resourceRegionSettingsRead(d, meta)
}

func resourceRegionSettingsRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).BackupConn

	resp, err := conn.DescribeRegionSettings(&backup.DescribeRegionSettingsInput{})
	if err != nil {
		return fmt.Errorf("error reading Backup Region Settings (%s): %w", d.Id(), err)
	}

	d.Set("resource_type_opt_in_preference", aws.BoolValueMap(resp.ResourceTypeOptInPreference))

	return nil
}
