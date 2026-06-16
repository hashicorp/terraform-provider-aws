// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package s3files

import (
	"context"
	"errors"
	"time"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3files"
	awstypes "github.com/aws/aws-sdk-go-v2/service/s3files/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_s3files_access_point", name="Access Point")
// @Tags(identifierAttribute="arn")
// @IdentityAttribute("id")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/s3files;s3files.GetAccessPointOutput")
// @Testing(existsTakesT=true, destroyTakesT=true)
// @Testing(hasNoPreExistingResource=true)
func newAccessPointResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &accessPointResource{}

	r.SetDefaultCreateTimeout(10 * time.Minute)
	r.SetDefaultDeleteTimeout(10 * time.Minute)

	return r, nil
}

type accessPointResource struct {
	framework.ResourceWithModel[accessPointResourceModel]
	framework.WithImportByIdentity
	framework.WithTimeouts
}

func (r *accessPointResource) Schema(ctx context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrFileSystemID: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Description: "File system ID",
			},
			names.AttrID: framework.IDAttribute(),
			names.AttrName: schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "Access point name",
			},
			names.AttrOwnerID: schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Description: "AWS account ID of the owner",
			},
			names.AttrStatus: schema.StringAttribute{
				Computed:    true,
				Description: "Access point status",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"posix_user": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[posixUserModel](ctx),
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"gid": schema.Int64Attribute{
							Required: true,
							PlanModifiers: []planmodifier.Int64{
								int64planmodifier.RequiresReplace(),
							},
							Description: "POSIX group ID",
						},
						"secondary_gids": schema.SetAttribute{
							ElementType: types.Int64Type,
							Optional:    true,
							PlanModifiers: []planmodifier.Set{
								setplanmodifier.RequiresReplace(),
							},
							Description: "Secondary POSIX group IDs",
						},
						"uid": schema.Int64Attribute{
							Required: true,
							PlanModifiers: []planmodifier.Int64{
								int64planmodifier.RequiresReplace(),
							},
							Description: "POSIX user ID",
						},
					},
				},
			},
			"root_directory": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[rootDirectoryModel](ctx),
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrPath: schema.StringAttribute{
							Optional: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
							Description: "Root directory path",
						},
					},
					Blocks: map[string]schema.Block{
						"creation_permissions": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[creationPermissionsModel](ctx),
							PlanModifiers: []planmodifier.List{
								listplanmodifier.RequiresReplace(),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"owner_gid": schema.Int64Attribute{
										Required: true,
										PlanModifiers: []planmodifier.Int64{
											int64planmodifier.RequiresReplace(),
										},
										Description: "Owner group ID",
									},
									"owner_uid": schema.Int64Attribute{
										Required: true,
										PlanModifiers: []planmodifier.Int64{
											int64planmodifier.RequiresReplace(),
										},
										Description: "Owner user ID",
									},
									names.AttrPermissions: schema.StringAttribute{
										Required: true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
										},
										Description: "POSIX permissions",
									},
								},
							},
						},
					},
				},
			},
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Delete: true,
			}),
		},
	}
}

func (r *accessPointResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data accessPointResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().S3FilesClient(ctx)

	input := s3files.CreateAccessPointInput{}
	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, data, &input))
	if response.Diagnostics.HasError() {
		return
	}

	input.ClientToken = aws.String(create.UniqueId(ctx))
	input.Tags = getTagsIn(ctx)

	output, err := conn.CreateAccessPoint(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err)
		return
	}

	data.ID = fwflex.StringToFramework(ctx, output.AccessPointId)

	accessPoint, err := waitAccessPointCreated(ctx, conn, data.ID.ValueString(), r.CreateTimeout(ctx, data.Timeouts))
	if err != nil {
		response.State.SetAttribute(ctx, path.Root(names.AttrID), data.ID)
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, data.ID.ValueString())
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Flatten(ctx, accessPoint, &data, fwflex.WithFieldNamePrefix("AccessPoint")))
	if response.Diagnostics.HasError() {
		return
	}

	// Preserve plan value for root_directory if not set
	var plan accessPointResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if plan.RootDirectory.IsNull() {
		data.RootDirectory = plan.RootDirectory
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, data))
}

func (r *accessPointResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data accessPointResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	// Remember if root_directory was null in state
	rootDirWasNull := data.RootDirectory.IsNull()

	conn := r.Meta().S3FilesClient(ctx)

	output, err := findAccessPointByID(ctx, conn, data.ID.ValueString())
	if retry.NotFound(err) {
		smerr.AddOne(ctx, &response.Diagnostics, fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, data.ID.ValueString())
		return
	}

	flattenAccessPointResource(ctx, output, &data, rootDirWasNull, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, data))
}

func flattenAccessPointResource(ctx context.Context, output *s3files.GetAccessPointOutput, data *accessPointResourceModel, rootDirWasNull bool, diags *diag.Diagnostics) { // nosemgrep:ci.semgrep.framework.manual-flattener-functions
	smerr.AddEnrich(ctx, diags, fwflex.Flatten(ctx, output, data, fwflex.WithFieldNamePrefix("AccessPoint")))

	// Preserve null root_directory if it was null in state
	if rootDirWasNull {
		data.RootDirectory = fwtypes.NewListNestedObjectValueOfNull[rootDirectoryModel](ctx)
	}
}

func (r *accessPointResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var data accessPointResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	// Only tags can be updated; transparent tagging handles the update
	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, data))
}

func (r *accessPointResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data accessPointResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().S3FilesClient(ctx)

	input := s3files.DeleteAccessPointInput{
		AccessPointId: data.ID.ValueStringPointer(),
	}

	_, err := conn.DeleteAccessPoint(ctx, &input)
	if err != nil {
		var nfe *awstypes.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return
		}
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, data.ID.ValueString())
		return
	}

	_, err = waitAccessPointDeleted(ctx, conn, data.ID.ValueString(), r.DeleteTimeout(ctx, data.Timeouts))
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, data.ID.ValueString())
	}
}

type accessPointResourceModel struct {
	framework.WithRegionModel
	ARN           types.String                                        `tfsdk:"arn"`
	FileSystemID  types.String                                        `tfsdk:"file_system_id"`
	ID            types.String                                        `tfsdk:"id"`
	Name          types.String                                        `tfsdk:"name"`
	OwnerID       types.String                                        `tfsdk:"owner_id"`
	PosixUser     fwtypes.ListNestedObjectValueOf[posixUserModel]     `tfsdk:"posix_user"`
	RootDirectory fwtypes.ListNestedObjectValueOf[rootDirectoryModel] `tfsdk:"root_directory"`
	Status        types.String                                        `tfsdk:"status"`
	Tags          tftags.Map                                          `tfsdk:"tags"`
	TagsAll       tftags.Map                                          `tfsdk:"tags_all"`
	Timeouts      timeouts.Value                                      `tfsdk:"timeouts"`
}

type posixUserModel struct {
	Gid           types.Int64                     `tfsdk:"gid"`
	SecondaryGids fwtypes.SetValueOf[types.Int64] `tfsdk:"secondary_gids"`
	Uid           types.Int64                     `tfsdk:"uid"`
}

type rootDirectoryModel struct {
	CreationPermissions fwtypes.ListNestedObjectValueOf[creationPermissionsModel] `tfsdk:"creation_permissions"`
	Path                types.String                                              `tfsdk:"path"`
}

type creationPermissionsModel struct {
	OwnerGid    types.Int64  `tfsdk:"owner_gid"`
	OwnerUid    types.Int64  `tfsdk:"owner_uid"`
	Permissions types.String `tfsdk:"permissions"`
}

func findAccessPointByID(ctx context.Context, conn *s3files.Client, id string) (*s3files.GetAccessPointOutput, error) {
	input := s3files.GetAccessPointInput{
		AccessPointId: &id,
	}

	output, err := conn.GetAccessPoint(ctx, &input)
	if err != nil {
		var nfe *awstypes.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return nil, &retry.NotFoundError{
				LastError: err,
			}
		}
		return nil, smarterr.NewError(err)
	}

	if output == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	return output, nil
}

func waitAccessPointCreated(ctx context.Context, conn *s3files.Client, id string, timeout time.Duration) (*s3files.GetAccessPointOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.LifeCycleStateCreating),
		Target:  enum.Slice(awstypes.LifeCycleStateAvailable, awstypes.LifeCycleStateError),
		Refresh: statusAccessPoint(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*s3files.GetAccessPointOutput); ok {
		return output, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitAccessPointDeleted(ctx context.Context, conn *s3files.Client, id string, timeout time.Duration) (*s3files.GetAccessPointOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.LifeCycleStateAvailable, awstypes.LifeCycleStateDeleting),
		Target:  []string{},
		Refresh: statusAccessPoint(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*s3files.GetAccessPointOutput); ok {
		return output, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func statusAccessPoint(_ context.Context, conn *s3files.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findAccessPointByID(ctx, conn, id)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", smarterr.NewError(err)
		}

		if output != nil && output.Status == awstypes.LifeCycleStateError {
			return nil, string(output.Status), smarterr.NewError(errors.New("in error state"))
		}

		return output, string(output.Status), nil
	}
}
