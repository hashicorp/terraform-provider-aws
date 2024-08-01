// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package connect

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/connect"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_connect_contact_flow_module")
func DataSourceContactFlowModule() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceContactFlowModuleRead,
		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"contact_flow_module_id": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{"contact_flow_module_id", names.AttrName},
			},
			names.AttrContent: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrInstanceID: {
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrName: {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{names.AttrName, "contact_flow_module_id"},
			},
			names.AttrState: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrStatus: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceContactFlowModuleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).ConnectConn(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	instanceID := d.Get(names.AttrInstanceID).(string)

	input := &connect.DescribeContactFlowModuleInput{
		InstanceId: aws.String(instanceID),
	}

	if v, ok := d.GetOk("contact_flow_module_id"); ok {
		input.ContactFlowModuleId = aws.String(v.(string))
	} else if v, ok := d.GetOk(names.AttrName); ok {
		name := v.(string)
		contactFlowModuleSummary, err := dataSourceGetContactFlowModuleSummaryByName(ctx, conn, instanceID, name)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "finding Connect Contact Flow Module Summary by name (%s): %s", name, err)
		}

		if contactFlowModuleSummary == nil {
			return sdkdiag.AppendErrorf(diags, "finding Connect Contact Flow Module Summary by name (%s): not found", name)
		}

		input.ContactFlowModuleId = contactFlowModuleSummary.Id
	}

	resp, err := conn.DescribeContactFlowModuleWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "getting Connect Contact Flow Module: %s", err)
	}

	if resp == nil || resp.ContactFlowModule == nil {
		return sdkdiag.AppendErrorf(diags, "getting Connect Contact Flow Module: empty response")
	}

	contactFlowModule := resp.ContactFlowModule

	d.Set(names.AttrARN, contactFlowModule.Arn)
	d.Set("contact_flow_module_id", contactFlowModule.Id)
	d.Set(names.AttrContent, contactFlowModule.Content)
	d.Set(names.AttrDescription, contactFlowModule.Description)
	d.Set(names.AttrName, contactFlowModule.Name)
	d.Set(names.AttrState, contactFlowModule.State)
	d.Set(names.AttrStatus, contactFlowModule.Status)

	if err := d.Set(names.AttrTags, KeyValueTags(ctx, contactFlowModule.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	d.SetId(fmt.Sprintf("%s:%s", instanceID, aws.StringValue(contactFlowModule.Id)))

	return diags
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
