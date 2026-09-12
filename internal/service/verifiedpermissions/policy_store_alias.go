// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package verifiedpermissions

import (
	"context"
	"fmt"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/verifiedpermissions"
	awstypes "github.com/aws/aws-sdk-go-v2/service/verifiedpermissions/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	ResNamePolicyStoreAlias = "Policy Store Alias"

	policyStoreAliasPrefix             = "policy-store-alias/"
	policyStoreAliasPropagationTimeout = 2 * time.Minute
)

// @FrameworkResource("aws_verifiedpermissions_policy_store_alias", name="Policy Store Alias")
// @IdentityAttribute("alias_name")
// @Testing(importStateIdAttribute="alias_name")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/verifiedpermissions;verifiedpermissions.GetPolicyStoreAliasOutput")
// @Testing(hasNoPreExistingResource=true)
func newPolicyStoreAliasResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &policyStoreAliasResource{}, nil
}

type policyStoreAliasResource struct {
	framework.ResourceWithModel[policyStoreAliasResourceModel]
	framework.WithImportByIdentity
}

func (r *policyStoreAliasResource) Schema(
	ctx context.Context,
	request resource.SchemaRequest,
	response *resource.SchemaResponse,
) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),

			"alias_name": schema.StringAttribute{
				Description: "Name of the policy store alias. Must begin with `policy-store-alias/`.",
				Required:    true,
				Validators:  policyStoreAliasNameValidators(),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},

			"created_at": schema.StringAttribute{
				Description: "Date and time when the policy store alias was created.",
				CustomType:  timetypes.RFC3339Type{},
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			"policy_store_id": schema.StringAttribute{
				Description: "ID of the policy store associated with the alias.",
				Required:    true,
				Validators:  policyStoreIDValidators(),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},

			names.AttrState: schema.StringAttribute{
				Description: "Current state of the policy store alias.",
				CustomType:  fwtypes.StringEnumType[awstypes.AliasState](),
				Computed:    true,
			},
		},
	}
}

func (r *policyStoreAliasResource) Create(
	ctx context.Context,
	request resource.CreateRequest,
	response *resource.CreateResponse,
) {
	var data policyStoreAliasResourceModel

	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().VerifiedPermissionsClient(ctx)

	aliasName := data.AliasName.ValueString()
	policyStoreID := data.PolicyStoreID.ValueString()

	input := verifiedpermissions.CreatePolicyStoreAliasInput{
		AliasName:     data.AliasName.ValueStringPointer(),
		PolicyStoreId: data.PolicyStoreID.ValueStringPointer(),
	}

	_, err := conn.CreatePolicyStoreAlias(ctx, &input)
	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(
				names.VerifiedPermissions,
				create.ErrActionCreating,
				ResNamePolicyStoreAlias,
				aliasName,
				err,
			),
			err.Error(),
		)
		return
	}

	output, err := tfresource.RetryWhenNotFound(
		ctx,
		policyStoreAliasPropagationTimeout,
		func(ctx context.Context) (*verifiedpermissions.GetPolicyStoreAliasOutput, error) {
			return findPolicyStoreAliasByNameAndPolicyStoreID(
				ctx,
				conn,
				aliasName,
				policyStoreID,
			)
		},
	)
	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(
				names.VerifiedPermissions,
				create.ErrActionCreating,
				ResNamePolicyStoreAlias,
				aliasName,
				err,
			),
			err.Error(),
		)
		return
	}

	response.Diagnostics.Append(flattenPolicyStoreAlias(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *policyStoreAliasResource) Read(
	ctx context.Context,
	request resource.ReadRequest,
	response *resource.ReadResponse,
) {
	var data policyStoreAliasResourceModel

	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().VerifiedPermissionsClient(ctx)

	aliasName := data.AliasName.ValueString()

	output, err := findPolicyStoreAliasByName(ctx, conn, aliasName)

	if retry.NotFound(err) {
		response.Diagnostics.Append(
			fwdiag.NewResourceNotFoundWarningDiagnostic(err),
		)
		response.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(
				names.VerifiedPermissions,
				create.ErrActionReading,
				ResNamePolicyStoreAlias,
				aliasName,
				err,
			),
			err.Error(),
		)
		return
	}

	response.Diagnostics.Append(flattenPolicyStoreAlias(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *policyStoreAliasResource) Delete(
	ctx context.Context,
	request resource.DeleteRequest,
	response *resource.DeleteResponse,
) {
	var data policyStoreAliasResourceModel

	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().VerifiedPermissionsClient(ctx)

	aliasName := data.AliasName.ValueString()

	input := verifiedpermissions.DeletePolicyStoreAliasInput{
		AliasName:    data.AliasName.ValueStringPointer(),
		DeletionMode: awstypes.DeletionModeHardDelete,
	}

	_, err := conn.DeletePolicyStoreAlias(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(
				names.VerifiedPermissions,
				create.ErrActionDeleting,
				ResNamePolicyStoreAlias,
				aliasName,
				err,
			),
			err.Error(),
		)
		return
	}

	_, err = tfresource.RetryUntilNotFound(
		ctx,
		policyStoreAliasPropagationTimeout,
		func(ctx context.Context) (any, error) {
			return findPolicyStoreAliasByName(ctx, conn, aliasName)
		},
	)
	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(
				names.VerifiedPermissions,
				create.ErrActionDeleting,
				ResNamePolicyStoreAlias,
				aliasName,
				err,
			),
			err.Error(),
		)
	}
}

func findPolicyStoreAliasByName(
	ctx context.Context,
	conn *verifiedpermissions.Client,
	aliasName string,
) (*verifiedpermissions.GetPolicyStoreAliasOutput, error) {
	input := verifiedpermissions.GetPolicyStoreAliasInput{
		AliasName: aws.String(aliasName),
	}

	output, err := conn.GetPolicyStoreAlias(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.AliasName == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	if output.State == awstypes.AliasStatePendingDeletion {
		return nil, &retry.NotFoundError{
			Message: fmt.Sprintf(
				"Verified Permissions Policy Store Alias (%s) is pending deletion",
				aliasName,
			),
		}
	}

	return output, nil
}

func findPolicyStoreAliasByNameAndPolicyStoreID(
	ctx context.Context,
	conn *verifiedpermissions.Client,
	aliasName string,
	expectedPolicyStoreID string,
) (*verifiedpermissions.GetPolicyStoreAliasOutput, error) {
	output, err := findPolicyStoreAliasByName(ctx, conn, aliasName)
	if err != nil {
		return nil, err
	}

	actualPolicyStoreID := aws.ToString(output.PolicyStoreId)
	if actualPolicyStoreID != expectedPolicyStoreID {
		return nil, &retry.NotFoundError{
			Message: fmt.Sprintf(
				"Verified Permissions Policy Store Alias (%s) is associated with Policy Store (%s), expected (%s)",
				aliasName,
				actualPolicyStoreID,
				expectedPolicyStoreID,
			),
		}
	}

	return output, nil
}

func flattenPolicyStoreAlias(
	ctx context.Context,
	source any,
	data *policyStoreAliasResourceModel,
) diag.Diagnostics {
	return fwflex.Flatten(ctx, source, data)
}

func policyStoreAliasNameValidators() []validator.String {
	return []validator.String{
		stringvalidator.LengthBetween(19, 150),
		stringvalidator.RegexMatches(
			regexache.MustCompile(`^policy-store-alias/[A-Za-z0-9/_-]*$`),
			fmt.Sprintf(
				"value must begin with %q and contain only letters, numbers, hyphens, underscores, and forward slashes",
				policyStoreAliasPrefix,
			),
		),
	}
}

func policyStoreIDValidators() []validator.String {
	return []validator.String{
		stringvalidator.LengthBetween(1, 200),
		stringvalidator.RegexMatches(
			regexache.MustCompile(`^[A-Za-z0-9/_-]+$`),
			"value must contain only letters, numbers, hyphens, underscores, and forward slashes",
		),
	}
}

type policyStoreAliasResourceModel struct {
	framework.WithRegionModel

	AliasARN      types.String                            `tfsdk:"arn"`
	AliasName     types.String                            `tfsdk:"alias_name"`
	CreatedAt     timetypes.RFC3339                       `tfsdk:"created_at"`
	PolicyStoreID types.String                            `tfsdk:"policy_store_id"`
	State         fwtypes.StringEnum[awstypes.AliasState] `tfsdk:"state"`
}
