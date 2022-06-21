package outposts

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/outposts"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func DataSourceOutpostAssets() *schema.Resource {
	return &schema.Resource{
		Read: DataSourceOutpostAssetsRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			"asset_ids": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func DataSourceOutpostAssetsRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).OutpostsConn
	outpost_id := aws.String(d.Get("arn").(string))

	input := &outposts.ListAssetsInput{
		OutpostIdentifier: outpost_id,
	}
	var asset_ids []string
	err := conn.ListAssetsPages(input, func(page *outposts.ListAssetsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}
		for _, asset := range page.Assets {
			if asset == nil {
				continue
			}
			asset_ids = append(asset_ids, aws.StringValue(asset.AssetId))
		}
		return !lastPage
	})

	if err != nil {
		return fmt.Errorf("error listing Outposts Assets: %w", err)
	}
	if len(asset_ids) == 0 {
		return fmt.Errorf("no Outposts Assets found matching criteria; try different search")
	}

	d.SetId(aws.StringValue(outpost_id))
	d.Set("asset_ids", asset_ids)

	return nil
}
