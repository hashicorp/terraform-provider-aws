// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package sts

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	awstypes "github.com/aws/aws-sdk-go-v2/service/sts/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/mapvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @EphemeralResource("aws_sts_web_identity_token", name="Web Identity Token")
func newWebIdentityTokenEphemeralResource(_ context.Context) (ephemeral.EphemeralResourceWithConfigure, error) {
	return &webIdentityTokenEphemeralResource{}, nil
}

type webIdentityTokenEphemeralResource struct {
	framework.EphemeralResourceWithModel[webIdentityTokenEphemeralResourceModel]
}

func (e *webIdentityTokenEphemeralResource) Schema(ctx context.Context, _ ephemeral.SchemaRequest, response *ephemeral.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"audience": schema.SetAttribute{
				CustomType:  fwtypes.SetOfStringType,
				Required:    true,
				ElementType: types.StringType,
				Validators: []validator.Set{
					setvalidator.SizeBetween(1, 10),
					setvalidator.ValueStringsAre(
						stringvalidator.LengthBetween(1, 1000),
					),
				},
				Description: "The intended recipients of the token (populates the `aud` claim in the JWT). Must contain between 1 and 10 items.",
			},
			"signing_algorithm": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.OneOf("RS256", "ES384"),
				},
				Description: "The cryptographic algorithm to use for signing the JWT. Valid values are `RS256` (RSA with SHA-256) and `ES384` (ECDSA using P-384 curve with SHA-384).",
			},
			"duration_seconds": schema.Int64Attribute{
				Optional: true,
				Validators: []validator.Int64{
					int64validator.Between(60, 3600),
				},
				Description: "The duration, in seconds, for which the JWT will remain valid. Value can range from 60 to 3600 seconds. Default is 300 seconds (5 minutes).",
			},
			names.AttrTags: schema.MapAttribute{
				CustomType:  fwtypes.MapOfStringType,
				Optional:    true,
				ElementType: types.StringType,
				Validators: []validator.Map{
					mapvalidator.SizeAtMost(50),
					mapvalidator.KeysAre(
						stringvalidator.LengthBetween(1, 128),
					),
					mapvalidator.ValueStringsAre(
						stringvalidator.LengthBetween(1, 256),
					),
				},
				Description: "Custom claims to include in the JWT. Maximum of 50 tags.",
			},
			"web_identity_token": schema.StringAttribute{
				Computed:    true,
				Sensitive:   true,
				Description: "The signed JWT token.",
			},
			"expiration": schema.StringAttribute{
				Computed:    true,
				Description: "The expiration time of the token in RFC3339 format.",
			},
		},
	}
}

func (e *webIdentityTokenEphemeralResource) Open(ctx context.Context, request ephemeral.OpenRequest, response *ephemeral.OpenResponse) {
	var data webIdentityTokenEphemeralResourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := e.Meta().STSClient(ctx)

	// Convert audience set to slice
	var audienceSlice []string
	response.Diagnostics.Append(data.Audience.ElementsAs(ctx, &audienceSlice, false)...)
	if response.Diagnostics.HasError() {
		return
	}

	input := &sts.GetWebIdentityTokenInput{
		Audience:         audienceSlice,
		SigningAlgorithm: data.SigningAlgorithm.ValueStringPointer(),
	}

	if !data.DurationSeconds.IsNull() {
		input.DurationSeconds = aws.Int32(int32(data.DurationSeconds.ValueInt64()))
	}

	if !data.Tags.IsNull() {
		var tagsMap map[string]string
		response.Diagnostics.Append(data.Tags.ElementsAs(ctx, &tagsMap, false)...)
		if response.Diagnostics.HasError() {
			return
		}

		var sdkTags []awstypes.Tag
		for k, v := range tagsMap {
			sdkTags = append(sdkTags, awstypes.Tag{
				Key:   aws.String(k),
				Value: aws.String(v),
			})
		}
		input.Tags = sdkTags
	}

	output, err := conn.GetWebIdentityToken(ctx, input)
	if err != nil {
		response.Diagnostics.AddError("creating STS Web Identity Token", err.Error())
		return
	}

	data.WebIdentityToken = types.StringValue(aws.ToString(output.WebIdentityToken))
	data.Expiration = types.StringValue(output.Expiration.Format(time.RFC3339))

	response.Diagnostics.Append(response.Result.Set(ctx, &data)...)
}

type webIdentityTokenEphemeralResourceModel struct {
	Audience         fwtypes.SetOfString `tfsdk:"audience"`
	SigningAlgorithm types.String        `tfsdk:"signing_algorithm"`
	DurationSeconds  types.Int64         `tfsdk:"duration_seconds"`
	Tags             fwtypes.MapOfString `tfsdk:"tags"`
	WebIdentityToken types.String        `tfsdk:"web_identity_token"`
	Expiration       types.String        `tfsdk:"expiration"`
}
