// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package connect

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/connect"
	awstypes "github.com/aws/aws-sdk-go-v2/service/connect/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_connect_contact_flow_module", name="Contact Flow Module")
// @Tags
func dataSourceContactFlowModule() *schema.Resource {
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

func dataSourceContactFlowModuleRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConnectClient(ctx)

	instanceID := d.Get(names.AttrInstanceID).(string)
	input := &connect.DescribeContactFlowModuleInput{
		InstanceId: aws.String(instanceID),
	}

	if v, ok := d.GetOk("contact_flow_module_id"); ok {
		input.ContactFlowModuleId = aws.String(v.(string))
	} else if v, ok := d.GetOk(names.AttrName); ok {
		name := v.(string)
		contactFlowModuleSummary, err := findContactFlowModuleSummaryByTwoPartKey(ctx, conn, instanceID, name)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading Connect Contact Flow Module (%s) summary: %s", name, err)
		}

		input.ContactFlowModuleId = contactFlowModuleSummary.Id
	}

	contactFlowModule, err := findContactFlowModule(ctx, conn, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Connect Contact Flow Module: %s", err)
	}

	contactFlowModuleID := aws.ToString(contactFlowModule.Id)
	d.SetId(contactFlowModuleCreateResourceID(instanceID, contactFlowModuleID))
	d.Set(names.AttrARN, contactFlowModule.Arn)
	d.Set("contact_flow_module_id", contactFlowModuleID)
	d.Set(names.AttrContent, contactFlowModule.Content)
	d.Set(names.AttrDescription, contactFlowModule.Description)
	d.Set(names.AttrName, contactFlowModule.Name)
	d.Set(names.AttrState, contactFlowModule.State)
	d.Set(names.AttrStatus, contactFlowModule.Status)

	setTagsOut(ctx, contactFlowModule.Tags)

	return diags
}

func findContactFlowModuleSummaryByTwoPartKey(ctx context.Context, conn *connect.Client, instanceID, name string) (*awstypes.ContactFlowModuleSummary, error) {
	const maxResults = 60
	input := &connect.ListContactFlowModulesInput{
		InstanceId: aws.String(instanceID),
		MaxResults: aws.Int32(maxResults),
	}

	return findContactFlowModuleSummary(ctx, conn, input, func(v *awstypes.ContactFlowModuleSummary) bool {
		return aws.ToString(v.Name) == name
	})
}

func findContactFlowModuleSummary(ctx context.Context, conn *connect.Client, input *connect.ListContactFlowModulesInput, filter tfslices.Predicate[*awstypes.ContactFlowModuleSummary]) (*awstypes.ContactFlowModuleSummary, error) {
	output, err := findContactFlowModuleSummaries(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findContactFlowModuleSummaries(ctx context.Context, conn *connect.Client, input *connect.ListContactFlowModulesInput, filter tfslices.Predicate[*awstypes.ContactFlowModuleSummary]) ([]awstypes.ContactFlowModuleSummary, error) {
	var output []awstypes.ContactFlowModuleSummary

	pages := connect.NewListContactFlowModulesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		for _, v := range page.ContactFlowModulesSummaryList {
			if filter(&v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}
