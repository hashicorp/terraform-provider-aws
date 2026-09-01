// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package bedrockagentcore

import (
	"context"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_bedrockagentcore_token_vault_cmk", name="Token Vault CMK")
func newTokenVaultCMKResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &tokenVaultCMKResource{}
	return r, nil
}

type tokenVaultCMKResource struct {
	framework.ResourceWithModel[tokenVaultCMKResourceModel]
	framework.WithNoOpDelete
}

func (r *tokenVaultCMKResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"token_vault_id": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString("default"),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"kms_configuration": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[kmsConfigurationModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtLeast(1),
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"key_type": schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.KeyType](),
							Required:   true,
						},
						names.AttrKMSKeyARN: schema.StringAttribute{
							CustomType: fwtypes.ARNType,
							Optional:   true,
						},
					},
				},
			},
		},
	}
}

func (r *tokenVaultCMKResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data tokenVaultCMKResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentCoreClient(ctx)

	tokenVaultID := fwflex.StringValueFromFramework(ctx, data.TokenVaultID)
	var input bedrockagentcorecontrol.SetTokenVaultCMKInput
	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, data, &input))
	if response.Diagnostics.HasError() {
		return
	}

	_, err := conn.SetTokenVaultCMK(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, tokenVaultID)
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, data))
}

func (r *tokenVaultCMKResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data tokenVaultCMKResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentCoreClient(ctx)

	tokenVaultID := fwflex.StringValueFromFramework(ctx, data.TokenVaultID)
	out, err := findTokenVaultByID(ctx, conn, tokenVaultID)
	if retry.NotFound(err) {
		smerr.AddOne(ctx, &response.Diagnostics, fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, tokenVaultID)
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Flatten(ctx, out, &data))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data))
}

func (r *tokenVaultCMKResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var new, old tokenVaultCMKResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Plan.Get(ctx, &new))
	smerr.AddEnrich(ctx, &response.Diagnostics, request.State.Get(ctx, &old))
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockAgentCoreClient(ctx)

	diff, d := fwflex.Diff(ctx, new, old)
	smerr.AddEnrich(ctx, &response.Diagnostics, d)
	if response.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		tokenVaultID := fwflex.StringValueFromFramework(ctx, new.TokenVaultID)
		var input bedrockagentcorecontrol.SetTokenVaultCMKInput
		smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, new, &input))
		if response.Diagnostics.HasError() {
			return
		}

		_, err := conn.SetTokenVaultCMK(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, tokenVaultID)
			return
		}
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &new))
}

func (r *tokenVaultCMKResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("token_vault_id"), request, response)
}

func findTokenVaultByID(ctx context.Context, conn *bedrockagentcorecontrol.Client, id string) (*bedrockagentcorecontrol.GetTokenVaultOutput, error) {
	input := bedrockagentcorecontrol.GetTokenVaultInput{
		TokenVaultId: aws.String(id),
	}

	return findTokenVault(ctx, conn, &input)
}

func findTokenVault(ctx context.Context, conn *bedrockagentcorecontrol.Client, input *bedrockagentcorecontrol.GetTokenVaultInput) (*bedrockagentcorecontrol.GetTokenVaultOutput, error) {
	out, err := conn.GetTokenVault(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, smarterr.NewError(&retry.NotFoundError{
			LastError: err,
		})
	}

	if err != nil {
		return nil, smarterr.NewError(err)
	}

	if out == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	return out, nil
}

type tokenVaultCMKResourceModel struct {
	framework.WithRegionModel
	KMSConfiguration fwtypes.ListNestedObjectValueOf[kmsConfigurationModel] `tfsdk:"kms_configuration"`
	TokenVaultID     types.String                                           `tfsdk:"token_vault_id"`
}

type kmsConfigurationModel struct {
	KeyType   fwtypes.StringEnum[awstypes.KeyType] `tfsdk:"key_type"`
	KMSKeyARN fwtypes.ARN                          `tfsdk:"kms_key_arn"`
}
