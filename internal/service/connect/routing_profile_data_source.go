package connect

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/connect"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

func DataSourceRoutingProfile() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceRoutingProfileRead,
		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"default_outbound_queue_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"instance_id": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 100),
			},
			"media_concurrencies": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"channel": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"concurrency": {
							Type:     schema.TypeInt,
							Computed: true,
						},
					},
				},
			},
			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{"name", "routing_profile_id"},
			},
			"queue_configs": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"channel": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"delay": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"priority": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"queue_arn": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"queue_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"queue_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"routing_profile_id": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{"routing_profile_id", "name"},
			},
			"tags": tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceRoutingProfileRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ConnectConn
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	instanceID := d.Get("instance_id").(string)

	input := &connect.DescribeRoutingProfileInput{
		InstanceId: aws.String(instanceID),
	}

	if v, ok := d.GetOk("routing_profile_id"); ok {
		input.RoutingProfileId = aws.String(v.(string))
	} else if v, ok := d.GetOk("name"); ok {
		name := v.(string)
		routingProfileSummary, err := dataSourceGetConnectRoutingProfileSummaryByName(ctx, conn, instanceID, name)

		if err != nil {
			return diag.FromErr(fmt.Errorf("error finding Connect Routing Profile Summary by name (%s): %w", name, err))
		}

		if routingProfileSummary == nil {
			return diag.FromErr(fmt.Errorf("error finding Connect Routing Profile Summary by name (%s): not found", name))
		}

		input.RoutingProfileId = routingProfileSummary.Id
	}

	resp, err := conn.DescribeRoutingProfileWithContext(ctx, input)

	if err != nil {
		return diag.FromErr(fmt.Errorf("error getting Connect Routing Profile: %w", err))
	}

	if resp == nil || resp.RoutingProfile == nil {
		return diag.FromErr(fmt.Errorf("error getting Connect Routing Profile: empty response"))
	}

	routingProfile := resp.RoutingProfile

	if err := d.Set("media_concurrencies", flattenRoutingProfileMediaConcurrencies(routingProfile.MediaConcurrencies)); err != nil {
		return diag.FromErr(err)
	}

	d.Set("arn", routingProfile.RoutingProfileArn)
	d.Set("default_outbound_queue_id", routingProfile.DefaultOutboundQueueId)
	d.Set("description", routingProfile.Description)
	d.Set("instance_id", instanceID)
	d.Set("name", routingProfile.Name)
	d.Set("routing_profile_id", routingProfile.RoutingProfileId)

	// getting the routing profile queues uses a separate API: ListRoutingProfileQueues
	queueConfigs, err := getRoutingProfileQueueConfigs(ctx, conn, instanceID, *routingProfile.RoutingProfileId)

	if err != nil {
		return diag.FromErr(fmt.Errorf("error finding Connect Routing Profile Queue Configs Summary by Routing Profile ID (%s): %w", *routingProfile.RoutingProfileId, err))
	}

	d.Set("queue_configs", queueConfigs)

	if err := d.Set("tags", KeyValueTags(routingProfile.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return diag.FromErr(fmt.Errorf("error setting tags: %s", err))
	}

	d.SetId(fmt.Sprintf("%s:%s", instanceID, aws.StringValue(routingProfile.RoutingProfileId)))

	return nil
}

func dataSourceGetConnectRoutingProfileSummaryByName(ctx context.Context, conn *connect.Connect, instanceID, name string) (*connect.RoutingProfileSummary, error) {
	var result *connect.RoutingProfileSummary

	input := &connect.ListRoutingProfilesInput{
		InstanceId: aws.String(instanceID),
		MaxResults: aws.Int64(ListRoutingProfilesMaxResults),
	}

	err := conn.ListRoutingProfilesPagesWithContext(ctx, input, func(page *connect.ListRoutingProfilesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, qs := range page.RoutingProfileSummaryList {
			if qs == nil {
				continue
			}

			if aws.StringValue(qs.Name) == name {
				result = qs
				return false
			}
		}

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}
