package cloudfront

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

func DataSourceRealtimeLogConfig() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceRealtimeLogConfigRead,

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

func dataSourceRealtimeLogConfigRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudFrontConn()

	name := d.Get("name").(string)
	logConfig, err := FindRealtimeLogConfigByName(ctx, conn, name)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading CloudFront Real-time Log Config (%s): %s", name, err)
	}
	d.SetId(
		aws.StringValue(logConfig.ARN),
	)
	d.Set("arn", logConfig.ARN)
	if err := d.Set("endpoint", flattenEndPoints(logConfig.EndPoints)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting endpoint: %s", err)
	}
	d.Set("fields", aws.StringValueSlice(logConfig.Fields))
	d.Set("name", logConfig.Name)
	d.Set("sampling_rate", logConfig.SamplingRate)

	return diags
}
