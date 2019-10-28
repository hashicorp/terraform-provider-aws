package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/directoryservice"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func dataSourceDirectoryService() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceDirectoryServiceRead,

		Schema: map[string]*schema.Schema{
			"filter": dataSourceFiltersSchema(),
			"directory_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"short_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"access_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"dns_ip_addresses": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"owner_dns_ip_addresses": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"owner_directory_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceDirectoryServiceRead(d *schema.ResourceData, meta interface{}) error {
	dsconn := meta.(*AWSClient).dsconn
	name, nameExists := d.GetOk("name")
	id, idExists := d.GetOk("directory_id")
	req := &directoryservice.DescribeDirectoriesInput{}

	if nameExists && idExists {
		return fmt.Errorf("directory_id and name arguments can't be used together")
	}

	if !nameExists && !idExists {
		return fmt.Errorf("Either name or directory_id must be set")
	}

	resp, err := dsconn.DescribeDirectories(req)

	if err != nil {
		return fmt.Errorf("error describing directories: %s", err)
	}

	var directoryDescriptionFound *directoryservice.DirectoryDescription

	for _, directoryDescription := range resp.DirectoryDescriptions {
		directoryDescriptionName := *directoryDescription.Name
		directoryDescriptionID := *directoryDescription.DirectoryId
		// Try to match by directory_id
		if idExists && directoryDescriptionID == id.(string) {
			directoryDescriptionFound = directoryDescription
			break
			//  Try to match by name
		} else if nameExists && directoryDescriptionName == name.(string) {
			directoryDescriptionFound = directoryDescription
			break
		}
	}

	if directoryDescriptionFound == nil {
		return fmt.Errorf("no matching directory service found")
	}

	idDirectoryService := (*directoryDescriptionFound.DirectoryId)
	d.SetId(idDirectoryService)
	d.Set("name", directoryDescriptionFound.Name)
	d.Set("short_name", directoryDescriptionFound.ShortName)
	d.Set("access_url", directoryDescriptionFound.AccessUrl)
	d.Set("dns_ip_addresses", directoryDescriptionFound.DnsIpAddrs)
	d.Set("owner_dns_ip_addresses", directoryDescriptionFound.OwnerDirectoryDescription.DnsIpAddrs)
	d.Set("owner_directory_id", directoryDescriptionFound.OwnerDirectoryDescription.DirectoryId)

	return nil
}
