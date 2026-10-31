// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package meta

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
)

// Function annotations are used for ephemeral registration to the Provider. DO NOT EDIT.
// @EphemeralResource(aws_credentials, name="Credentials")
func newEphemeralCredentials(_ context.Context) (ephemeral.EphemeralResourceWithConfigure, error) {
	return &ephemeralCredentials{}, nil
}

type ephemeralCredentials struct {
	framework.EphemeralResourceWithModel[ephemeralCredentialsModel]
}

func (e *ephemeralCredentials) Schema(ctx context.Context, _ ephemeral.SchemaRequest, resp *ephemeral.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"access_key_id": schema.StringAttribute{
				Computed:  true,
				Sensitive: true,
			},
			"secret_access_key": schema.StringAttribute{
				Computed:  true,
				Sensitive: true,
			},
			"session_token": schema.StringAttribute{
				Computed:  true,
				Sensitive: true,
			},
		},
	}
}

func (e *ephemeralCredentials) Open(ctx context.Context, req ephemeral.OpenRequest, resp *ephemeral.OpenResponse) {
	var data ephemeralCredentialsModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	credentialsProvider := e.Meta().CredentialsProvider(ctx)
	if credentialsProvider == nil {
		resp.Diagnostics.AddError("Missing Credentials Provider", "No AWS credentials provider is configured.")
		return
	}

	creds, err := credentialsProvider.Retrieve(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to Retrieve Credentials", err.Error())
		return
	}

	data.AccessKeyID = types.StringValue(creds.AccessKeyID)
	data.SecretAccessKey = types.StringValue(creds.SecretAccessKey)
	data.SessionToken = types.StringValue(creds.SessionToken)

	resp.Diagnostics.Append(resp.Result.Set(ctx, &data)...)
}

type ephemeralCredentialsModel struct {
	framework.WithRegionModel
	AccessKeyID     types.String `tfsdk:"access_key_id"`
	SecretAccessKey types.String `tfsdk:"secret_access_key"`
	SessionToken    types.String `tfsdk:"session_token"`
}
