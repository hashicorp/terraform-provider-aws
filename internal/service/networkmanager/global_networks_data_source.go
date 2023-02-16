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

func DataSourceGlobalNetworks() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceGlobalNetworksRead,

		Schema: map[string]*schema.Schema{
			"ids": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"tags": tftags.TagsSchema(),
		},
	}
}

func dataSourceGlobalNetworksRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).NetworkManagerConn()
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig
	tagsToMatch := tftags.New(d.Get("tags").(map[string]interface{})).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	output, err := FindGlobalNetworks(ctx, conn, &networkmanager.DescribeGlobalNetworksInput{})

	if err != nil {
		return diag.Errorf("error listing Network Manager Global Networks: %s", err)
	}

	var globalNetworkIDs []string

	for _, v := range output {
		if len(tagsToMatch) > 0 {
			if !KeyValueTags(v.Tags).ContainsAll(tagsToMatch) {
				continue
			}
		}

		globalNetworkIDs = append(globalNetworkIDs, aws.StringValue(v.GlobalNetworkId))
	}

	d.SetId(meta.(*conns.AWSClient).Region)
	d.Set("ids", globalNetworkIDs)

	return nil
}
