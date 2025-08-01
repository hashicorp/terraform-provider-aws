// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fsx

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/fsx"
	awstypes "github.com/aws/aws-sdk-go-v2/service/fsx/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdkid "github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tfstringvalidator "github.com/hashicorp/terraform-provider-aws/internal/framework/validators/stringvalidator"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_fsx_s3_access_point_attachment", name="S3 Access Point Attachment")
func newS3AccessPointAttachmentResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &s3AccessPointAttachmentResource{}

	r.SetDefaultCreateTimeout(15 * time.Minute)
	r.SetDefaultDeleteTimeout(15 * time.Minute)

	return r, nil
}

type s3AccessPointAttachmentResource struct {
	framework.ResourceWithModel[s3AccessPointAttachmentResourceModel]
	framework.WithTimeouts
	framework.WithNoUpdate
}

func (r *s3AccessPointAttachmentResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrName: schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexache.MustCompile(`^[a-z0-9][a-z0-9-]{1,48}[a-z0-9]$`), "must between 3 and 50 lowercase letters, numbers, or hyphens"),
					tfstringvalidator.SuffixNoneOf("-ext-s3alias"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"s3_access_point_alias": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"s3_access_point_arn": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrType: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.S3AccessPointAttachmentType](),
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"openzfs_configuration": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[s3AccessPointOpenZFSConfigurationModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtMost(1),
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"volume_id": schema.StringAttribute{
							Required: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
					},
					Blocks: map[string]schema.Block{
						"file_system_identity": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[openZFSFileSystemIdentityModel](ctx),
							Validators: []validator.List{
								listvalidator.IsRequired(),
								listvalidator.SizeAtLeast(1),
								listvalidator.SizeAtMost(1),
							},
							PlanModifiers: []planmodifier.List{
								listplanmodifier.RequiresReplace(),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrType: schema.StringAttribute{
										CustomType: fwtypes.StringEnumType[awstypes.OpenZFSFileSystemUserType](),
										Required:   true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
										},
									},
								},
								Blocks: map[string]schema.Block{
									"posix_user": schema.ListNestedBlock{
										CustomType: fwtypes.NewListNestedObjectTypeOf[openZFSPosixFileSystemUserModel](ctx),
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
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
												},
												"secondary_gids": schema.ListAttribute{
													CustomType:  fwtypes.ListOfInt64Type,
													ElementType: types.Int64Type,
													Optional:    true,
													Validators: []validator.List{
														listvalidator.SizeAtMost(15),
													},
													PlanModifiers: []planmodifier.List{
														listplanmodifier.RequiresReplace(),
													},
												},
												"uid": schema.Int64Attribute{
													Required: true,
													PlanModifiers: []planmodifier.Int64{
														int64planmodifier.RequiresReplace(),
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			"s3_access_point": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[s3AccessPointModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrPolicy: schema.StringAttribute{
							CustomType: fwtypes.IAMPolicyType,
							Optional:   true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
					},
					Blocks: map[string]schema.Block{
						names.AttrVPCConfiguration: schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[s3AccessPointVpcConfigurationModel](ctx),
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							PlanModifiers: []planmodifier.List{
								listplanmodifier.RequiresReplace(),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrVPCID: schema.StringAttribute{
										Optional: true,
										PlanModifiers: []planmodifier.String{
											stringplanmodifier.RequiresReplace(),
										},
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

func (r *s3AccessPointAttachmentResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data s3AccessPointAttachmentResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().FSxClient(ctx)

	name := fwflex.StringValueFromFramework(ctx, data.Name)
	var input fsx.CreateAndAttachS3AccessPointInput
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.ClientRequestToken = aws.String(sdkid.UniqueId())

	_, err := conn.CreateAndAttachS3AccessPoint(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating FSx S3 Access Point Attachment (%s)", name), err.Error())

		return
	}

	output, err := waitS3AccessPointAttachmentCreated(ctx, conn, name, r.CreateTimeout(ctx, data.Timeouts))

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for FSx S3 Access Point Attachment (%s) create", name), err.Error())

		return
	}

	// Set values for unknowns.
	data.S3AccessPointAlias = fwflex.StringToFramework(ctx, output.S3AccessPoint.Alias)
	data.S3AccessPointARN = fwflex.StringToFramework(ctx, output.S3AccessPoint.ResourceARN)

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *s3AccessPointAttachmentResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data s3AccessPointAttachmentResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().FSxClient(ctx)

	name := fwflex.StringValueFromFramework(ctx, data.Name)
	output, err := findS3AccessPointAttachmentByName(ctx, conn, name)

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading FSx S3 Access Point Attachment (%s)", name), err.Error())

		return
	}

	// s3_access_point.policy is write-only.
	// Copy value from State.
	policy := fwtypes.IAMPolicyNull()
	s3AccessPoint, diags := data.S3AccessPoint.ToPtr(ctx)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}
	if s3AccessPoint != nil {
		policy = s3AccessPoint.Policy
	}

	// S3 access point alias and ARN are handled at the top level.
	data.S3AccessPointAlias = fwflex.StringToFramework(ctx, output.S3AccessPoint.Alias)
	data.S3AccessPointARN = fwflex.StringToFramework(ctx, output.S3AccessPoint.ResourceARN)
	if policy.IsNull() && output.S3AccessPoint.VpcConfiguration == nil {
		output.S3AccessPoint = nil
	}

	// Set attributes for import.
	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	// s3_access_point.policy is write-only.
	if !policy.IsNull() {
		s3AccessPoint, diags := data.S3AccessPoint.ToPtr(ctx)
		response.Diagnostics.Append(diags...)
		if response.Diagnostics.HasError() {
			return
		}
		s3AccessPoint.Policy = policy

		tfS3AccessPoint, diags := fwtypes.NewListNestedObjectValueOfPtr(ctx, s3AccessPoint)
		response.Diagnostics.Append(diags...)
		if response.Diagnostics.HasError() {
			return
		}
		data.S3AccessPoint = tfS3AccessPoint
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *s3AccessPointAttachmentResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data s3AccessPointAttachmentResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().FSxClient(ctx)

	name := fwflex.StringValueFromFramework(ctx, data.Name)
	input := fsx.DetachAndDeleteS3AccessPointInput{
		ClientRequestToken: aws.String(sdkid.UniqueId()),
		Name:               aws.String(name),
	}

	_, err := conn.DetachAndDeleteS3AccessPoint(ctx, &input)

	if errs.IsA[*awstypes.S3AccessPointAttachmentNotFound](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting FSx S3 Access Point Attachment (%s)", name), err.Error())

		return
	}

	if _, err := waitS3AccessPointAttachmentDeleted(ctx, conn, name, r.DeleteTimeout(ctx, data.Timeouts)); err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for FSx S3 Access Point Attachment (%s) delete", name), err.Error())

		return
	}
}

func (r *s3AccessPointAttachmentResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrName), request, response)
}

func findS3AccessPointAttachmentByName(ctx context.Context, conn *fsx.Client, name string) (*awstypes.S3AccessPointAttachment, error) {
	input := fsx.DescribeS3AccessPointAttachmentsInput{
		Names: []string{name},
	}
	output, err := findS3AccessPointAttachment(ctx, conn, &input, tfslices.PredicateTrue[*awstypes.S3AccessPointAttachment]())

	if err != nil {
		return nil, err
	}

	if output.S3AccessPoint == nil {
		return nil, tfresource.NewEmptyResultError(name)
	}

	return output, nil
}

func findS3AccessPointAttachment(ctx context.Context, conn *fsx.Client, input *fsx.DescribeS3AccessPointAttachmentsInput, filter tfslices.Predicate[*awstypes.S3AccessPointAttachment]) (*awstypes.S3AccessPointAttachment, error) {
	output, err := findS3AccessPointAttachments(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findS3AccessPointAttachments(ctx context.Context, conn *fsx.Client, input *fsx.DescribeS3AccessPointAttachmentsInput, filter tfslices.Predicate[*awstypes.S3AccessPointAttachment]) ([]awstypes.S3AccessPointAttachment, error) {
	var output []awstypes.S3AccessPointAttachment

	pages := fsx.NewDescribeS3AccessPointAttachmentsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.S3AccessPointAttachmentNotFound](err) {
			return nil, &retry.NotFoundError{
				LastError: err,
			}
		}

		if err != nil {
			return nil, err
		}

		for _, v := range page.S3AccessPointAttachments {
			if filter(&v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}

func statusS3AccessPointAttachment(conn *fsx.Client, name string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findS3AccessPointAttachmentByName(ctx, conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Lifecycle), nil
	}
}

func waitS3AccessPointAttachmentCreated(ctx context.Context, conn *fsx.Client, name string, timeout time.Duration) (*awstypes.S3AccessPointAttachment, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.S3AccessPointAttachmentLifecycleCreating),
		Target:  enum.Slice(awstypes.S3AccessPointAttachmentLifecycleAvailable),
		Refresh: statusS3AccessPointAttachment(conn, name),
		Timeout: timeout,
		Delay:   30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.S3AccessPointAttachment); ok {
		if v := output.LifecycleTransitionReason; v != nil {
			tfresource.SetLastError(err, errors.New(aws.ToString(v.Message)))
		}

		return output, err
	}

	return nil, err
}

func waitS3AccessPointAttachmentDeleted(ctx context.Context, conn *fsx.Client, name string, timeout time.Duration) (*awstypes.S3AccessPointAttachment, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.S3AccessPointAttachmentLifecycleDeleting),
		Target:  []string{},
		Refresh: statusS3AccessPointAttachment(conn, name),
		Timeout: timeout,
		Delay:   30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.S3AccessPointAttachment); ok {
		if v := output.LifecycleTransitionReason; v != nil {
			tfresource.SetLastError(err, errors.New(aws.ToString(v.Message)))
		}

		return output, err
	}

	return nil, err
}

type s3AccessPointAttachmentResourceModel struct {
	framework.WithRegionModel
	Name                 types.String                                                            `tfsdk:"name"`
	OpenZFSConfiguration fwtypes.ListNestedObjectValueOf[s3AccessPointOpenZFSConfigurationModel] `tfsdk:"openzfs_configuration"`
	S3AccessPoint        fwtypes.ListNestedObjectValueOf[s3AccessPointModel]                     `tfsdk:"s3_access_point"`
	S3AccessPointAlias   types.String                                                            `tfsdk:"s3_access_point_alias"`
	S3AccessPointARN     types.String                                                            `tfsdk:"s3_access_point_arn"`
	Timeouts             timeouts.Value                                                          `tfsdk:"timeouts"`
	Type                 fwtypes.StringEnum[awstypes.S3AccessPointAttachmentType]                `tfsdk:"type"`
}

type s3AccessPointOpenZFSConfigurationModel struct {
	FileSystemIdentity fwtypes.ListNestedObjectValueOf[openZFSFileSystemIdentityModel] `tfsdk:"file_system_identity"`
	VolumeID           types.String                                                    `tfsdk:"volume_id"`
}

type openZFSFileSystemIdentityModel struct {
	PosixUser fwtypes.ListNestedObjectValueOf[openZFSPosixFileSystemUserModel] `tfsdk:"posix_user"`
	Type      fwtypes.StringEnum[awstypes.OpenZFSFileSystemUserType]           `tfsdk:"type"`
}

type openZFSPosixFileSystemUserModel struct {
	GID           types.Int64         `tfsdk:"gid"`
	SecondaryGIDs fwtypes.ListOfInt64 `tfsdk:"secondary_gids"`
	UID           types.Int64         `tfsdk:"uid"`
}

type s3AccessPointModel struct {
	Policy           fwtypes.IAMPolicy                                                   `tfsdk:"policy"`
	VPCConfiguration fwtypes.ListNestedObjectValueOf[s3AccessPointVpcConfigurationModel] `tfsdk:"vpc_configuration"`
}

type s3AccessPointVpcConfigurationModel struct {
	VpcID types.String `tfsdk:"vpc_id"`
}
