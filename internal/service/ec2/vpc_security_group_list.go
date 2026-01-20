// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-framework/list"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for list resource registration to the Provider. DO NOT EDIT.
// @SDKListResource("aws_security_group")
func newSecurityGroupResourceAsListResource() inttypes.ListResourceForSDK {
	l := listResourceSecurityGroup{}
	l.SetResourceSchema(resourceSecurityGroup())
	return &l
}

var _ list.ListResource = &listResourceSecurityGroup{}
var _ list.ListResourceWithRawV5Schemas = &listResourceSecurityGroup{}

type listResourceSecurityGroup struct {
	framework.ListResourceWithSDKv2Resource
}

type listSecurityGroupModel struct {
	framework.WithRegionModel
	GroupIDs fwtypes.ListOfString `tfsdk:"group_ids"`
	Filters  customListFilters    `tfsdk:"filter"`
}

func (l *listResourceSecurityGroup) ListResourceConfigSchema(ctx context.Context, request list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = listschema.Schema{
		Attributes: map[string]listschema.Attribute{
			"group_ids": listschema.ListAttribute{
				CustomType: fwtypes.ListOfStringType,
				Optional:   true,
			},
		},
		Blocks: map[string]listschema.Block{
			names.AttrFilter: listschema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[customListFilterModel](ctx),
				NestedObject: listschema.NestedBlockObject{
					Attributes: map[string]listschema.Attribute{
						names.AttrName: listschema.StringAttribute{
							Required: true,
						},
						names.AttrValues: listschema.ListAttribute{
							CustomType: fwtypes.ListOfStringType,
							Required:   true,
						},
					},
				},
			},
		},
	}
}

func (l *listResourceSecurityGroup) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	awsClient := l.Meta()
	conn := awsClient.EC2Client(ctx)

	var query listSecurityGroupModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	tflog.Info(ctx, "Listing resources")

	var input ec2.DescribeSecurityGroupsInput
	if diags := fwflex.Expand(ctx, query, &input); diags.HasError() {
		stream.Results = list.ListResultsStreamDiagnostics(diags)
		return
	}

	stream.Results = func(yield func(list.ListResult) bool) {

		for item, err := range listSecurityGroups(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			groupID := aws.ToString(item.GroupId)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrID), groupID)

			result := request.NewListResult(ctx)
			tags := keyValueTags(ctx, item.Tags)
			setTagsOut(ctx, item.Tags)

			rd := l.ResourceData()
			rd.SetId(groupID)

			tflog.Info(ctx, "Reading resource")
			diags := resourceSecurityGroupRead(ctx, rd, awsClient)
			if diags.HasError() {
				tflog.Error(ctx, "Reading resource", map[string]any{
					names.AttrID: groupID,
					"diags":      sdkdiag.DiagnosticsString(diags),
				})
				continue
			}
			if rd.Id() == "" {
				// Resource is logically deleted
				continue
			}

			if v, ok := tags["Name"]; ok {
				result.DisplayName = fmt.Sprintf("%s (%s)", v.ValueString(), groupID)
			} else {
				result.DisplayName = aws.ToString(item.GroupName)
			}

			l.SetResult(ctx, awsClient, request.IncludeResource, &result, rd)
			if result.Diagnostics.HasError() {
				tflog.Error(ctx, "Setting result", map[string]any{
					names.AttrID: groupID,
					"diags":      result.Diagnostics,
				})
				continue
			}

			if !yield(result) {
				return
			}
		}
	}
}

func listSecurityGroups(ctx context.Context, conn *ec2.Client, input *ec2.DescribeSecurityGroupsInput) iter.Seq2[awstypes.SecurityGroup, error] {
	return func(yield func(awstypes.SecurityGroup, error) bool) {
		pages := ec2.NewDescribeSecurityGroupsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.SecurityGroup{}, fmt.Errorf("listing EC2 Security Groups: %w", err))
				return
			}

			for _, item := range page.SecurityGroups {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}
