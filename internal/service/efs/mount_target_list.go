// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package efs

import (
	"context"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/efs"
	awstypes "github.com/aws/aws-sdk-go-v2/service/efs/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	tfiter "github.com/hashicorp/terraform-provider-aws/internal/iter"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKListResource("aws_efs_mount_target")
func newMountTargetResourceAsListResource() inttypes.ListResourceForSDK {
	return framework.NewSDKListResource(resourceMountTarget(), framework.SDKListSpec[awstypes.MountTargetDescription, listMountTargetModel]{
		ConfigSchema: listMountTargetConfigSchema,
		Items: func(ctx context.Context, c *conns.AWSClient, query listMountTargetModel) (iter.Seq2[awstypes.MountTargetDescription, error], diag.Diagnostics) {
			var input efs.DescribeMountTargetsInput
			if diags := fwflex.Expand(ctx, query, &input); diags.HasError() {
				return nil, diags
			}

			return listMountTargets(ctx, c.EFSClient(ctx), &input), nil
		},
		ID: func(mt awstypes.MountTargetDescription, _ listMountTargetModel) string {
			return aws.ToString(mt.MountTargetId)
		},
		Flatten: resourceMountTargetFlatten,
		// Flatten describes each mount target's security groups, so a listing
		// that only wants identity should not pay for that call. Identity here
		// is the resource ID, which the shared loop already sets.
		FlattenWhenRequested: true,
	})
}

func listMountTargetConfigSchema(ctx context.Context) listschema.Schema {
	return listschema.Schema{
		Attributes: map[string]listschema.Attribute{
			"access_point_id": listschema.StringAttribute{
				Optional:    true,
				Description: "ID of the access point whose mount targets you want to list.",
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(
						path.MatchRoot("access_point_id"),
						path.MatchRoot(names.AttrFileSystemID),
					),
				},
			},
			names.AttrFileSystemID: listschema.StringAttribute{
				Optional:    true,
				Description: "ID of the file system whose mount targets you want to list.",
			},
		},
	}
}

type listMountTargetModel struct {
	framework.WithRegionModel
	AccessPointID types.String `tfsdk:"access_point_id"`
	FileSystemID  types.String `tfsdk:"file_system_id"`
}

func listMountTargets(ctx context.Context, conn *efs.Client, input *efs.DescribeMountTargetsInput, optFns ...func(*efs.Options)) iter.Seq2[awstypes.MountTargetDescription, error] {
	return tfiter.ConcatValuesWithError(listMountTargetPages(ctx, conn, input, optFns...))
}
