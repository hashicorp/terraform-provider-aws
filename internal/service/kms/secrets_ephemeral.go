// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kms

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/kms"
	awstypes "github.com/aws/aws-sdk-go-v2/service/kms/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	itypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @EphemeralResource(aws_kms_secrets, name="Secrets")
func newEphemeralSecrets(_ context.Context) (ephemeral.EphemeralResourceWithConfigure, error) {
	return &ephemeralSecrets{}, nil
}

type ephemeralSecrets struct {
	framework.EphemeralResourceWithConfigure
}

func (e *ephemeralSecrets) Schema(ctx context.Context, _ ephemeral.SchemaRequest, response *ephemeral.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"plaintext": schema.MapAttribute{
				CustomType: fwtypes.MapOfStringType,
				Computed:   true,
				Sensitive:  true,
			},
		},
		Blocks: map[string]schema.Block{
			"secret": schema.SetNestedBlock{
				CustomType: fwtypes.NewSetNestedObjectTypeOf[epSecrets](ctx),
				Validators: []validator.Set{
					setvalidator.IsRequired(),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"context": schema.MapAttribute{
							CustomType: fwtypes.MapOfStringType,
							Optional:   true,
						},
						"encryption_algorithm": schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.EncryptionAlgorithmSpec](),
							Optional:   true,
						},
						"grant_tokens": schema.ListAttribute{
							CustomType: fwtypes.ListOfStringType,
							Optional:   true,
						},
						names.AttrKeyID: schema.StringAttribute{
							Optional: true,
						},
						names.AttrName: schema.StringAttribute{
							Required: true,
						},
						"payload": schema.StringAttribute{
							Required: true,
						},
					},
				},
			},
		},
	}
}

func (e *ephemeralSecrets) Open(ctx context.Context, request ephemeral.OpenRequest, response *ephemeral.OpenResponse) {
	var data epSecretData
	conn := e.Meta().KMSClient(ctx)

	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	secrets, diags := data.Secrets.ToSlice(ctx)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	plaintext := make(map[string]attr.Value)

	for _, v := range secrets {
		input := kms.DecryptInput{}
		response.Diagnostics.Append(fwflex.Expand(ctx, v, &input)...)
		if response.Diagnostics.HasError() {
			return
		}
		input.EncryptionContext = fwflex.ExpandFrameworkStringValueMap(ctx, v.Context)

		payload, err := itypes.Base64Decode(v.Payload.ValueString())
		if err != nil {
			response.Diagnostics.AddError(
				"invalid base64 value for secret",
				err.Error(),
			)
			return
		}

		input.CiphertextBlob = payload

		output, err := conn.Decrypt(ctx, &input)
		if err != nil {
			response.Diagnostics.AddError(
				"failed to decrypt secret",
				err.Error(),
			)
			return
		}

		plaintext[v.Name.ValueString()] = fwflex.StringValueToFramework(ctx, string(output.Plaintext))
	}

	data.Plaintext = fwtypes.NewMapValueOfMust[types.String](ctx, plaintext)

	response.Diagnostics.Append(response.Result.Set(ctx, &data)...)
}

type epSecretData struct {
	Plaintext fwtypes.MapValueOf[types.String]          `tfsdk:"plaintext"`
	Secrets   fwtypes.SetNestedObjectValueOf[epSecrets] `tfsdk:"secret"`
}

type epSecrets struct {
	Context             fwtypes.MapValueOf[types.String]                     `tfsdk:"context"`
	EncryptionAlgorithm fwtypes.StringEnum[awstypes.EncryptionAlgorithmSpec] `tfsdk:"encryption_algorithm"`
	GrantTokens         fwtypes.ListValueOf[types.String]                    `tfsdk:"grant_tokens"`
	KeyID               types.String                                         `tfsdk:"key_id"`
	Name                types.String                                         `tfsdk:"name"`
	Payload             types.String                                         `tfsdk:"payload"`
}
