package kinesis

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

func DataSourceStream() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceStreamRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"closed_shards": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"creation_timestamp": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"open_shards": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"retention_period": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"shard_level_metrics": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"stream_mode_details": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"stream_mode": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"tags": tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceStreamRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).KinesisConn
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig
	name := d.Get("name").(string)

	output, err := FindStreamByName(conn, name)

	if err != nil {
		return fmt.Errorf("error reading Kinesis Stream (%s): %w", name, err)
	}

	d.SetId(aws.StringValue(output.StreamARN))
	d.Set("arn", output.StreamARN)

	var closedShards []*string
	for _, v := range filterShards(output.Shards, false) {
		closedShards = append(closedShards, v.ShardId)
	}

	d.Set("closed_shards", aws.StringValueSlice(closedShards))
	d.Set("creation_timestamp", aws.TimeValue(output.StreamCreationTimestamp).Unix())
	d.Set("name", output.StreamName)

	var openShards []*string
	for _, v := range filterShards(output.Shards, true) {
		openShards = append(openShards, v.ShardId)
	}
	d.Set("open_shards", aws.StringValueSlice(openShards))

	d.Set("retention_period", output.RetentionPeriodHours)

	var shardLevelMetrics []*string
	for _, v := range output.EnhancedMonitoring {
		shardLevelMetrics = append(shardLevelMetrics, v.ShardLevelMetrics...)
	}
	d.Set("shard_level_metrics", aws.StringValueSlice(shardLevelMetrics))

	d.Set("status", output.StreamStatus)

	if details := output.StreamModeDetails; details != nil {
		if err := d.Set("stream_mode_details", []interface{}{flattenStreamModeDetails(details)}); err != nil {
			return fmt.Errorf("error setting stream_mode_details: %w", err)
		}
	} else {
		d.Set("stream_mode_details", nil)
	}

	tags, err := ListTags(conn, name)

	if err != nil {
		return fmt.Errorf("error listing tags for Kinesis Stream (%s): %w", name, err)
	}

	if err := d.Set("tags", tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	return nil
}
