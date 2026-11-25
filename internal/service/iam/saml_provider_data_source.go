// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package iam

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_iam_saml_provider", name="SAML Provider")
// @Tags
// @Testing(tagsTest=false)
func dataSourceSAMLProvider() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceSAMLProviderRead,

		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				names.AttrARN: {
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: verify.ValidARN,
				},
				"context_provider": dataSourceContextProviderSchema(),
				"create_date": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"identity_center_instance_arn": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"identity_store_arn": {
					Type:     schema.TypeString,
					Computed: true,
				},
				names.AttrName: {
					Type:     schema.TypeString,
					Computed: true,
				},
				"saml_key_hashes": {
					Type:     schema.TypeList,
					Computed: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"issuer_url": {
								Type:     schema.TypeString,
								Computed: true,
							},
							"key_hashes": {
								Type:     schema.TypeList,
								Computed: true,
								Elem:     &schema.Schema{Type: schema.TypeString},
							},
						},
					},
				},
				"saml_metadata_document": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"saml_provider_uuid": {
					Type:     schema.TypeString,
					Computed: true,
				},
				names.AttrTags: tftags.TagsSchemaOptionalComputedWarnShouldBeComputedOnly(),
				"valid_until": {
					Type:     schema.TypeString,
					Computed: true,
				},
			}
		},
	}
}

func dataSourceSAMLProviderRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).IAMClient(ctx)

	arn := d.Get(names.AttrARN).(string)
	output, err := findSAMLProviderByARN(ctx, conn, arn)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IAM SAML Provider (%s): %s", arn, err)
	}

	name, err := nameFromSAMLProviderARN(arn)

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	d.SetId(arn)
	if err := d.Set("context_provider", flattenContextProvider(output.ContextProvider)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting context_provider: %s", err)
	}
	if output.CreateDate != nil {
		d.Set("create_date", aws.ToTime(output.CreateDate).Format(time.RFC3339))
	} else {
		d.Set("create_date", nil)
	}
	d.Set("identity_center_instance_arn", output.IdentityCenterInstanceArn)
	d.Set("identity_store_arn", output.IdentityStoreArn)
	d.Set(names.AttrName, name)
	if err := d.Set("saml_key_hashes", flattenSAMLKeyHashEntries(output.SAMLKeyHashes)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting saml_key_hashes: %s", err)
	}
	d.Set("saml_metadata_document", output.SAMLMetadataDocument)
	d.Set("saml_provider_uuid", output.SAMLProviderUUID)
	if output.ValidUntil != nil {
		d.Set("valid_until", aws.ToTime(output.ValidUntil).Format(time.RFC3339))
	} else {
		d.Set("valid_until", nil)
	}

	setTagsOut(ctx, output.Tags)

	return diags
}

// dataSourceContextProviderSchema returns the schema of the trusted identity
// propagation context provider attribute shared by the SAML and OIDC provider
// data sources.
func dataSourceContextProviderSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Computed: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"key_definition": {
					Type:     schema.TypeList,
					Computed: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"source_key": {
								Type:     schema.TypeString,
								Computed: true,
							},
							"target_key": {
								Type:     schema.TypeString,
								Computed: true,
							},
						},
					},
				},
				"subject_source_key": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"subject_type_source_key": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"subject_type_value": {
					Type:     schema.TypeString,
					Computed: true,
				},
			},
		},
	}
}
