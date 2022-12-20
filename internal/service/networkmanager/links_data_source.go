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

func DataSourceLinks() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceLinksRead,

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
			"provider_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"site_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"tags": tftags.TagsSchema(),
			"type": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func dataSourceLinksRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).NetworkManagerConn()
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig
	tagsToMatch := tftags.New(d.Get("tags").(map[string]interface{})).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	input := &networkmanager.GetLinksInput{
		GlobalNetworkId: aws.String(d.Get("global_network_id").(string)),
	}

	if v, ok := d.GetOk("provider_name"); ok {
		input.Provider = aws.String(v.(string))
	}

	if v, ok := d.GetOk("site_id"); ok {
		input.SiteId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("type"); ok {
		input.Type = aws.String(v.(string))
	}

	output, err := FindLinks(ctx, conn, input)

	if err != nil {
		return diag.Errorf("error listing Network Manager Links: %s", err)
	}

	var linkIDs []string

	for _, v := range output {
		if len(tagsToMatch) > 0 {
			if !KeyValueTags(v.Tags).ContainsAll(tagsToMatch) {
				continue
			}
		}

		linkIDs = append(linkIDs, aws.StringValue(v.LinkId))
	}

	d.SetId(meta.(*conns.AWSClient).Region)
	d.Set("ids", linkIDs)

	return nil
}
