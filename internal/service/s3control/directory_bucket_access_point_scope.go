// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package s3control

import (
	"context"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3control"
	awstypes "github.com/aws/aws-sdk-go-v2/service/s3control/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	fwvalidators "github.com/hashicorp/terraform-provider-aws/internal/framework/validators"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_s3control_directory_bucket_access_point_scope", name="Directory Bucket Access Point Scope")
func newDirectoryBucketAccessPointScopeResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &directoryBucketAccessPointScopeResource{}
	return r, nil
}

type directoryBucketAccessPointScopeResource struct {
	framework.ResourceWithModel[directoryBucketAccessPointScopeModel]
}

const (
	ResNameDirectoryBucketAccessPointScope = "Directory Bucket Access Point Scope"
)

var AccessPointForDirectoryBucketNameRegex = regexache.MustCompile(`^(?:[0-9a-z.-]+)--(?:[0-9a-za-z]+(?:-[0-9a-za-z]+)+)--xa-s3$`)

func (r *directoryBucketAccessPointScopeResource) Schema(ctx context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	s := schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrAccountID: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					fwvalidators.AWSAccountID(),
				},
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(AccessPointForDirectoryBucketNameRegex,
						"must be in the format [access_point_name]--[azid]--xa-s3"),
				},
			},
		},
		Blocks: map[string]schema.Block{
			names.AttrScope: schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[scopeModel](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrPermissions: schema.ListAttribute{
							CustomType: fwtypes.ListOfStringEnumType[awstypes.ScopePermission](),
							Optional:   true,
						},
						"prefixes": schema.ListAttribute{
							CustomType: fwtypes.ListOfStringType,
							Optional:   true,
						},
					},
				},
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtMost(1),
				},
			},
		},
	}

	response.Schema = s
}

func (r *directoryBucketAccessPointScopeResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan directoryBucketAccessPointScopeModel
	conn := r.Meta().S3ControlClient(ctx)

	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	input := s3control.PutAccessPointScopeInput{}
	response.Diagnostics.Append(fwflex.Expand(ctx, plan, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	id, _ := plan.setID()
	_, err := conn.PutAccessPointScope(ctx, &input)
	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.S3Control, create.ErrActionCreating, ResNameDirectoryBucketAccessPointScope, id, err),
			err.Error(),
		)
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &plan)...)
}

func (r *directoryBucketAccessPointScopeResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data directoryBucketAccessPointScopeModel
	conn := r.Meta().S3ControlClient(ctx)

	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	id, _ := data.setID()
	output, err := findDirectoryAccessPointScopeByTwoPartKey(ctx, conn, data.AccountID.ValueString(), data.Name.ValueString())

	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.S3Control, create.ErrActionReading, ResNameDirectoryBucketAccessPointScope, id, err),
			err.Error(),
		)
		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *directoryBucketAccessPointScopeResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state, plan directoryBucketAccessPointScopeModel
	conn := r.Meta().S3ControlClient(ctx)

	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	if !plan.Scope.Equal(state.Scope) {
		input := s3control.PutAccessPointScopeInput{}
		response.Diagnostics.Append(fwflex.Expand(ctx, plan, &input)...)
		if response.Diagnostics.HasError() {
			return
		}

		id, _ := plan.setID()
		_, err := conn.PutAccessPointScope(ctx, &input)
		if err != nil {
			response.Diagnostics.AddError(
				create.ProblemStandardMessage(names.S3Control, create.ErrActionUpdating, ResNameDirectoryBucketAccessPointScope, id, err),
				err.Error(),
			)
			return
		}
	}

	response.Diagnostics.Append(response.State.Set(ctx, &plan)...)
}

func (r *directoryBucketAccessPointScopeResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data directoryBucketAccessPointScopeModel
	conn := r.Meta().S3ControlClient(ctx)

	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "deleting S3Control Directory Bucket Access Point Scope", map[string]any{
		names.AttrName:      data.Name.ValueString(),
		names.AttrAccountID: data.AccountID.ValueString(),
	})

	input := s3control.DeleteAccessPointScopeInput{
		AccountId: data.AccountID.ValueStringPointer(),
		Name:      data.Name.ValueStringPointer(),
	}

	_, err := conn.DeleteAccessPointScope(ctx, &input)
	if tfawserr.ErrCodeEquals(err, errCodeNoSuchAccessPoint) {
		return
	}

	id, _ := data.setID()
	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.S3Control, create.ErrActionDeleting, ResNameDirectoryBucketAccessPointScope, id, err),
			err.Error(),
		)
		return
	}
}

func (r *directoryBucketAccessPointScopeResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	parts, err := intflex.ExpandResourceId(request.ID, directoryBucketAccessPointScopeIdPartsCount, false)
	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.S3Control, create.ErrActionImporting, ResNameDirectoryBucketAccessPointScope, request.ID, err),
			err.Error(),
		)
		return
	}

	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root(names.AttrName), parts[0])...)
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root(names.AttrAccountID), parts[1])...)
}

func findDirectoryAccessPointScopeByTwoPartKey(ctx context.Context, conn *s3control.Client, accountID, name string) (*s3control.GetAccessPointScopeOutput, error) {
	input := s3control.GetAccessPointScopeInput{
		AccountId: aws.String(accountID),
		Name:      aws.String(name),
	}

	output, err := conn.GetAccessPointScope(ctx, &input)
	if tfawserr.ErrCodeEquals(err, errCodeNoSuchAccessPoint) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Scope == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output, nil
}

const (
	directoryBucketAccessPointScopeIdPartsCount = 2
)

func (d *directoryBucketAccessPointScopeModel) setID() (string, error) {
	parts := []string{
		d.Name.ValueString(),
		d.AccountID.ValueString(),
	}

	return intflex.FlattenResourceId(parts, directoryBucketAccessPointScopeIdPartsCount, false)
}

type directoryBucketAccessPointScopeModel struct {
	framework.WithRegionModel
	AccountID types.String                                `tfsdk:"account_id"`
	Name      types.String                                `tfsdk:"name"`
	Scope     fwtypes.ListNestedObjectValueOf[scopeModel] `tfsdk:"scope"`
}

type scopeModel struct {
	Permissions fwtypes.ListValueOf[fwtypes.StringEnum[awstypes.ScopePermission]] `tfsdk:"permissions"`
	Prefixes    fwtypes.ListOfString                                              `tfsdk:"prefixes"`
}
