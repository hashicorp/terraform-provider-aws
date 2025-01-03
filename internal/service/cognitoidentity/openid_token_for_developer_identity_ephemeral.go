// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cognitoidentity

import (
	"context"
	"fmt"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentity"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/mapvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @EphemeralResource("aws_cognito_identity_openid_token_for_developer_identity", name="Cognito Identity Open ID Token for Developer Identity")
func newEphemeralCognitoIdentityOpenIDTokenForDeveloperIdentity(context.Context) (ephemeral.EphemeralResourceWithConfigure, error) {
	return &ephemeralCognitoIdentityOpenIDTokenForDeveloperIdentity{}, nil
}

const (
	EPNameCognitoIdentityOpenIDToken = "Cognito Identity Open ID Token for Developer Identity Ephemeral Resource"
)

type ephemeralCognitoIdentityOpenIDTokenForDeveloperIdentity struct {
	framework.EphemeralResourceWithConfigure
}

func (e *ephemeralCognitoIdentityOpenIDTokenForDeveloperIdentity) Metadata(_ context.Context, req ephemeral.MetadataRequest, resp *ephemeral.MetadataResponse) {
	resp.TypeName = "aws_cognito_identity_openid_token_for_developer_identity"
}

func (e *ephemeralCognitoIdentityOpenIDTokenForDeveloperIdentity) Schema(ctx context.Context, req ephemeral.SchemaRequest, resp *ephemeral.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"identity_id": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexache.MustCompile(`[\w-]+:[0-9a-f-]+`), "A unique identifier in the format REGION:GUID."),
					stringvalidator.LengthBetween(1, 55),
				},
			},
			names.AttrID: framework.IDAttribute(),
			"identity_pool_id": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexache.MustCompile(`[\w-]+:[0-9a-f-]+`), "A unique identifier in the format REGION:GUID."),
					stringvalidator.LengthBetween(1, 55),
				},
			},
			"logins": schema.MapAttribute{
				CustomType: fwtypes.NewMapTypeOf[types.String](ctx),
				Required:   true,
				Validators: []validator.Map{
					mapvalidator.KeysAre(
						stringvalidator.LengthBetween(1, 128),
					),
					mapvalidator.ValueStringsAre(
						stringvalidator.LengthBetween(1, 50000),
					),
					mapvalidator.SizeAtMost(10),
				},
			},
			"principal_tags": schema.MapAttribute{
				CustomType: fwtypes.NewMapTypeOf[types.String](ctx),
				Optional:   true,
				Validators: []validator.Map{
					mapvalidator.KeysAre(
						stringvalidator.LengthBetween(1, 128),
					),
					mapvalidator.ValueStringsAre(
						stringvalidator.LengthBetween(1, 256),
					),
					mapvalidator.SizeAtMost(50),
				},
			},
			"token_duration": schema.Int64Attribute{
				Optional: true,
				Validators: []validator.Int64{
					int64validator.Between(1, 86400),
				},
			},
			"token": schema.StringAttribute{
				Computed:  true,
				Sensitive: true,
			},
		},
	}
}

func (e *ephemeralCognitoIdentityOpenIDTokenForDeveloperIdentity) Open(ctx context.Context, req ephemeral.OpenRequest, resp *ephemeral.OpenResponse) {
	conn := e.Meta().CognitoIdentityClient(ctx)

	var data ephemeralCognitoIdentityOpenIDTokenForDeveloperIdentityModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var input cognitoidentity.GetOpenIdTokenForDeveloperIdentityInput
	resp.Diagnostics.Append(flex.Expand(ctx, data, &input)...)

	out, err := conn.GetOpenIdTokenForDeveloperIdentity(ctx, &input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.CognitoIdentity, create.ErrActionReading, EPNameCognitoIdentityOpenIDToken, data.IdentityId.String(), err),
			err.Error(),
		)
		return
	}

	data.Token = types.StringPointerValue(out.Token)
	data.IdentityId = types.StringPointerValue(out.IdentityId)
	data.ID = types.StringValue(
		fmt.Sprintf("%s/%s", data.IdentityId.ValueString(), data.IdentityPoolId.ValueString()),
	)
	resp.Diagnostics.Append(resp.Result.Set(ctx, &data)...)
}

type ephemeralCognitoIdentityOpenIDTokenForDeveloperIdentityModel struct {
	IdentityId     types.String                     `tfsdk:"identity_id"`
	IdentityPoolId types.String                     `tfsdk:"identity_pool_id"`
	ID             types.String                     `tfsdk:"id"`
	Logins         fwtypes.MapValueOf[types.String] `tfsdk:"logins"`
	PrincipalTags  fwtypes.MapValueOf[types.String] `tfsdk:"principal_tags"`
	TokenDuration  types.Int64                      `tfsdk:"token_duration"`
	Token          types.String                     `tfsdk:"token"`
}
