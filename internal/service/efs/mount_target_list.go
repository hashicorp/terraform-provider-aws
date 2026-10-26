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
	"github.com/hashicorp/terraform-plugin-framework/list"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	tfiter "github.com/hashicorp/terraform-provider-aws/internal/iter"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKListResource("aws_efs_mount_target")
func newMountTargetResourceAsListResource() inttypes.ListResourceForSDK {
	l := mountTargetListResource{}
	l.SetResourceSchema(resourceMountTarget())
	return &l
}

var _ list.ListResource = &mountTargetListResource{}

type mountTargetListResource struct {
	framework.ListResourceWithSDKv2Resource
}

func (l *mountTargetListResource) ListResourceConfigSchema(ctx context.Context, request list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = listschema.Schema{
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

func (l *mountTargetListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().EFSClient(ctx)

	var query listMountTargetModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	var input efs.DescribeMountTargetsInput
	if diags := fwflex.Expand(ctx, query, &input); diags.HasError() {
		stream.Results = list.ListResultsStreamDiagnostics(diags)
		return
	}

	stream.Results = func(yield func(list.ListResult) bool) {
		for item, err := range listMountTargets(ctx, conn, &input) {
			if err != nil {
				result := fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			id := aws.ToString(item.MountTargetId)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrID), id)

			result := request.NewListResult(ctx)

			rd := l.ResourceData()
			rd.SetId(id)

			if request.IncludeResource {
				if err := resourceMountTargetFlatten(ctx, l.Meta(), &item, rd); err != nil {
					tflog.Error(ctx, "Reading EFS (Elastic File System) Mount Target", map[string]any{
						"error": err.Error(),
					})
					continue
				}
			}

			result.DisplayName = id

			l.SetResult(ctx, l.Meta(), request.IncludeResource, rd, &result)
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

type listMountTargetModel struct {
	framework.WithRegionModel
	AccessPointID types.String `tfsdk:"access_point_id"`
	FileSystemID  types.String `tfsdk:"file_system_id"`
}

func listMountTargets(ctx context.Context, conn *efs.Client, input *efs.DescribeMountTargetsInput, optFns ...func(*efs.Options)) iter.Seq2[awstypes.MountTargetDescription, error] {
	return tfiter.ConcatValuesWithError(listMountTargetPages(ctx, conn, input, optFns...))
}
