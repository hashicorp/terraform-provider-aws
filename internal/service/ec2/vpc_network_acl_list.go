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
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKListResource("aws_network_acl")
func newNetworkACLResourceAsListResource() inttypes.ListResourceForSDK {
	l := networkACLListResource{}
	l.SetResourceSchema(resourceNetworkACL())

	return &l
}

var _ list.ListResource = &networkACLListResource{}

type networkACLListResource struct {
	framework.ListResourceWithSDKv2Resource
}

type networkACLListResourceModel struct {
	framework.WithRegionModel
	NetworkACLIDs fwtypes.ListValueOf[types.String] `tfsdk:"network_acl_ids"`
	Filters       customListFilters                 `tfsdk:"filter"`
}

func (l *networkACLListResource) ListResourceConfigSchema(ctx context.Context, _ list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = listschema.Schema{
		Attributes: map[string]listschema.Attribute{
			"network_acl_ids": listschema.ListAttribute{
				CustomType:  fwtypes.ListOfStringType,
				ElementType: types.StringType,
				Optional:    true,
			},
		},
		Blocks: map[string]listschema.Block{
			names.AttrFilter: customListFiltersBlock(ctx),
		},
	}
}

func (l *networkACLListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	awsClient := l.Meta()
	conn := awsClient.EC2Client(ctx)

	var query networkACLListResourceModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	var input ec2.DescribeNetworkAclsInput
	if diags := fwflex.Expand(ctx, query, &input); diags.HasError() {
		stream.Results = list.ListResultsStreamDiagnostics(diags)
		return
	}

	tflog.Info(ctx, "Listing resources")

	stream.Results = func(yield func(list.ListResult) bool) {
		for networkACL, err := range listNetworkACLs(ctx, conn, &input) {
			if err != nil {
				yield(fwdiag.NewListResultErrorDiagnostic(err))
				return
			}

			if aws.ToBool(networkACL.IsDefault) {
				continue
			}

			id := aws.ToString(networkACL.NetworkAclId)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrID), id)

			result := request.NewListResult(ctx)
			rd := l.ResourceData()
			rd.SetId(id)

			if request.IncludeResource {
				if diags := resourceNetworkACLFlatten(ctx, awsClient, rd, &networkACL); diags.HasError() {
					tflog.Error(ctx, "Reading EC2 Network ACL", map[string]any{
						"diags": diags,
					})
					continue
				}
			}

			result.DisplayName = id

			l.SetResult(ctx, awsClient, request.IncludeResource, rd, &result)
			if result.Diagnostics.HasError() {
				yield(result)
				return
			}

			if !yield(result) {
				return
			}
		}
	}
}

func listNetworkACLs(ctx context.Context, conn *ec2.Client, input *ec2.DescribeNetworkAclsInput) iter.Seq2[awstypes.NetworkAcl, error] {
	return func(yield func(awstypes.NetworkAcl, error) bool) {
		pages := ec2.NewDescribeNetworkAclsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.NetworkAcl{}, fmt.Errorf("listing EC2 Network ACLs: %w", err))
				return
			}

			for _, networkACL := range page.NetworkAcls {
				if !yield(networkACL, nil) {
					return
				}
			}
		}
	}
}
