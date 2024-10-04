// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ds

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/directoryservice/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_directory_service_directory", name="Directory")
func dataSourceDirectory() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceDirectoryRead,

		Schema: map[string]*schema.Schema{
			"access_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrAlias: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"connect_settings": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrAvailabilityZones: {
							Type:     schema.TypeSet,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"connect_ips": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"customer_dns_ips": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"customer_username": {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrSubnetIDs: {
							Type:     schema.TypeSet,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						names.AttrVPCID: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"directory_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"dns_ip_addresses": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"edition": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"enable_sso": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"radius_settings": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"authentication_protocol": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"display_label": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"radius_port": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"radius_retries": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"radius_servers": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"radius_timeout": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"use_same_username": {
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
			},
			"security_group_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"short_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrSize: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
			names.AttrType: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"vpc_settings": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrAvailabilityZones: {
							Type:     schema.TypeSet,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						names.AttrSubnetIDs: {
							Type:     schema.TypeSet,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						names.AttrVPCID: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceDirectoryRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DSClient(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	dir, err := findDirectoryByID(ctx, conn, d.Get("directory_id").(string))

	if err != nil {
		return sdkdiag.AppendFromErr(diags, tfresource.SingularDataSourceFindError("Directory Service Directory", err))
	}

	d.SetId(aws.ToString(dir.DirectoryId))
	d.Set("access_url", dir.AccessUrl)
	d.Set(names.AttrAlias, dir.Alias)
	if dir.ConnectSettings != nil {
		if err := d.Set("connect_settings", []interface{}{flattenDirectoryConnectSettingsDescription(dir.ConnectSettings, dir.DnsIpAddrs)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting connect_settings: %s", err)
		}
	} else {
		d.Set("connect_settings", nil)
	}
	d.Set(names.AttrDescription, dir.Description)
	if dir.Type == awstypes.DirectoryTypeAdConnector {
		d.Set("dns_ip_addresses", dir.ConnectSettings.ConnectIps)
	} else if dir.Type == awstypes.DirectoryTypeSharedMicrosoftAd {
		d.Set("dns_ip_addresses", dir.OwnerDirectoryDescription.DnsIpAddrs)
	} else {
		d.Set("dns_ip_addresses", dir.DnsIpAddrs)
	}
	d.Set("edition", dir.Edition)
	d.Set("enable_sso", dir.SsoEnabled)
	d.Set(names.AttrName, dir.Name)
	if dir.RadiusSettings != nil {
		if err := d.Set("radius_settings", []interface{}{flattenRadiusSettings(dir.RadiusSettings)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting radius_settings: %s", err)
		}
	} else {
		d.Set("radius_settings", nil)
	}
	if dir.Type == awstypes.DirectoryTypeAdConnector {
		d.Set("security_group_id", dir.ConnectSettings.SecurityGroupId)
	} else if dir.VpcSettings != nil {
		d.Set("security_group_id", dir.VpcSettings.SecurityGroupId)
	} else {
		d.Set("security_group_id", nil)
	}
	d.Set("short_name", dir.ShortName)
	d.Set(names.AttrSize, dir.Size)
	d.Set(names.AttrType, dir.Type)
	if dir.VpcSettings != nil {
		if err := d.Set("vpc_settings", []interface{}{flattenDirectoryVpcSettingsDescription(dir.VpcSettings)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting vpc_settings: %s", err)
		}
	} else {
		d.Set("vpc_settings", nil)
	}

	tags, err := listTags(ctx, conn, d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing tags for Directory Service Directory (%s): %s", d.Id(), err)
	}

	if err := d.Set(names.AttrTags, tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	return diags
}

func flattenRadiusSettings(apiObject *awstypes.RadiusSettings) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
		"authentication_protocol": apiObject.AuthenticationProtocol,
		"radius_retries":          apiObject.RadiusRetries,
		"use_same_username":       apiObject.UseSameUsername,
	}

	if v := apiObject.DisplayLabel; v != nil {
		tfMap["display_label"] = aws.ToString(v)
	}

	if v := apiObject.RadiusPort; v != nil {
		tfMap["radius_port"] = aws.ToInt32(v)
	}

	if v := apiObject.RadiusServers; v != nil {
		tfMap["radius_servers"] = v
	}

	if v := apiObject.RadiusTimeout; v != nil {
		tfMap["radius_timeout"] = aws.ToInt32(v)
	}

	return tfMap
}
