// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appconfig

import (
	"context"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_appconfig_configuration_profile", name="Configuration Profile")
// @Tags(identifierAttribute="arn")
func dataSourceConfigurationProfile() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceConfigurationProfileRead,

		Schema: map[string]*schema.Schema{
			names.AttrApplicationID: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringMatch(regexache.MustCompile(`[a-z\d]{4,7}`), ""),
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"configuration_profile_id": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringMatch(regexache.MustCompile(`[a-z\d]{4,7}`), ""),
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"location_uri": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"kms_key_identifier": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"retrieval_role_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
			names.AttrType: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"validator": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrContent: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrType: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceConfigurationProfileRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppConfigClient(ctx)

	applicationID := d.Get(names.AttrApplicationID).(string)
	configurationProfileID := d.Get("configuration_profile_id").(string)
	id := configurationProfileCreateResourceID(configurationProfileID, applicationID)

	output, err := findConfigurationProfileByTwoPartKey(ctx, conn, applicationID, configurationProfileID)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading AppConfig Configuration Profile (%s): %s", id, err)
	}

	d.SetId(id)
	d.Set(names.AttrApplicationID, output.ApplicationId)
	d.Set(names.AttrARN, configurationProfileARN(ctx, meta.(*conns.AWSClient), applicationID, configurationProfileID))
	d.Set("configuration_profile_id", output.Id)
	d.Set(names.AttrDescription, output.Description)
	d.Set("kms_key_identifier", output.KmsKeyIdentifier)
	d.Set("location_uri", output.LocationUri)
	d.Set(names.AttrName, output.Name)
	d.Set("retrieval_role_arn", output.RetrievalRoleArn)
	d.Set(names.AttrType, output.Type)
	if err := d.Set("validator", flattenValidators(output.Validators)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting validator: %s", err)
	}

	return diags
}
