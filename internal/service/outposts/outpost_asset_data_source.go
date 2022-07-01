package outposts

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/outposts"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func DataSourceOutpostAsset() *schema.Resource {
	return &schema.Resource{
		Read: DataSourceOutpostAssetRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			"asset_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"asset_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"host_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"rack_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func DataSourceOutpostAssetRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).OutpostsConn
	outpost_id := aws.String(d.Get("arn").(string))

	input := &outposts.ListAssetsInput{
		OutpostIdentifier: outpost_id,
	}

	var results []*outposts.AssetInfo
	err := conn.ListAssetsPages(input, func(page *outposts.ListAssetsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}
		for _, asset := range page.Assets {
			if asset == nil {
				continue
			}
			if v, ok := d.GetOk("asset_id"); ok && v.(string) != aws.StringValue(asset.AssetId) {
				continue
			}
			results = append(results, asset)
		}
		return !lastPage
	})

	if err != nil {
		return fmt.Errorf("error listing Outposts Asset: %w", err)
	}
	if len(results) == 0 {
		return fmt.Errorf("no Outposts Asset found matching criteria; try different search")
	}

	asset := results[0]

	d.SetId(aws.StringValue(outpost_id))
	d.Set("asset_id", asset.AssetId)
	d.Set("asset_type", asset.AssetType)
	d.Set("host_id", asset.ComputeAttributes.HostId)
	d.Set("rack_id", asset.RackId)
	return nil
}
