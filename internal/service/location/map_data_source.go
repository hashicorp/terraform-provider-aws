package location

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/locationservice"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

func DataSourceMap() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceMapRead,
		Schema: map[string]*schema.Schema{
			"configuration": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"style": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"create_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"map_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"map_name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 100),
			},
			"tags": tftags.TagsSchemaComputed(),
			"update_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceMapRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LocationConn

	input := &locationservice.DescribeMapInput{}

	if v, ok := d.GetOk("map_name"); ok {
		input.MapName = aws.String(v.(string))
	}

	output, err := conn.DescribeMap(input)

	if err != nil {
		return fmt.Errorf("error getting Location Service Map: %w", err)
	}

	if output == nil {
		return fmt.Errorf("error getting Location Service Map: empty response")
	}

	d.SetId(aws.StringValue(output.MapName))

	if output.Configuration != nil {
		d.Set("configuration", []interface{}{flattenConfiguration(output.Configuration)})
	} else {
		d.Set("configuration", nil)
	}

	d.Set("create_time", aws.TimeValue(output.CreateTime).Format(time.RFC3339))
	d.Set("description", output.Description)
	d.Set("map_arn", output.MapArn)
	d.Set("map_name", output.MapName)
	d.Set("update_time", aws.TimeValue(output.UpdateTime).Format(time.RFC3339))
	d.Set("tags", KeyValueTags(output.Tags).IgnoreAWS().IgnoreConfig(meta.(*conns.AWSClient).IgnoreTagsConfig).Map())

	return nil
}
