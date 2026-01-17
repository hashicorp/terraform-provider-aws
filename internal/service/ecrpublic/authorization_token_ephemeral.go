// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ecrpublic

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecrpublic"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	itypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @EphemeralResource(aws_ecrpublic_authorization_token, name="AuthorizationToken")
func newAuthorizationTokenEphemeralResource(_ context.Context) (ephemeral.EphemeralResourceWithConfigure, error) {
	return &authorizationTokenEphemeralResource{}, nil
}

type authorizationTokenEphemeralResource struct {
	framework.EphemeralResourceWithModel[authorizationTokenEphemeralResourceModel]
}

func (e *authorizationTokenEphemeralResource) Schema(ctx context.Context, _ ephemeral.SchemaRequest, response *ephemeral.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"authorization_token": schema.StringAttribute{
				Computed:  true,
				Sensitive: true,
			},
			"expires_at": schema.StringAttribute{
				Computed: true,
			},
			names.AttrPassword: schema.StringAttribute{
				Computed:  true,
				Sensitive: true,
			},
			names.AttrUserName: schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (e *authorizationTokenEphemeralResource) Open(ctx context.Context, request ephemeral.OpenRequest, response *ephemeral.OpenResponse) {
	conn := e.Meta().ECRPublicClient(ctx)
	data := authorizationTokenEphemeralResourceModel{}

	smerr.AddEnrich(ctx, &response.Diagnostics, request.Config.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	input := ecrpublic.GetAuthorizationTokenInput{}

	out, err := conn.GetAuthorizationToken(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err)
		return
	}

	if len(out.AuthorizationData) == 0 {
		smerr.AddError(ctx, &response.Diagnostics, fmt.Errorf("no authorization data returned"))
		return
	}

	authorizationData := out.AuthorizationData[0]
	authorizationToken := aws.ToString(authorizationData.AuthorizationToken)
	expiresAt := aws.ToTime(authorizationData.ExpiresAt).Format(time.RFC3339)
	authBytes, err := itypes.Base64Decode(authorizationToken)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, fmt.Errorf("decoding ECR Public authorization token: %w", err))
		return
	}

	basicAuthorization := strings.Split(string(authBytes), ":")
	if len(basicAuthorization) != 2 {
		smerr.AddError(ctx, &response.Diagnostics, fmt.Errorf("unknown ECR Public authorization token format"))
		return
	}

	userName := basicAuthorization[0]
	password := basicAuthorization[1]

	data.AuthorizationToken = types.StringValue(authorizationToken)
	data.ExpiresAt = types.StringValue(expiresAt)
	data.UserName = types.StringValue(userName)
	data.Password = types.StringValue(password)

	smerr.AddEnrich(ctx, &response.Diagnostics, response.Result.Set(ctx, &data))
}

type authorizationTokenEphemeralResourceModel struct {
	framework.WithRegionModel
	AuthorizationToken types.String `tfsdk:"authorization_token"`
	ExpiresAt          types.String `tfsdk:"expires_at"`
	Password           types.String `tfsdk:"password"`
	UserName           types.String `tfsdk:"user_name"`
}
