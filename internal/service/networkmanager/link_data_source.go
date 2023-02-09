package networkmanager

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

func DataSourceLink() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceLinkRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"bandwidth": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"download_speed": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"upload_speed": {
							Type:     schema.TypeInt,
							Computed: true,
						},
					},
				},
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"global_network_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"link_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"provider_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"site_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tftags.TagsSchemaComputed(),
			"type": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceLinkRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).NetworkManagerConn()
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	globalNetworkID := d.Get("global_network_id").(string)
	linkID := d.Get("link_id").(string)
	link, err := FindLinkByTwoPartKey(ctx, conn, globalNetworkID, linkID)

	if err != nil {
		return diag.Errorf("error reading Network Manager Link (%s): %s", linkID, err)
	}

	d.SetId(linkID)
	d.Set("arn", link.LinkArn)
	if link.Bandwidth != nil {
		if err := d.Set("bandwidth", []interface{}{flattenBandwidth(link.Bandwidth)}); err != nil {
			return diag.Errorf("error setting bandwidth: %s", err)
		}
	} else {
		d.Set("bandwidth", nil)
	}
	d.Set("description", link.Description)
	d.Set("global_network_id", link.GlobalNetworkId)
	d.Set("link_id", link.LinkId)
	d.Set("provider_name", link.Provider)
	d.Set("site_id", link.SiteId)
	d.Set("type", link.Type)

	if err := d.Set("tags", KeyValueTags(link.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return diag.Errorf("error setting tags: %s", err)
	}

	return nil
}
