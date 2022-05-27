package cloudfront

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func DataSourceRealtimeLogConfig() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceRealtimeLogConfigRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"endpoint": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"kinesis_stream_config": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"role_arn": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"stream_arn": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
						"stream_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"fields": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"sampling_rate": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func dataSourceRealtimeLogConfigRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CloudFrontConn

	name := d.Get("name").(string)
	logConfig, err := FindRealtimeLogConfigByName(conn, name)
	if err != nil {
		return fmt.Errorf("error reading CloudFront Real-time Log Config (%s): %w", name, err)
	}
	d.SetId(
		aws.StringValue(logConfig.ARN),
	)
	d.Set("arn", logConfig.ARN)
	if err := d.Set("endpoint", flattenEndPoints(logConfig.EndPoints)); err != nil {
		return fmt.Errorf("error setting endpoint: %w", err)
	}
	d.Set("fields", aws.StringValueSlice(logConfig.Fields))
	d.Set("name", logConfig.Name)
	d.Set("sampling_rate", logConfig.SamplingRate)

	return nil
}
