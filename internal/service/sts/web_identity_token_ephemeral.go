// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package sts

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	awstypes "github.com/aws/aws-sdk-go-v2/service/sts/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
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
				CustomType: fwtypes.SetOfStringType,
				Required:   true,
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
			"duration_seconds": schema.Int32Attribute{
				Optional: true,
				Validators: []validator.Int32{
					int32validator.Between(60, 3600),
				},
				Description: "The duration, in seconds, for which the JWT will remain valid. Value can range from 60 to 3600 seconds. Default is 300 seconds (5 minutes).",
			},
			names.AttrTags: tftags.TagsAttribute(),
			"web_identity_token": schema.StringAttribute{
				Computed:    true,
				Sensitive:   true,
				Description: "The signed JWT token.",
			},
			"expiration": schema.StringAttribute{
				CustomType:  timetypes.RFC3339Type{},
				Computed:    true,
				Description: "The expiration time of the token in RFC3339 format.",
			},
		},
	}
}

func (e *webIdentityTokenEphemeralResource) Open(ctx context.Context, request ephemeral.OpenRequest, response *ephemeral.OpenResponse) {
	conn := e.Meta().STSClient(ctx)
	var data webIdentityTokenEphemeralResourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Config.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	var input sts.GetWebIdentityTokenInput
	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Expand(ctx, data, &input))
	if response.Diagnostics.HasError() {
		return
	}

	// expand tags since this is not using transparent tagging
	if !data.Tags.IsNull() {
		tags := tftags.New(ctx, data.Tags)
		tagMap := make([]awstypes.Tag, 0, len(tags.Map()))
		for k, v := range tags.Map() {
			tag := awstypes.Tag{
				Key:   aws.String(k),
				Value: aws.String(v),
			}

			tagMap = append(tagMap, tag)
		}

		input.Tags = tagMap
	}

	output, err := conn.GetWebIdentityToken(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err)
		return
	}

	data.WebIdentityToken = fwflex.StringToFramework(ctx, output.WebIdentityToken)
	data.Expiration = timetypes.NewRFC3339TimePointerValue(output.Expiration)

	response.Diagnostics.Append(response.Result.Set(ctx, &data)...)
}

type webIdentityTokenEphemeralResourceModel struct {
	Audience         fwtypes.SetOfString `tfsdk:"audience"`
	SigningAlgorithm types.String        `tfsdk:"signing_algorithm"`
	DurationSeconds  types.Int32         `tfsdk:"duration_seconds"`
	Tags             tftags.Map          `tfsdk:"tags"`
	WebIdentityToken types.String        `tfsdk:"web_identity_token"`
	Expiration       timetypes.RFC3339   `tfsdk:"expiration"`
}
