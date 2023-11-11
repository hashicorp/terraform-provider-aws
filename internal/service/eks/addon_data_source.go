// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package eks

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

// @SDKDataSource("aws_eks_addon")
func dataSourceAddon() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceAddonRead,
		Schema: map[string]*schema.Schema{
			"addon_name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"addon_version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cluster_name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validClusterName,
			},
			"configuration_values": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"modified_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"service_account_role_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceAddonRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EKSClient(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	addonName := d.Get("addon_name").(string)
	clusterName := d.Get("cluster_name").(string)
	id := AddonCreateResourceID(clusterName, addonName)

	addon, err := findAddonByTwoPartKey(ctx, conn, clusterName, addonName)

	if err != nil {
		return diag.Errorf("reading EKS Add-On (%s): %s", id, err)
	}

	d.SetId(id)
	d.Set("addon_version", addon.AddonVersion)
	d.Set("arn", addon.AddonArn)
	d.Set("configuration_values", addon.ConfigurationValues)
	d.Set("created_at", aws.ToTime(addon.CreatedAt).Format(time.RFC3339))
	d.Set("modified_at", aws.ToTime(addon.ModifiedAt).Format(time.RFC3339))
	d.Set("service_account_role_arn", addon.ServiceAccountRoleArn)

	if err := d.Set("tags", KeyValueTags(ctx, addon.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return diag.Errorf("setting tags: %s", err)
	}

	return nil
}
