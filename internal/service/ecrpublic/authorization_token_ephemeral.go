// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ecrpublic

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecrpublic"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @EphemeralResource(aws_ecrpublic_authorization_token, name="AuthorizationToken")
func newAuthorizationTokenEphemeralResource(_ context.Context) (ephemeral.EphemeralResourceWithConfigure, error) {
	return &authorizationTokenEphemeralResource{}, nil
}

type authorizationTokenEphemeralResource struct {
	framework.EphemeralResourceWithModel[authorizationTokenEphemeralResourceModel]
}

func (e *authorizationTokenEphemeralResource) Schema(_ context.Context, _ ephemeral.SchemaRequest, response *ephemeral.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"authorization_token": schema.StringAttribute{
				Computed:  true,
				Sensitive: true,
			},
			"expires_at": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
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
	var data authorizationTokenEphemeralResourceModel

	smerr.AddEnrich(ctx, &response.Diagnostics, request.Config.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	var input ecrpublic.GetAuthorizationTokenInput
	output, err := conn.GetAuthorizationToken(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err)
		return
	}

	authorizationData := output.AuthorizationData
	authorizationToken := aws.ToString(authorizationData.AuthorizationToken)
	authBytes, err := inttypes.Base64Decode(authorizationToken)
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

	data.AuthorizationToken = fwflex.StringValueToFramework(ctx, authorizationToken)
	data.ExpiresAt = timetypes.NewRFC3339TimePointerValue(authorizationData.ExpiresAt)
	data.UserName = fwflex.StringValueToFramework(ctx, userName)
	data.Password = fwflex.StringValueToFramework(ctx, password)

	smerr.AddEnrich(ctx, &response.Diagnostics, response.Result.Set(ctx, &data))
}

type authorizationTokenEphemeralResourceModel struct {
	framework.WithRegionModel
	AuthorizationToken types.String      `tfsdk:"authorization_token"`
	ExpiresAt          timetypes.RFC3339 `tfsdk:"expires_at"`
	Password           types.String      `tfsdk:"password"`
	UserName           types.String      `tfsdk:"user_name"`
}
