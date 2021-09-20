package ds

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directoryservice"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

func DataSourceDirectory() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceDirectoryRead,

		Schema: map[string]*schema.Schema{
			"directory_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"size": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"alias": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"short_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tftags.TagsSchema(),
			"vpc_settings": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"subnet_ids": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"vpc_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"availability_zones": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"connect_settings": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"connect_ips": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"customer_username": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"customer_dns_ips": {
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
						"availability_zones": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"enable_sso": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"access_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"dns_ip_addresses": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
			"security_group_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"edition": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceDirectoryRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DirectoryServiceConn
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	directoryID := d.Get("directory_id").(string)
	out, err := conn.DescribeDirectories(&directoryservice.DescribeDirectoriesInput{
		DirectoryIds: []*string{aws.String(directoryID)},
	})
	if err != nil {
		if tfawserr.ErrMessageContains(err, directoryservice.ErrCodeEntityDoesNotExistException, "") {
			return fmt.Errorf("DirectoryService Directory (%s) not found", directoryID)
		}
		return fmt.Errorf("error reading DirectoryService Directory: %w", err)
	}

	if out == nil || len(out.DirectoryDescriptions) == 0 {
		return fmt.Errorf("error reading DirectoryService Directory (%s): empty output", directoryID)
	}

	d.SetId(directoryID)

	dir := out.DirectoryDescriptions[0]
	log.Printf("[DEBUG] Received DS directory: %s", dir)

	d.Set("access_url", dir.AccessUrl)
	d.Set("alias", dir.Alias)
	d.Set("description", dir.Description)

	var addresses []interface{}
	if aws.StringValue(dir.Type) == directoryservice.DirectoryTypeAdconnector {
		addresses = flex.FlattenStringList(dir.ConnectSettings.ConnectIps)
	} else {
		addresses = flex.FlattenStringList(dir.DnsIpAddrs)
	}
	if err := d.Set("dns_ip_addresses", addresses); err != nil {
		return fmt.Errorf("error setting dns_ip_addresses: %w", err)
	}

	d.Set("name", dir.Name)
	d.Set("short_name", dir.ShortName)
	d.Set("size", dir.Size)
	d.Set("edition", dir.Edition)
	d.Set("type", dir.Type)

	if err := d.Set("vpc_settings", flattenVPCSettings(dir.VpcSettings)); err != nil {
		return fmt.Errorf("error setting VPC settings: %w", err)
	}

	if err := d.Set("connect_settings", flattenConnectSettings(dir.DnsIpAddrs, dir.ConnectSettings)); err != nil {
		return fmt.Errorf("error setting connect settings: %w", err)
	}

	d.Set("enable_sso", dir.SsoEnabled)

	var securityGroupId *string
	if aws.StringValue(dir.Type) == directoryservice.DirectoryTypeAdconnector && dir.ConnectSettings != nil {
		securityGroupId = dir.ConnectSettings.SecurityGroupId
	} else if dir.VpcSettings != nil {
		securityGroupId = dir.VpcSettings.SecurityGroupId
	}
	d.Set("security_group_id", securityGroupId)

	tags, err := tftags.DirectoryserviceListTags(conn, d.Id())
	if err != nil {
		return fmt.Errorf("error listing tags for Directory Service Directory (%s): %w", d.Id(), err)
	}

	if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	return nil
}
