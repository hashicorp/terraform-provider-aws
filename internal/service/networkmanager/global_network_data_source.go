package networkmanager

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

func DataSourceGlobalNetwork() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceGlobalNetworkRead,

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
			"tags": tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceGlobalNetworkRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).NetworkManagerConn()
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	globalNetworkID := d.Get("global_network_id").(string)
	globalNetwork, err := FindGlobalNetworkByID(ctx, conn, globalNetworkID)

	if err != nil {
		return diag.Errorf("error reading Network Manager Global Network (%s): %s", globalNetworkID, err)
	}

	d.SetId(globalNetworkID)
	d.Set("arn", globalNetwork.GlobalNetworkArn)
	d.Set("description", globalNetwork.Description)
	d.Set("global_network_id", globalNetwork.GlobalNetworkId)

	if err := d.Set("tags", KeyValueTags(globalNetwork.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return diag.Errorf("error setting tags: %s", err)
	}

	return nil
}
