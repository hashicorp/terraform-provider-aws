// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package cognitoidp

import (
	"context"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Values that GetUserPoolSessionConfiguration returns for a user pool that has never been configured.
// Delete restores them because the API has no operation that removes the configuration.
const (
	defaultUserPoolSessionTracking = awstypes.UserSessionTrackingTypeDisabled
	defaultMaxSessionsPerUser      = 20
)

// @FrameworkResource("aws_cognito_user_pool_session_configuration", name="User Pool Session Configuration")
// @IdentityAttribute("user_pool_id")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider;cognitoidentityprovider.GetUserPoolSessionConfigurationOutput")
// @Testing(importStateIdAttribute="user_pool_id")
// @Testing(hasNoPreExistingResource=true)
func newUserPoolSessionConfigurationResource(context.Context) (resource.ResourceWithConfigure, error) {
	return &userPoolSessionConfigurationResource{}, nil
}

type userPoolSessionConfigurationResource struct {
	framework.ResourceWithModel[userPoolSessionConfigurationResourceModel]
	framework.WithImportByIdentity
}

func (r *userPoolSessionConfigurationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrUserPoolID: schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"user_session_tracking": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.UserSessionTrackingType](),
				Required:   true,
			},
		},
		Blocks: map[string]schema.Block{
			"token_session": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[tokenSessionModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"max_per_user": schema.Int32Attribute{
							Required: true,
							Validators: []validator.Int32{
								int32validator.Between(1, 100),
							},
						},
					},
				},
			},
		},
	}
}

func (r *userPoolSessionConfigurationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().CognitoIDPClient(ctx)

	var plan userPoolSessionConfigurationResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	var input cognitoidentityprovider.SetUserPoolSessionConfigurationInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &input))
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := conn.SetUserPoolSessionConfiguration(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.UserPoolID.String())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *userPoolSessionConfigurationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().CognitoIDPClient(ctx)

	var state userPoolSessionConfigurationResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findUserPoolSessionConfigurationByUserPoolID(ctx, conn, state.UserPoolID.ValueString())
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.UserPoolID.String())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, out, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *userPoolSessionConfigurationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().CognitoIDPClient(ctx)

	var plan, state userPoolSessionConfigurationResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	diff, d := flex.Diff(ctx, plan, state)
	smerr.AddEnrich(ctx, &resp.Diagnostics, d)
	if resp.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		// SetUserPoolSessionConfiguration replaces the whole configuration, so send every argument.
		var input cognitoidentityprovider.SetUserPoolSessionConfigurationInput
		smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &input))
		if resp.Diagnostics.HasError() {
			return
		}

		_, err := conn.SetUserPoolSessionConfiguration(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.UserPoolID.String())
			return
		}
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *userPoolSessionConfigurationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().CognitoIDPClient(ctx)

	var state userPoolSessionConfigurationResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	input := cognitoidentityprovider.SetUserPoolSessionConfigurationInput{
		UserPoolId:          state.UserPoolID.ValueStringPointer(),
		UserSessionTracking: defaultUserPoolSessionTracking,
		TokenSession: &awstypes.UserPoolTokenSessionConfigType{
			MaxPerUser: aws.Int32(defaultMaxSessionsPerUser),
		},
	}

	_, err := conn.SetUserPoolSessionConfiguration(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.UserPoolID.String())
		return
	}
}

func findUserPoolSessionConfigurationByUserPoolID(ctx context.Context, conn *cognitoidentityprovider.Client, userPoolID string) (*cognitoidentityprovider.GetUserPoolSessionConfigurationOutput, error) {
	input := cognitoidentityprovider.GetUserPoolSessionConfigurationInput{
		UserPoolId: aws.String(userPoolID),
	}

	out, err := conn.GetUserPoolSessionConfiguration(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, smarterr.NewError(&retry.NotFoundError{
			LastError: err,
		})
	}
	if err != nil {
		return nil, smarterr.NewError(err)
	}

	if out == nil || out.TokenSession == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	return out, nil
}

type userPoolSessionConfigurationResourceModel struct {
	framework.WithRegionModel
	TokenSession        fwtypes.ListNestedObjectValueOf[tokenSessionModel]   `tfsdk:"token_session"`
	UserPoolID          types.String                                         `tfsdk:"user_pool_id"`
	UserSessionTracking fwtypes.StringEnum[awstypes.UserSessionTrackingType] `tfsdk:"user_session_tracking"`
}

type tokenSessionModel struct {
	MaxPerUser types.Int32 `tfsdk:"max_per_user"`
}
