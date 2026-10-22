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
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for list resource registration to the Provider. DO NOT EDIT.
// @SDKListResource("aws_ebs_volume")
func newEBSVolumeResourceAsListResource() inttypes.ListResourceForSDK {
	return framework.NewSDKListResource(resourceEBSVolume(), framework.SDKListSpec[awstypes.Volume, listEBSVolumeModel]{
		ConfigSchema: listEBSVolumeConfigSchema,
		Items: func(ctx context.Context, c *conns.AWSClient, query listEBSVolumeModel) (iter.Seq2[awstypes.Volume, error], diag.Diagnostics) {
			var input ec2.DescribeVolumesInput
			if diags := fwflex.Expand(ctx, query, &input); diags.HasError() {
				return nil, diags
			}

			return listEBSVolumes(ctx, c.EC2Client(ctx), &input), nil
		},
		ID: func(v awstypes.Volume, _ listEBSVolumeModel) string {
			return aws.ToString(v.VolumeId)
		},
		Flatten: resourceEBSVolumeFlatten,
		DisplayName: func(v awstypes.Volume) string {
			return listDisplayName(v.Tags, aws.ToString(v.VolumeId))
		},
	})
}

type listEBSVolumeModel struct {
	framework.WithRegionModel
	VolumeIDs fwtypes.ListValueOf[types.String] `tfsdk:"volume_ids"`
	Filters   customListFilters                 `tfsdk:"filter"`
}

func listEBSVolumeConfigSchema(ctx context.Context) listschema.Schema {
	return listschema.Schema{
		Attributes: map[string]listschema.Attribute{
			"volume_ids": listschema.ListAttribute{
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

func listEBSVolumes(ctx context.Context, conn *ec2.Client, input *ec2.DescribeVolumesInput) iter.Seq2[awstypes.Volume, error] {
	return func(yield func(awstypes.Volume, error) bool) {
		pages := ec2.NewDescribeVolumesPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.Volume{}, fmt.Errorf("listing EC2 EBS Volumes: %w", err))
				return
			}

			for _, item := range page.Volumes {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}
