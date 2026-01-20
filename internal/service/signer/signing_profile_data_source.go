// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package signer

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/signer"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_signer_signing_profile", name="Signing Profile")
func dataSourceSigningProfile() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceSigningProfileRead,

		Schema: map[string]*schema.Schema{
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"platform_display_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"platform_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"revocation_record": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"revocation_effective_from": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"revoked_at": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"revoked_by": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"signature_validity_period": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrValue: {
							Type:     schema.TypeInt,
							Computed: true,
						},
						names.AttrType: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"signing_material": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrCertificateARN: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"signing_parameters": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			names.AttrStatus: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
			names.AttrVersion: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"version_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceSigningProfileRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SignerClient(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig(ctx)

	profileName := d.Get(names.AttrName).(string)
	signingProfileOutput, err := conn.GetSigningProfile(ctx, &signer.GetSigningProfileInput{
		ProfileName: aws.String(profileName),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Signer signing profile (%s): %s", d.Id(), err)
	}

	if err := d.Set("platform_id", signingProfileOutput.PlatformId); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting signer signing profile platform id: %s", err)
	}

	if v := signingProfileOutput.SignatureValidityPeriod; v != nil {
		if err := d.Set("signature_validity_period", []any{
			map[string]any{
				names.AttrValue: v.Value,
				names.AttrType:  v.Type,
			},
		}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting signature_validity_period: %s", err)
		}
	}

	if err := d.Set("platform_display_name", signingProfileOutput.PlatformDisplayName); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting signer signing profile platform display name: %s", err)
	}

	if err := d.Set(names.AttrARN, signingProfileOutput.Arn); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting signer signing profile arn: %s", err)
	}

	if err := d.Set(names.AttrVersion, signingProfileOutput.ProfileVersion); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting signer signing profile version: %s", err)
	}

	if err := d.Set("version_arn", signingProfileOutput.ProfileVersionArn); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting signer signing profile version arn: %s", err)
	}

	if signingProfileOutput.SigningMaterial != nil {
		if err := d.Set("signing_material", flattenSigningMaterial(signingProfileOutput.SigningMaterial)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting signing_material: %s", err)
		}
	}

	if signingProfileOutput.SigningParameters != nil {
		if err := d.Set("signing_parameters", flex.FlattenStringValueMap(signingProfileOutput.SigningParameters)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting signing_parameters: %s", err)
		}
	}

	if err := d.Set(names.AttrStatus, signingProfileOutput.Status); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting signer signing profile status: %s", err)
	}

	if err := d.Set(names.AttrTags, keyValueTags(ctx, signingProfileOutput.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting signer signing profile tags: %s", err)
	}

	if err := d.Set("revocation_record", flattenSigningProfileRevocationRecord(signingProfileOutput.RevocationRecord)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting signer signing profile revocation record: %s", err)
	}

	d.SetId(aws.ToString(signingProfileOutput.ProfileName))

	return diags
}
