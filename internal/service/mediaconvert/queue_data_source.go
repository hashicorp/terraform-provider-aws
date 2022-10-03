package mediaconvert

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/mediaconvert"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

func DataSourceQueue() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceQueueRead,

		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  mediaconvert.QueueStatusActive,
				ValidateFunc: validation.StringInSlice([]string{
					mediaconvert.QueueStatusActive,
					mediaconvert.QueueStatusPaused,
				}, false),
			},
			"tags": tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceQueueRead(d *schema.ResourceData, meta interface{}) error {
	conn, err := GetAccountClient(meta.(*conns.AWSClient))
	if err != nil {
		return fmt.Errorf("error getting Media Convert Account Client: %w", err)
	}

	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	id := d.Get("id").(string)

	resp, err := conn.GetQueue(&mediaconvert.GetQueueInput{
		Name: aws.String(id),
	})

	if err != nil {
		return fmt.Errorf("error getting Media Convert Queue (%s): %w", id, err)
	}

	if resp == nil || resp.Queue == nil {
		return fmt.Errorf("error getting Media Convert Queue (%s): empty response", id)
	}

	d.SetId(aws.StringValue(resp.Queue.Name))
	d.Set("name", resp.Queue.Name)
	d.Set("arn", resp.Queue.Arn)
	d.Set("status", resp.Queue.Status)

	tags, err := ListTags(conn, aws.StringValue(resp.Queue.Arn))

	if err != nil {
		return fmt.Errorf("error listing tags for Media Convert Queue (%s): %w", id, err)
	}

	if err := d.Set("tags", tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	return nil
}
