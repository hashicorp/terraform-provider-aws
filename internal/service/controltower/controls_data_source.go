package controltower

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/controltower"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func DataSourceControls() *schema.Resource {
	return &schema.Resource{
		Read: DataSourceControlsRead,

		Schema: map[string]*schema.Schema{
			"target_identifier": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			"enabled_controls": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func DataSourceControlsRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ControlTowerConn
	target_identifier := aws.String(d.Get("target_identifier").(string))

	input := &controltower.ListEnabledControlsInput{
		TargetIdentifier: target_identifier,
	}

	var controls []string
	err := conn.ListEnabledControlsPages(input, func(page *controltower.ListEnabledControlsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}
		for _, control := range page.EnabledControls {
			if control == nil {
				continue
			}
			controls = append(controls, aws.StringValue(control.ControlIdentifier))
		}
		return !lastPage
	})

	if err != nil {
		return fmt.Errorf("error listing ControlTower Target Identifier: %w", err)
	}
	if len(controls) == 0 {
		return fmt.Errorf("no Enabled Controls found matching criteria; try different search")
	}

	d.SetId(aws.StringValue(target_identifier))
	d.Set("enabled_controls", controls)
	return nil
}
