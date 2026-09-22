// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
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
// @SDKListResource("aws_default_security_group")
func newDefaultSecurityGroupResourceAsListResource() inttypes.ListResourceForSDK {
	l := listResourceDefaultSecurityGroup{}
	l.SetResourceSchema(resourceDefaultSecurityGroup())
	return &l
}

var _ list.ListResource = &listResourceDefaultSecurityGroup{}
var _ list.ListResourceWithRawV5Schemas = &listResourceDefaultSecurityGroup{}

type listResourceDefaultSecurityGroup struct {
	framework.ListResourceWithSDKv2Resource
}

type listDefaultSecurityGroupModel struct {
	framework.WithRegionModel
	Filters customListFilters `tfsdk:"filter"`
}

func (l *listResourceDefaultSecurityGroup) ListResourceConfigSchema(ctx context.Context, request list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = listschema.Schema{
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

func (l *listResourceDefaultSecurityGroup) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	awsClient := l.Meta()
	conn := awsClient.EC2Client(ctx)

	var query listDefaultSecurityGroupModel
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

	// Restrict results to default security groups.
	input.Filters = append(input.Filters, newAttributeFilterList(
		map[string]string{
			"group-name": defaultSecurityGroupName,
		},
	)...)

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

			rd := l.ResourceData()
			rd.SetId(groupID)

			diags := resourceSecurityGroupFlatten(ctx, awsClient, rd, &item)
			if diags.HasError() {
				tflog.Error(ctx, "Reading resource", map[string]any{
					"diags": sdkdiag.DiagnosticsString(diags),
				})
				continue
			}

			if v, ok := tags["Name"]; ok {
				result.DisplayName = fmt.Sprintf("%s (%s)", v.ValueString(), groupID)
			} else {
				result.DisplayName = aws.ToString(item.GroupName)
			}

			l.SetResult(ctx, awsClient, request.IncludeResource, rd, &result)
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
