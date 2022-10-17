package connect

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/connect"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

func DataSourceContactFlow() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceContactFlowRead,
		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"contact_flow_id": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{"contact_flow_id", "name"},
			},
			"content": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"instance_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{"name", "contact_flow_id"},
			},
			"tags": tftags.TagsSchemaComputed(),
			"type": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func dataSourceContactFlowRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ConnectConn
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	instanceID := d.Get("instance_id").(string)

	input := &connect.DescribeContactFlowInput{
		InstanceId: aws.String(instanceID),
	}

	if v, ok := d.GetOk("contact_flow_id"); ok {
		input.ContactFlowId = aws.String(v.(string))
	} else if v, ok := d.GetOk("name"); ok {
		name := v.(string)
		contactFlowSummary, err := dataSourceGetContactFlowSummaryByName(ctx, conn, instanceID, name)

		if err != nil {
			return diag.FromErr(fmt.Errorf("error finding Connect Contact Flow Summary by name (%s): %w", name, err))
		}

		if contactFlowSummary == nil {
			return diag.FromErr(fmt.Errorf("error finding Connect Contact Flow Summary by name (%s): not found", name))
		}

		input.ContactFlowId = contactFlowSummary.Id
	}

	resp, err := conn.DescribeContactFlow(input)

	if err != nil {
		return diag.FromErr(fmt.Errorf("error getting Connect Contact Flow: %w", err))
	}

	if resp == nil || resp.ContactFlow == nil {
		return diag.FromErr(fmt.Errorf("error getting Connect Contact Flow: empty response"))
	}

	contactFlow := resp.ContactFlow

	d.Set("arn", contactFlow.Arn)
	d.Set("instance_id", instanceID)
	d.Set("contact_flow_id", contactFlow.Id)
	d.Set("name", contactFlow.Name)
	d.Set("description", contactFlow.Description)
	d.Set("content", contactFlow.Content)
	d.Set("type", contactFlow.Type)

	if err := d.Set("tags", KeyValueTags(contactFlow.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return diag.FromErr(fmt.Errorf("error setting tags: %s", err))
	}

	d.SetId(fmt.Sprintf("%s:%s", instanceID, aws.StringValue(contactFlow.Id)))

	return nil
}

func dataSourceGetContactFlowSummaryByName(ctx context.Context, conn *connect.Connect, instanceID, name string) (*connect.ContactFlowSummary, error) {
	var result *connect.ContactFlowSummary

	input := &connect.ListContactFlowsInput{
		InstanceId: aws.String(instanceID),
		MaxResults: aws.Int64(ListContactFlowsMaxResults),
	}

	err := conn.ListContactFlowsPagesWithContext(ctx, input, func(page *connect.ListContactFlowsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, cf := range page.ContactFlowSummaryList {
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
