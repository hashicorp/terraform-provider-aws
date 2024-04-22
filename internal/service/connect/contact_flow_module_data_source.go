// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package connect

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/connect"
	awstypes "github.com/aws/aws-sdk-go-v2/service/connect/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

// @SDKDataSource("aws_connect_contact_flow_module")
func DataSourceContactFlowModule() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceContactFlowModuleRead,
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
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).ConnectClient(ctx)
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
			return sdkdiag.AppendErrorf(diags, "finding Connect Contact Flow Module Summary by name (%s): %s", name, err)
		}

		input.ContactFlowModuleId = contactFlowModuleSummary.Id
	}

	resp, err := conn.DescribeContactFlowModule(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "getting Connect Contact Flow Module: %s", err)
	}

	if resp == nil || resp.ContactFlowModule == nil {
		return sdkdiag.AppendErrorf(diags, "getting Connect Contact Flow Module: empty response")
	}

	contactFlowModule := resp.ContactFlowModule

	d.Set("arn", contactFlowModule.Arn)
	d.Set("contact_flow_module_id", contactFlowModule.Id)
	d.Set("content", contactFlowModule.Content)
	d.Set("description", contactFlowModule.Description)
	d.Set("name", contactFlowModule.Name)
	d.Set("state", contactFlowModule.State)
	d.Set("status", contactFlowModule.Status)

	if err := d.Set("tags", KeyValueTags(ctx, contactFlowModule.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	d.SetId(fmt.Sprintf("%s:%s", instanceID, aws.ToString(contactFlowModule.Id)))

	return diags
}

func dataSourceGetContactFlowModuleSummaryByName(ctx context.Context, conn *connect.Client, instanceID, name string) (*awstypes.ContactFlowModuleSummary, error) {
	var result *awstypes.ContactFlowModuleSummary

	input := &connect.ListContactFlowModulesInput{
		InstanceId: aws.String(instanceID),
		MaxResults: aws.Int32(ListContactFlowModulesMaxResults),
	}

	pages := connect.NewListContactFlowModulesPaginator(conn, input)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return result, nil
		}

		for _, cf := range page.ContactFlowModulesSummaryList {
			cf := cf
			if aws.ToString(cf.Name) == name {
				result = &cf
			}
		}
	}

	return result, nil
}
