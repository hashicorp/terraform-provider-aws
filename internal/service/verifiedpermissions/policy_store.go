// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package verifiedpermissions

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/verifiedpermissions"
	awstypes "github.com/aws/aws-sdk-go-v2/service/verifiedpermissions/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	sdkid "github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_verifiedpermissions_policy_store", name="Policy Store")
// @Tags(identifierAttribute="arn")
// @Testing(tagsTest=false)
func newPolicyStoreResource(context.Context) (resource.ResourceWithConfigure, error) {
	r := &policyStoreResource{}

	return r, nil
}

const (
	ResNamePolicyStore = "Policy Store"
)

type policyStoreResource struct {
	framework.ResourceWithModel[policyStoreResourceModel]
	framework.WithImportByID
}

func (r *policyStoreResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	s := schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrDeletionProtection: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.DeletionProtection](),
				Optional:   true,
				Computed:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
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
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
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

func (r *policyStoreResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data policyStoreResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().VerifiedPermissionsClient(ctx)

	var input verifiedpermissions.CreatePolicyStoreInput
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	clientToken := sdkid.UniqueId()
	input.ClientToken = aws.String(clientToken)
	input.Tags = getTagsIn(ctx)

	output, err := conn.CreatePolicyStore(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.VerifiedPermissions, create.ErrActionCreating, ResNamePolicyStore, clientToken, err),
			err.Error(),
		)
		return
	}

	// Set values for unknowns.
	data.ID = fwflex.StringToFramework(ctx, output.PolicyStoreId)

	policyStore, err := findPolicyStoreByID(ctx, conn, data.ID.ValueString())

	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.VerifiedPermissions, create.ErrActionReading, ResNamePolicyStore, data.PolicyStoreID.ValueString(), err),
			err.Error(),
		)
		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, policyStore, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *policyStoreResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data policyStoreResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().VerifiedPermissionsClient(ctx)

	output, err := findPolicyStoreByID(ctx, conn, data.ID.ValueString())

	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.VerifiedPermissions, create.ErrActionReading, ResNamePolicyStore, data.PolicyStoreID.ValueString(), err),
			err.Error(),
		)
		return
	}

	// Set attributes for import.
	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *policyStoreResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var old, new policyStoreResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().VerifiedPermissionsClient(ctx)

	if !new.DeletionProtection.Equal(old.DeletionProtection) || !new.Description.Equal(old.Description) || !new.ValidationSettings.Equal(old.ValidationSettings) {
		var input verifiedpermissions.UpdatePolicyStoreInput
		response.Diagnostics.Append(fwflex.Expand(ctx, new, &input)...)
		if response.Diagnostics.HasError() {
			return
		}

		_, err := conn.UpdatePolicyStore(ctx, &input)

		if err != nil {
			response.Diagnostics.AddError(
				create.ProblemStandardMessage(names.VerifiedPermissions, create.ErrActionUpdating, ResNamePolicyStore, old.PolicyStoreID.ValueString(), err),
				err.Error(),
			)
			return
		}
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *policyStoreResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data policyStoreResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().VerifiedPermissionsClient(ctx)

	tflog.Debug(ctx, "deleting Verified Permissions Policy Store", map[string]any{
		names.AttrID: data.ID.ValueString(),
	})

	input := verifiedpermissions.DeletePolicyStoreInput{
		PolicyStoreId: fwflex.StringFromFramework(ctx, data.ID),
	}
	_, err := conn.DeletePolicyStore(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.VerifiedPermissions, create.ErrActionDeleting, ResNamePolicyStore, data.PolicyStoreID.ValueString(), err),
			err.Error(),
		)
		return
	}
}

type policyStoreResourceModel struct {
	framework.WithRegionModel
	ARN                types.String                                        `tfsdk:"arn"`
	Description        types.String                                        `tfsdk:"description"`
	DeletionProtection fwtypes.StringEnum[awstypes.DeletionProtection]     `tfsdk:"deletion_protection"`
	ID                 types.String                                        `tfsdk:"id"`
	PolicyStoreID      types.String                                        `tfsdk:"policy_store_id"`
	Tags               tftags.Map                                          `tfsdk:"tags"`
	TagsAll            tftags.Map                                          `tfsdk:"tags_all"`
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
			LastError: err,
		}
	}
	if err != nil {
		return nil, err
	}

	if out == nil || out.Arn == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return out, nil
}
