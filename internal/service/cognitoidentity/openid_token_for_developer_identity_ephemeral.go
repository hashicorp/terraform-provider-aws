// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cognitoidentity

import (
	"context"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentity"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/mapvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
)

// @EphemeralResource("aws_cognito_identity_openid_token_for_developer_identity", name="Open ID Connect Token For Developer Identity")
func newOpenIDTokenForDeveloperIdentityEphemeralResource(context.Context) (ephemeral.EphemeralResourceWithConfigure, error) {
	return &openIDTokenForDeveloperIdentityEphemeralResource{}, nil
}

type openIDTokenForDeveloperIdentityEphemeralResource struct {
	framework.EphemeralResourceWithConfigure
}

func (e *openIDTokenForDeveloperIdentityEphemeralResource) Schema(ctx context.Context, request ephemeral.SchemaRequest, response *ephemeral.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"identity_id": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexache.MustCompile(`[\w-]+:[0-9a-f-]+`), "A unique identifier in the format REGION:GUID."),
					stringvalidator.LengthBetween(1, 55),
				},
			},
			"identity_pool_id": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexache.MustCompile(`[\w-]+:[0-9a-f-]+`), "A unique identifier in the format REGION:GUID."),
					stringvalidator.LengthBetween(1, 55),
				},
			},
			"logins": schema.MapAttribute{
				CustomType: fwtypes.MapOfStringType,
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
				CustomType: fwtypes.MapOfStringType,
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

func (e *openIDTokenForDeveloperIdentityEphemeralResource) Open(ctx context.Context, request ephemeral.OpenRequest, response *ephemeral.OpenResponse) {
	var data openIDTokenForDeveloperIdentityEphemeralResourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := e.Meta().CognitoIdentityClient(ctx)

	var input cognitoidentity.GetOpenIdTokenForDeveloperIdentityInput
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	output, err := conn.GetOpenIdTokenForDeveloperIdentity(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError("creating Cognito Identity Open ID Connect Token For Developer Identity", err.Error())

		return
	}

	data.IdentityID = fwflex.StringToFramework(ctx, output.IdentityId)
	data.Token = fwflex.StringToFramework(ctx, output.Token)

	response.Diagnostics.Append(response.Result.Set(ctx, &data)...)
}

type openIDTokenForDeveloperIdentityEphemeralResourceModel struct {
	IdentityID     types.String        `tfsdk:"identity_id"`
	IdentityPoolID types.String        `tfsdk:"identity_pool_id"`
	Logins         fwtypes.MapOfString `tfsdk:"logins"`
	PrincipalTags  fwtypes.MapOfString `tfsdk:"principal_tags"`
	Token          types.String        `tfsdk:"token"`
	TokenDuration  types.Int64         `tfsdk:"token_duration"`
}
