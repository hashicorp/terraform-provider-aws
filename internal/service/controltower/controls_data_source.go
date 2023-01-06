package controltower

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/controltower"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func DataSourceControls() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: DataSourceControlsRead,

		Schema: map[string]*schema.Schema{
			"enabled_controls": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"target_identifier": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
		},
	}
}

func DataSourceControlsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ControlTowerConn()

	targetIdentifier := d.Get("target_identifier").(string)
	input := &controltower.ListEnabledControlsInput{
		TargetIdentifier: aws.String(targetIdentifier),
	}

	var controls []string
	err := conn.ListEnabledControlsPagesWithContext(ctx, input, func(page *controltower.ListEnabledControlsOutput, lastPage bool) bool {
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
		return diag.Errorf("listing ControlTower Controls (%s): %s", targetIdentifier, err)
	}

	d.SetId(targetIdentifier)
	d.Set("enabled_controls", controls)

	return nil
}
