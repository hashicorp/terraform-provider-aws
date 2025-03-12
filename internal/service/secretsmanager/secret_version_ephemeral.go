// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package secretsmanager

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @EphemeralResource(aws_secretsmanager_secret_version, name="Secret Version")
func newEphemeralSecretVersion(_ context.Context) (ephemeral.EphemeralResourceWithConfigure, error) {
	return &ephemeralSecrets{}, nil
}

type ephemeralSecrets struct {
	framework.EphemeralResourceWithConfigure
}

func (e *ephemeralSecrets) Schema(ctx context.Context, _ ephemeral.SchemaRequest, response *ephemeral.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrCreatedDate: schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			"secret_id": schema.StringAttribute{
				Required: true,
			},
			"secret_binary": schema.StringAttribute{
				Computed:  true,
				Sensitive: true,
			},
			"secret_string": schema.StringAttribute{
				Computed:  true,
				Sensitive: true,
			},
			"version_id": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"version_stage": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"version_stages": schema.ListAttribute{
				CustomType: fwtypes.ListOfStringType,
				Computed:   true,
			},
		},
	}
}

func (e *ephemeralSecrets) Open(ctx context.Context, request ephemeral.OpenRequest, response *ephemeral.OpenResponse) {
	var data epSecretVersionData
	conn := e.Meta().SecretsManagerClient(ctx)

	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	input := secretsmanager.GetSecretValueInput{}
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	output, err := findSecretVersion(ctx, conn, &input)
	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.SecretsManager, create.ErrActionReading, DSNameSecretVersions, data.ARN.String(), err),
			err.Error(),
		)
		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data, fwflex.WithIgnoredFieldNamesAppend("SecretBinary"))...)
	if response.Diagnostics.HasError() {
		return
	}

	data.SecretBinary = fwflex.StringValueToFramework(ctx, string(output.SecretBinary))

	response.Diagnostics.Append(response.Result.Set(ctx, &data)...)
}

type epSecretVersionData struct {
	ARN           types.String                      `tfsdk:"arn"`
	CreatedDate   timetypes.RFC3339                 `tfsdk:"created_date"`
	SecretID      types.String                      `tfsdk:"secret_id"`
	SecretBinary  types.String                      `tfsdk:"secret_binary"`
	SecretString  types.String                      `tfsdk:"secret_string"`
	VersionID     types.String                      `tfsdk:"version_id"`
	VersionStage  types.String                      `tfsdk:"version_stage"`
	VersionStages fwtypes.ListValueOf[types.String] `tfsdk:"version_stages"`
}
