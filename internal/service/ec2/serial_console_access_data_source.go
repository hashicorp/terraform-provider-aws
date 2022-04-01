package ec2

import (
	"context"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func DataSourceSerialConsoleAccess() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceSerialConsoleAccessRead,

		Schema: map[string]*schema.Schema{
			"enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
		},
	}
}
func dataSourceSerialConsoleAccessRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EC2Conn

	output, err := conn.GetSerialConsoleAccessStatus(&ec2.GetSerialConsoleAccessStatusInput{})

	if err != nil {
		return diag.Errorf("error reading EC2 Serial Console Access: %s", err)
	}

	d.SetId(meta.(*conns.AWSClient).Region)
	d.Set("enabled", output.SerialConsoleAccessEnabled)

	return nil
}
