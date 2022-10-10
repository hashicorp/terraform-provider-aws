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

func DataSourcePlaceIndex() *schema.Resource {
	return &schema.Resource{
		Read: dataSourcePlaceIndexRead,
		Schema: map[string]*schema.Schema{
			"create_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"data_source": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"data_source_configuration": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"intended_use": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"index_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"index_name": {
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

func dataSourcePlaceIndexRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LocationConn

	input := &locationservice.DescribePlaceIndexInput{}

	if v, ok := d.GetOk("index_name"); ok {
		input.IndexName = aws.String(v.(string))
	}

	output, err := conn.DescribePlaceIndex(input)

	if err != nil {
		return fmt.Errorf("error getting Location Service Place Index: %w", err)
	}

	if output == nil {
		return fmt.Errorf("error getting Location Service Place Index: empty response")
	}

	d.SetId(aws.StringValue(output.IndexName))
	d.Set("create_time", aws.TimeValue(output.CreateTime).Format(time.RFC3339))
	d.Set("data_source", output.DataSource)

	if output.DataSourceConfiguration != nil {
		d.Set("data_source_configuration", []interface{}{flattenDataSourceConfiguration(output.DataSourceConfiguration)})
	} else {
		d.Set("data_source_configuration", nil)
	}

	d.Set("description", output.Description)
	d.Set("index_arn", output.IndexArn)
	d.Set("index_name", output.IndexName)
	d.Set("tags", KeyValueTags(output.Tags).IgnoreAWS().IgnoreConfig(meta.(*conns.AWSClient).IgnoreTagsConfig).Map())
	d.Set("update_time", aws.TimeValue(output.UpdateTime).Format(time.RFC3339))

	return nil
}
