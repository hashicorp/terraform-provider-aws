package location

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

func DataSourceRouteCalculator() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceRouteCalculatorRead,

		Schema: map[string]*schema.Schema{
			"calculator_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"calculator_name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 100),
			},
			"create_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"data_source": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"update_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceRouteCalculatorRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).LocationConn()

	out, err := findRouteCalculatorByName(ctx, conn, d.Get("calculator_name").(string))
	if err != nil {
		return diag.Errorf("reading Location Service Route Calculator (%s): %s", d.Get("calculator_name").(string), err)
	}

	if out == nil {
		return diag.Errorf("reading Location Service Route Calculator (%s): empty response", d.Get("calculator_name").(string))
	}

	d.SetId(aws.StringValue(out.CalculatorName))
	d.Set("calculator_arn", out.CalculatorArn)
	d.Set("calculator_name", out.CalculatorName)
	d.Set("create_time", aws.TimeValue(out.CreateTime).Format(time.RFC3339))
	d.Set("data_source", out.DataSource)
	d.Set("description", out.Description)
	d.Set("update_time", aws.TimeValue(out.UpdateTime).Format(time.RFC3339))

	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	if err := d.Set("tags", KeyValueTags(out.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return diag.Errorf("listing tags for Location Service Route Calculator (%s): %s", d.Id(), err)
	}

	return nil
}
