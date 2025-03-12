// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package secretsmanager

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @EphemeralResource("aws_secretsmanager_random_password", name="Random Password")
func newEphemeralRandomPassword(context.Context) (ephemeral.EphemeralResourceWithConfigure, error) {
	return &ephemeralRandomPassword{}, nil
}

const (
	ERNameRandomPassword = "Random Password Ephemeral Resource"
)

type ephemeralRandomPassword struct {
	framework.EphemeralResourceWithConfigure
}

func (e *ephemeralRandomPassword) Schema(ctx context.Context, req ephemeral.SchemaRequest, resp *ephemeral.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"exclude_characters": schema.StringAttribute{
				Optional: true,
			},
			"exclude_lowercase": schema.BoolAttribute{
				Optional: true,
			},
			"exclude_numbers": schema.BoolAttribute{
				Optional: true,
			},
			"exclude_punctuation": schema.BoolAttribute{
				Optional: true,
			},
			"exclude_uppercase": schema.BoolAttribute{
				Optional: true,
			},
			"include_space": schema.BoolAttribute{
				Optional: true,
			},
			"password_length": schema.Int64Attribute{
				Optional: true,
			},
			"require_each_included_type": schema.BoolAttribute{
				Optional: true,
			},
			"random_password": schema.StringAttribute{
				Computed:  true,
				Sensitive: true,
			},
		},
	}
}

func (e *ephemeralRandomPassword) Open(ctx context.Context, req ephemeral.OpenRequest, resp *ephemeral.OpenResponse) {
	conn := e.Meta().SecretsManagerClient(ctx)

	var data ephemeralRandomPasswordModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := secretsmanager.GetRandomPasswordInput{}
	resp.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if resp.Diagnostics.HasError() {
		return
	}

	output, err := conn.GetRandomPassword(ctx, &input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.SecretsManager, create.ErrActionOpening, ERNameRandomPassword, "", err),
			err.Error(),
		)
		return
	}

	data.RandomPassword = fwflex.StringToFramework(ctx, output.RandomPassword)

	resp.Diagnostics.Append(resp.Result.Set(ctx, &data)...)
}

type ephemeralRandomPasswordModel struct {
	ExcludeCharacters       types.String `tfsdk:"exclude_characters"`
	ExcludeLowercase        types.Bool   `tfsdk:"exclude_lowercase"`
	ExcludeNumbers          types.Bool   `tfsdk:"exclude_numbers"`
	ExcludePunctuation      types.Bool   `tfsdk:"exclude_punctuation"`
	ExcludeUppercase        types.Bool   `tfsdk:"exclude_uppercase"`
	IncludeSpace            types.Bool   `tfsdk:"include_space"`
	PasswordLength          types.Int64  `tfsdk:"password_length"`
	RandomPassword          types.String `tfsdk:"random_password"`
	RequireEachIncludedType types.Bool   `tfsdk:"require_each_included_type"`
}
