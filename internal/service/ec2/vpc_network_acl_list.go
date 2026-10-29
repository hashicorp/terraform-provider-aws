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
	"github.com/hashicorp/terraform-plugin-framework/diag"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tfiter "github.com/hashicorp/terraform-provider-aws/internal/iter"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKListResource("aws_network_acl")
func newNetworkACLResourceAsListResource() inttypes.ListResourceForSDK {
	return framework.NewSDKListResource(resourceNetworkACL(), framework.SDKListSpec[awstypes.NetworkAcl, networkACLListResourceModel]{
		ConfigSchema: listNetworkACLConfigSchema,
		Items: func(ctx context.Context, c *conns.AWSClient, query networkACLListResourceModel) (iter.Seq2[awstypes.NetworkAcl, error], diag.Diagnostics) {
			var input ec2.DescribeNetworkAclsInput
			if diags := fwflex.Expand(ctx, query, &input); diags.HasError() {
				return nil, diags
			}

			// The default network ACL belongs to aws_default_network_acl, and
			// DescribeNetworkAcls cannot filter it out server side.
			return tfiter.FilterValues(listNetworkACLs(ctx, c.EC2Client(ctx), &input), func(n awstypes.NetworkAcl) bool {
				return !aws.ToBool(n.IsDefault)
			}), nil
		},
		ID: func(n awstypes.NetworkAcl, _ networkACLListResourceModel) string {
			return aws.ToString(n.NetworkAclId)
		},
		Flatten: resourceNetworkACLFlatten,
	})
}

type networkACLListResourceModel struct {
	framework.WithRegionModel
	NetworkACLIDs fwtypes.ListValueOf[types.String] `tfsdk:"network_acl_ids"`
	Filters       customListFilters                 `tfsdk:"filter"`
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

func listNetworkACLConfigSchema(ctx context.Context) listschema.Schema {
	return listschema.Schema{
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
