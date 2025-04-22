// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ecr

import (
	"context"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_ecr_repository_creation_template", name="Repository Creation Template")
func dataSourceRepositoryCreationTemplate() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceRepositoryCreationTemplateRead,

		Schema: map[string]*schema.Schema{
			"applied_for": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"custom_role_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrEncryptionConfiguration: {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"encryption_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrKMSKey: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"image_tag_mutability": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"lifecycle_policy": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrPrefix: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(2, 30),
					validation.StringMatch(
						regexache.MustCompile(`(?:ROOT|(?:[a-z0-9]+(?:[._-][a-z0-9]+)*/)*[a-z0-9]+(?:[._-][a-z0-9]+)*)`),
						"must only include alphanumeric, underscore, period, hyphen, or slash characters, or be the string `ROOT`"),
				),
			},
			"registry_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"repository_policy": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrResourceTags: tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceRepositoryCreationTemplateRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ECRClient(ctx)

	prefix := d.Get(names.AttrPrefix).(string)

	rct, registryID, err := findRepositoryCreationTemplateByRepositoryPrefix(ctx, conn, prefix)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ECR Repository Creation Template (%s): %s", prefix, err)
	}

	d.SetId(aws.ToString(rct.Prefix))
	d.Set("applied_for", rct.AppliedFor)
	d.Set("custom_role_arn", rct.CustomRoleArn)
	d.Set(names.AttrDescription, rct.Description)
	if err := d.Set(names.AttrEncryptionConfiguration, flattenRepositoryEncryptionConfigurationForRepositoryCreationTemplate(rct.EncryptionConfiguration)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting encryption_configuration: %s", err)
	}
	d.Set("image_tag_mutability", rct.ImageTagMutability)

	policy, err := structure.NormalizeJsonString(aws.ToString(rct.LifecyclePolicy))
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	d.Set("lifecycle_policy", policy)
	d.Set(names.AttrPrefix, rct.Prefix)
	d.Set("registry_id", registryID)

	policy, err = structure.NormalizeJsonString(aws.ToString(rct.RepositoryPolicy))
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	d.Set("repository_policy", policy)
	d.Set(names.AttrResourceTags, keyValueTags(ctx, rct.ResourceTags).Map())

	return diags
}
