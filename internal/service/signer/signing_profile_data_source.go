// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package signer

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/signer"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_signer_signing_profile")
func DataSourceSigningProfile() *schema.Resource {
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

func dataSourceSigningProfileRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SignerClient(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

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

	if err := d.Set("signature_validity_period", []interface{}{
		map[string]interface{}{
			names.AttrValue: signingProfileOutput.SignatureValidityPeriod.Value,
			names.AttrType:  signingProfileOutput.SignatureValidityPeriod.Type,
		},
	}); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting signer signing profile signature validity period: %s", err)
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

	if err := d.Set(names.AttrStatus, signingProfileOutput.Status); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting signer signing profile status: %s", err)
	}

	if err := d.Set(names.AttrTags, KeyValueTags(ctx, signingProfileOutput.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting signer signing profile tags: %s", err)
	}

	if err := d.Set("revocation_record", flattenSigningProfileRevocationRecord(signingProfileOutput.RevocationRecord)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting signer signing profile revocation record: %s", err)
	}

	d.SetId(aws.ToString(signingProfileOutput.ProfileName))

	return diags
}
