package networkmanager

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

func DataSourceSite() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceSiteRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"global_network_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"location": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"address": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"latitude": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"longitude": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"site_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"tags": tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceSiteRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).NetworkManagerConn()
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	globalNetworkID := d.Get("global_network_id").(string)
	siteID := d.Get("site_id").(string)
	site, err := FindSiteByTwoPartKey(ctx, conn, globalNetworkID, siteID)

	if err != nil {
		return diag.Errorf("error reading Network Manager Site (%s): %s", siteID, err)
	}

	d.SetId(siteID)
	d.Set("arn", site.SiteArn)
	d.Set("description", site.Description)
	d.Set("global_network_id", site.GlobalNetworkId)
	if site.Location != nil {
		if err := d.Set("location", []interface{}{flattenLocation(site.Location)}); err != nil {
			return diag.Errorf("error setting location: %s", err)
		}
	} else {
		d.Set("location", nil)
	}
	d.Set("site_id", site.SiteId)

	if err := d.Set("tags", KeyValueTags(site.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return diag.Errorf("error setting tags: %s", err)
	}

	return nil
}
