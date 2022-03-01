package ec2

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func DataSourceSerialConsoleAccess() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceSerialConsoleAccessRead,

		Schema: map[string]*schema.Schema{
			"enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
		},
	}
}
func dataSourceSerialConsoleAccessRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	res, err := conn.GetSerialConsoleAccessStatus(&ec2.GetSerialConsoleAccessStatusInput{})
	if err != nil {
		return fmt.Errorf("Error reading serial console access toggle: %w", err)
	}

	d.SetId(meta.(*conns.AWSClient).Region)
	d.Set("enabled", res.SerialConsoleAccessEnabled)

	return nil
}
