// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package verifiedpermissions

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/verifiedpermissions"
	awstypes "github.com/aws/aws-sdk-go-v2/service/verifiedpermissions/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource(name="Policy Store")
func newResourcePolicyStore(context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourcePolicyStore{}

	return r, nil
}

const (
	ResNamePolicyStore = "Policy Store"
)

type resourcePolicyStore struct {
	framework.ResourceWithConfigure
}

func (r *resourcePolicyStore) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_verifiedpermissions_policy_store"
}

func (r *resourcePolicyStore) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	s := schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
			},
			names.AttrID: framework.IDAttribute(),
			"policy_store_id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"validation_settings": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[validationSettings](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrMode: schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.ValidationMode](),
							Required:   true,
						},
					},
				},
			},
		},
	}

	response.Schema = s
}

func (r *resourcePolicyStore) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	conn := r.Meta().VerifiedPermissionsClient(ctx)
	var plan resourcePolicyStoreData

	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)

	if response.Diagnostics.HasError() {
		return
	}

	input := &verifiedpermissions.CreatePolicyStoreInput{}
	response.Diagnostics.Append(flex.Expand(ctx, plan, input)...)

	if response.Diagnostics.HasError() {
		return
	}

	clientToken := id.UniqueId()
	input.ClientToken = aws.String(clientToken)

	output, err := conn.CreatePolicyStore(ctx, input)

	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.VerifiedPermissions, create.ErrActionCreating, ResNamePolicyStore, clientToken, err),
			err.Error(),
		)
		return
	}

	state := plan
	state.ID = flex.StringToFramework(ctx, output.PolicyStoreId)

	response.Diagnostics.Append(flex.Flatten(ctx, output, &state)...)

	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *resourcePolicyStore) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	conn := r.Meta().VerifiedPermissionsClient(ctx)
	var state resourcePolicyStoreData

	response.Diagnostics.Append(request.State.Get(ctx, &state)...)

	if response.Diagnostics.HasError() {
		return
	}

	output, err := findPolicyStoreByID(ctx, conn, state.ID.ValueString())

	if tfresource.NotFound(err) {
		response.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.VerifiedPermissions, create.ErrActionReading, ResNamePolicyStore, state.PolicyStoreID.ValueString(), err),
			err.Error(),
		)
		return
	}

	response.Diagnostics.Append(flex.Flatten(ctx, output, &state)...)

	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *resourcePolicyStore) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	conn := r.Meta().VerifiedPermissionsClient(ctx)
	var state, plan resourcePolicyStoreData

	response.Diagnostics.Append(request.State.Get(ctx, &state)...)

	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)

	if response.Diagnostics.HasError() {
		return
	}

	if !plan.Description.Equal(state.Description) || !plan.ValidationSettings.Equal(state.ValidationSettings) {
		input := &verifiedpermissions.UpdatePolicyStoreInput{}
		response.Diagnostics.Append(flex.Expand(ctx, plan, input)...)

		if response.Diagnostics.HasError() {
			return
		}

		output, err := conn.UpdatePolicyStore(ctx, input)

		if err != nil {
			response.Diagnostics.AddError(
				create.ProblemStandardMessage(names.VerifiedPermissions, create.ErrActionUpdating, ResNamePolicyStore, state.PolicyStoreID.ValueString(), err),
				err.Error(),
			)
			return
		}

		response.Diagnostics.Append(flex.Flatten(ctx, output, &plan)...)
	}

	response.Diagnostics.Append(response.State.Set(ctx, &plan)...)
}

func (r *resourcePolicyStore) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	conn := r.Meta().VerifiedPermissionsClient(ctx)
	var state resourcePolicyStoreData

	response.Diagnostics.Append(request.State.Get(ctx, &state)...)

	if response.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "deleting Verified Permissions Policy Store", map[string]interface{}{
		names.AttrID: state.ID.ValueString(),
	})

	input := &verifiedpermissions.DeletePolicyStoreInput{
		PolicyStoreId: flex.StringFromFramework(ctx, state.ID),
	}

	_, err := conn.DeletePolicyStore(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.VerifiedPermissions, create.ErrActionDeleting, ResNamePolicyStore, state.PolicyStoreID.ValueString(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourcePolicyStore) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrID), request, response)
}

type resourcePolicyStoreData struct {
	ARN                types.String                                        `tfsdk:"arn"`
	Description        types.String                                        `tfsdk:"description"`
	ID                 types.String                                        `tfsdk:"id"`
	PolicyStoreID      types.String                                        `tfsdk:"policy_store_id"`
	ValidationSettings fwtypes.ListNestedObjectValueOf[validationSettings] `tfsdk:"validation_settings"`
}

type validationSettings struct {
	Mode fwtypes.StringEnum[awstypes.ValidationMode] `tfsdk:"mode"`
}

func findPolicyStoreByID(ctx context.Context, conn *verifiedpermissions.Client, id string) (*verifiedpermissions.GetPolicyStoreOutput, error) {
	in := &verifiedpermissions.GetPolicyStoreInput{
		PolicyStoreId: aws.String(id),
	}

	out, err := conn.GetPolicyStore(ctx, in)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}
	if err != nil {
		return nil, err
	}

	if out == nil || out.Arn == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}
