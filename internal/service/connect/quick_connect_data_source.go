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

func DataSourceQuickConnect() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceQuickConnectRead,
		Schema: map[string]*schema.Schema{
			"arn": {
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
			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{"name", "quick_connect_id"},
			},
			"quick_connect_config": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"phone_config": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"phone_number": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
						"queue_config": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"contact_flow_id": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"queue_id": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
						"quick_connect_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"user_config": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"contact_flow_id": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"user_id": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
					},
				},
			},
			"quick_connect_id": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{"quick_connect_id", "name"},
			},
			"tags": tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceQuickConnectRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ConnectConn
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	instanceID := d.Get("instance_id").(string)

	input := &connect.DescribeQuickConnectInput{
		InstanceId: aws.String(instanceID),
	}

	if v, ok := d.GetOk("quick_connect_id"); ok {
		input.QuickConnectId = aws.String(v.(string))
	} else if v, ok := d.GetOk("name"); ok {
		name := v.(string)
		quickConnectSummary, err := dataSourceGetQuickConnectSummaryByName(ctx, conn, instanceID, name)

		if err != nil {
			return diag.FromErr(fmt.Errorf("error finding Connect Quick Connect Summary by name (%s): %w", name, err))
		}

		if quickConnectSummary == nil {
			return diag.FromErr(fmt.Errorf("error finding Connect Quick Connect Summary by name (%s): not found", name))
		}

		input.QuickConnectId = quickConnectSummary.Id
	}

	resp, err := conn.DescribeQuickConnectWithContext(ctx, input)

	if err != nil {
		return diag.FromErr(fmt.Errorf("error getting Connect Quick Connect: %w", err))
	}

	if resp == nil || resp.QuickConnect == nil {
		return diag.FromErr(fmt.Errorf("error getting Connect Quick Connect: empty response"))
	}

	quickConnect := resp.QuickConnect

	d.Set("arn", quickConnect.QuickConnectARN)
	d.Set("description", quickConnect.Description)
	d.Set("name", quickConnect.Name)
	d.Set("quick_connect_id", quickConnect.QuickConnectId)

	if err := d.Set("quick_connect_config", flattenQuickConnectConfig(quickConnect.QuickConnectConfig)); err != nil {
		return diag.FromErr(fmt.Errorf("error setting quick_connect_config: %s", err))
	}

	if err := d.Set("tags", KeyValueTags(quickConnect.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return diag.FromErr(fmt.Errorf("error setting tags: %s", err))
	}

	d.SetId(fmt.Sprintf("%s:%s", instanceID, aws.StringValue(quickConnect.QuickConnectId)))

	return nil
}

func dataSourceGetQuickConnectSummaryByName(ctx context.Context, conn *connect.Connect, instanceID, name string) (*connect.QuickConnectSummary, error) {
	var result *connect.QuickConnectSummary

	input := &connect.ListQuickConnectsInput{
		InstanceId: aws.String(instanceID),
		MaxResults: aws.Int64(ListQuickConnectsMaxResults),
	}

	err := conn.ListQuickConnectsPagesWithContext(ctx, input, func(page *connect.ListQuickConnectsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, cf := range page.QuickConnectSummaryList {
			if cf == nil {
				continue
			}

			if aws.StringValue(cf.Name) == name {
				result = cf
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
