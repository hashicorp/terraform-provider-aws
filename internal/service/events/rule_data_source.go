package events

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go/service/eventbridge"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

func DataSourceRule() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceRuleRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_by": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
			"event_bus_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"event_pattern": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
			"managed_by": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"role_arn": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
			"schedule_expression": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
			"state": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceRuleRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EventsConn
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	name := d.Get("name").(string)

	input := &eventbridge.DescribeRuleInput{
		Name: aws.String(name),
	}

	if v, ok := d.GetOk("event_bus_name"); ok {
		input.EventBusName = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Reading EventBridge rule (%s)", d.Id())
	output, err := conn.DescribeRule(input)
	if err != nil {
		return fmt.Errorf("error getting EventBridge rule (%s): %w", d.Id(), err)
	}

	if output == nil {
		return fmt.Errorf("error getting EventBridge rule (%s): empty response", d.Id())
	}

	log.Printf("[DEBUG] Found EventBridge rule: %#v", *output)

	d.Set("arn", output.Arn)
	d.Set("created_by", output.CreatedBy)
	d.Set("description", output.Description)
	d.Set("event_bus_name", output.EventBusName)
	d.Set("event_pattern", output.EventPattern)
	d.Set("managed_by", output.ManagedBy)
	d.Set("name", output.Name)
	d.Set("role_arn", output.RoleArn)
	d.Set("schedule_expression", output.ScheduleExpression)
	d.Set("state", output.State)

	tags, err := ListTags(conn, *output.Arn)
	if err != nil {
		return fmt.Errorf("error listing tags for rule (%s): %w", name, err)
	}
	if err := d.Set("tags", tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	d.SetId(*output.Arn)

	return nil
}
