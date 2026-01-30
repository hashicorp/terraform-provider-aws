// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package verifiedpermissions

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/verifiedpermissions"
	awstypes "github.com/aws/aws-sdk-go-v2/service/verifiedpermissions/types"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_verifiedpermissions_schema", name="Schema")
func newSchemaResource(context.Context) (resource.ResourceWithConfigure, error) {
	return &schemaResource{}, nil
}

const (
	ResNamePolicyStoreSchema = "Schema"
)

type schemaResource struct {
	framework.ResourceWithModel[schemaResourceModel]
	framework.WithImportByID
}

func (r *schemaResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Version: 1,
		Attributes: map[string]schema.Attribute{
			names.AttrID: framework.IDAttribute(),
			"namespaces": schema.SetAttribute{
				CustomType:  fwtypes.SetOfStringType,
				ElementType: types.StringType,
				Computed:    true,
			},
			"policy_store_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"definition": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[definitionData](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrValue: schema.StringAttribute{
							CustomType: jsontypes.NormalizedType{},
							Required:   true,
						},
					},
				},
			},
		},
	}
}

func (r *schemaResource) UpgradeState(ctx context.Context) map[int64]resource.StateUpgrader {
	schemaV0 := schemaSchemaV0()

	return map[int64]resource.StateUpgrader{
		0: {
			PriorSchema:   &schemaV0,
			StateUpgrader: upgradeSchemaStateFromV0,
		},
	}
}

func (r *schemaResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	conn := r.Meta().VerifiedPermissionsClient(ctx)
	var plan schemaResourceModel

	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	input := verifiedpermissions.PutSchemaInput{}
	response.Diagnostics.Append(flex.Expand(ctx, plan, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	out, err := conn.PutSchema(ctx, &input)
	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.VerifiedPermissions, create.ErrActionCreating, ResNamePolicyStoreSchema, plan.PolicyStoreID.ValueString(), err),
			err.Error(),
		)
		return
	}

	response.Diagnostics.Append(flex.Flatten(ctx, out, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}
	plan.ID = flex.StringToFramework(ctx, out.PolicyStoreId)

	response.Diagnostics.Append(response.State.Set(ctx, &plan)...)
}

func (r *schemaResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	conn := r.Meta().VerifiedPermissionsClient(ctx)
	var state schemaResourceModel

	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	out, err := findSchemaByPolicyStoreID(ctx, conn, state.ID.ValueString())

	if retry.NotFound(err) {
		response.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.VerifiedPermissions, create.ErrActionReading, ResNamePolicyStoreSchema, state.PolicyStoreID.ValueString(), err),
			err.Error(),
		)
		return
	}

	response.Diagnostics.Append(flex.Flatten(ctx, out, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Misalignment between input and output data structures requires manual value assignment
	definition, d := flattenDefinition(ctx, out)
	response.Diagnostics.Append(d...)
	if response.Diagnostics.HasError() {
		return
	}
	state.Definition = definition

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *schemaResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	conn := r.Meta().VerifiedPermissionsClient(ctx)
	var state, plan schemaResourceModel

	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	if !plan.Definition.Equal(state.Definition) {
		input := verifiedpermissions.PutSchemaInput{}

		response.Diagnostics.Append(flex.Expand(ctx, plan, &input)...)
		if response.Diagnostics.HasError() {
			return
		}

		_, err := conn.PutSchema(ctx, &input)
		if err != nil {
			response.Diagnostics.AddError(
				create.ProblemStandardMessage(names.VerifiedPermissions, create.ErrActionUpdating, ResNamePolicyStoreSchema, state.PolicyStoreID.ValueString(), err),
				err.Error(),
			)
			return
		}

		out, err := findSchemaByPolicyStoreID(ctx, conn, state.ID.ValueString())
		if err != nil {
			response.Diagnostics.AddError(
				create.ProblemStandardMessage(names.VerifiedPermissions, create.ErrActionUpdating, ResNamePolicyStoreSchema, state.PolicyStoreID.ValueString(), err),
				err.Error(),
			)
			return
		}

		response.Diagnostics.Append(flex.Flatten(ctx, out, &plan)...)
		if response.Diagnostics.HasError() {
			return
		}

		// Misalignment between input and output data structures requires manual value assignment
		definition, d := flattenDefinition(ctx, out)
		response.Diagnostics.Append(d...)
		if response.Diagnostics.HasError() {
			return
		}
		plan.Definition = definition
	}

	response.Diagnostics.Append(response.State.Set(ctx, &plan)...)
}

func (r *schemaResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	conn := r.Meta().VerifiedPermissionsClient(ctx)
	var state schemaResourceModel

	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "deleting Verified Permissions Policy Store Schema", map[string]any{
		names.AttrID: state.ID.ValueString(),
	})

	input := verifiedpermissions.PutSchemaInput{
		PolicyStoreId: flex.StringFromFramework(ctx, state.ID),
		Definition: &awstypes.SchemaDefinitionMemberCedarJson{
			Value: "{}",
		},
	}

	_, err := conn.PutSchema(ctx, &input)
	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.VerifiedPermissions, create.ErrActionDeleting, ResNamePolicyStoreSchema, state.PolicyStoreID.ValueString(), err),
			err.Error(),
		)
		return
	}
}

type schemaResourceModel struct {
	framework.WithRegionModel
	ID            types.String                                    `tfsdk:"id"`
	Definition    fwtypes.ListNestedObjectValueOf[definitionData] `tfsdk:"definition"`
	Namespaces    fwtypes.SetOfString                             `tfsdk:"namespaces"`
	PolicyStoreID types.String                                    `tfsdk:"policy_store_id"`
}

type definitionData struct {
	Value jsontypes.Normalized `tfsdk:"value"`
}

func findSchemaByPolicyStoreID(ctx context.Context, conn *verifiedpermissions.Client, id string) (*verifiedpermissions.GetSchemaOutput, error) {
	in := &verifiedpermissions.GetSchemaInput{
		PolicyStoreId: aws.String(id),
	}

	out, err := conn.GetSchema(ctx, in)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}
	if err != nil {
		return nil, err
	}

	if out == nil || out.Schema == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return out, nil
}

func (m definitionData) Expand(ctx context.Context) (result any, diags diag.Diagnostics) {
	return &awstypes.SchemaDefinitionMemberCedarJson{
		Value: m.Value.ValueString(),
	}, diags
}

func flattenDefinition(ctx context.Context, input *verifiedpermissions.GetSchemaOutput) (fwtypes.ListNestedObjectValueOf[definitionData], diag.Diagnostics) {
	var diags diag.Diagnostics
	if input == nil {
		return fwtypes.NewListNestedObjectValueOfNull[definitionData](ctx), diags
	}

	data := []definitionData{
		{
			Value: jsontypes.NewNormalizedPointerValue(input.Schema),
		},
	}

	return fwtypes.NewListNestedObjectValueOfValueSlice(ctx, data)
}
