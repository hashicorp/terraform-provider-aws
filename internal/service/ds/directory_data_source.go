package ds

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directoryservice"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func DataSourceDirectory() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceDirectoryRead,

		Schema: map[string]*schema.Schema{
			"access_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"alias": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"connect_settings": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"availability_zones": {
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
						"subnet_ids": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"vpc_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"description": {
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
			"name": {
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
			"size": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tftags.TagsSchemaComputed(),
			"type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"vpc_settings": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"availability_zones": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"subnet_ids": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"vpc_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceDirectoryRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DSConn
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	dir, err := FindDirectoryByID(conn, d.Get("directory_id").(string))

	if err != nil {
		return tfresource.SingularDataSourceFindError("Directory Service Directory", err)
	}

	d.SetId(aws.StringValue(dir.DirectoryId))
	d.Set("access_url", dir.AccessUrl)
	d.Set("alias", dir.Alias)
	if dir.ConnectSettings != nil {
		if err := d.Set("connect_settings", []interface{}{flattenDirectoryConnectSettingsDescription(dir.ConnectSettings, dir.DnsIpAddrs)}); err != nil {
			return fmt.Errorf("setting connect_settings: %w", err)
		}
	} else {
		d.Set("connect_settings", nil)
	}
	d.Set("description", dir.Description)
	if aws.StringValue(dir.Type) == directoryservice.DirectoryTypeAdconnector {
		d.Set("dns_ip_addresses", aws.StringValueSlice(dir.ConnectSettings.ConnectIps))
	} else if aws.StringValue(dir.Type) == directoryservice.DirectoryTypeSharedMicrosoftAd {
		d.Set("dns_ip_addresses", aws.StringValueSlice(dir.OwnerDirectoryDescription.DnsIpAddrs))
	} else {
		d.Set("dns_ip_addresses", aws.StringValueSlice(dir.DnsIpAddrs))
	}
	d.Set("edition", dir.Edition)
	d.Set("enable_sso", dir.SsoEnabled)
	d.Set("name", dir.Name)
	if dir.RadiusSettings != nil {
		if err := d.Set("radius_settings", []interface{}{flattenRadiusSettings(dir.RadiusSettings)}); err != nil {
			return fmt.Errorf("setting radius_settings: %w", err)
		}
	} else {
		d.Set("radius_settings", nil)
	}
	if aws.StringValue(dir.Type) == directoryservice.DirectoryTypeAdconnector {
		d.Set("security_group_id", dir.ConnectSettings.SecurityGroupId)
	} else if dir.VpcSettings != nil {
		d.Set("security_group_id", dir.VpcSettings.SecurityGroupId)
	} else {
		d.Set("security_group_id", nil)
	}
	d.Set("short_name", dir.ShortName)
	d.Set("size", dir.Size)
	d.Set("type", dir.Type)
	if dir.VpcSettings != nil {
		if err := d.Set("vpc_settings", []interface{}{flattenDirectoryVpcSettingsDescription(dir.VpcSettings)}); err != nil {
			return fmt.Errorf("setting vpc_settings: %w", err)
		}
	} else {
		d.Set("vpc_settings", nil)
	}

	tags, err := ListTags(conn, d.Id())

	if err != nil {
		return fmt.Errorf("listing tags for Directory Service Directory (%s): %w", d.Id(), err)
	}

	if err := d.Set("tags", tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("setting tags: %w", err)
	}

	return nil
}

func flattenRadiusSettings(apiObject *directoryservice.RadiusSettings) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.AuthenticationProtocol; v != nil {
		tfMap["authentication_protocol"] = aws.StringValue(v)
	}

	if v := apiObject.DisplayLabel; v != nil {
		tfMap["display_label"] = aws.StringValue(v)
	}

	if v := apiObject.RadiusPort; v != nil {
		tfMap["radius_port"] = aws.Int64Value(v)
	}

	if v := apiObject.RadiusRetries; v != nil {
		tfMap["radius_retries"] = aws.Int64Value(v)
	}

	if v := apiObject.RadiusServers; v != nil {
		tfMap["radius_servers"] = aws.StringValueSlice(v)
	}

	if v := apiObject.RadiusTimeout; v != nil {
		tfMap["radius_timeout"] = aws.Int64Value(v)
	}

	if v := apiObject.UseSameUsername; v != nil {
		tfMap["use_same_username"] = aws.BoolValue(v)
	}

	return tfMap
}
