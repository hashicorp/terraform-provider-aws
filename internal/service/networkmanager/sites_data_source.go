package networkmanager

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/networkmanager"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

func DataSourceSites() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceSitesRead,

		Schema: map[string]*schema.Schema{
			"global_network_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"ids": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"tags": tftags.TagsSchema(),
		},
	}
}

func dataSourceSitesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).NetworkManagerConn()
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig
	tagsToMatch := tftags.New(d.Get("tags").(map[string]interface{})).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	output, err := FindSites(ctx, conn, &networkmanager.GetSitesInput{
		GlobalNetworkId: aws.String(d.Get("global_network_id").(string)),
	})

	if err != nil {
		return diag.Errorf("error listing Network Manager Sites: %s", err)
	}

	var siteIDs []string

	for _, v := range output {
		if len(tagsToMatch) > 0 {
			if !KeyValueTags(v.Tags).ContainsAll(tagsToMatch) {
				continue
			}
		}

		siteIDs = append(siteIDs, aws.StringValue(v.SiteId))
	}

	d.SetId(meta.(*conns.AWSClient).Region)
	d.Set("ids", siteIDs)

	return nil
}
