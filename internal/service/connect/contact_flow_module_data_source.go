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

func DataSourceContactFlowModule() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceContactFlowModuleRead,
		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"contact_flow_module_id": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{"contact_flow_module_id", "name"},
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
				ExactlyOneOf: []string{"name", "contact_flow_module_id"},
			},
			"state": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceContactFlowModuleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ConnectConn
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	instanceID := d.Get("instance_id").(string)

	input := &connect.DescribeContactFlowModuleInput{
		InstanceId: aws.String(instanceID),
	}

	if v, ok := d.GetOk("contact_flow_module_id"); ok {
		input.ContactFlowModuleId = aws.String(v.(string))
	} else if v, ok := d.GetOk("name"); ok {
		name := v.(string)
		contactFlowModuleSummary, err := dataSourceGetContactFlowModuleSummaryByName(ctx, conn, instanceID, name)

		if err != nil {
			return diag.FromErr(fmt.Errorf("error finding Connect Contact Flow Module Summary by name (%s): %w", name, err))
		}

		if contactFlowModuleSummary == nil {
			return diag.FromErr(fmt.Errorf("error finding Connect Contact Flow Module Summary by name (%s): not found", name))
		}

		input.ContactFlowModuleId = contactFlowModuleSummary.Id
	}

	resp, err := conn.DescribeContactFlowModuleWithContext(ctx, input)

	if err != nil {
		return diag.FromErr(fmt.Errorf("error getting Connect Contact Flow Module: %w", err))
	}

	if resp == nil || resp.ContactFlowModule == nil {
		return diag.FromErr(fmt.Errorf("error getting Connect Contact Flow Module: empty response"))
	}

	contactFlowModule := resp.ContactFlowModule

	d.Set("arn", contactFlowModule.Arn)
	d.Set("contact_flow_module_id", contactFlowModule.Id)
	d.Set("content", contactFlowModule.Content)
	d.Set("description", contactFlowModule.Description)
	d.Set("name", contactFlowModule.Name)
	d.Set("state", contactFlowModule.State)
	d.Set("status", contactFlowModule.Status)

	if err := d.Set("tags", KeyValueTags(contactFlowModule.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return diag.FromErr(fmt.Errorf("error setting tags: %s", err))
	}

	d.SetId(fmt.Sprintf("%s:%s", instanceID, aws.StringValue(contactFlowModule.Id)))

	return nil
}

func dataSourceGetContactFlowModuleSummaryByName(ctx context.Context, conn *connect.Connect, instanceID, name string) (*connect.ContactFlowModuleSummary, error) {
	var result *connect.ContactFlowModuleSummary

	input := &connect.ListContactFlowModulesInput{
		InstanceId: aws.String(instanceID),
		MaxResults: aws.Int64(ListContactFlowModulesMaxResults),
	}

	err := conn.ListContactFlowModulesPagesWithContext(ctx, input, func(page *connect.ListContactFlowModulesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, cf := range page.ContactFlowModulesSummaryList {
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
