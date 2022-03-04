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

func DataSourceDevices() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceDevicesRead,

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

func dataSourceDevicesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).NetworkManagerConn
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig
	tagsToMatch := tftags.New(d.Get("tags").(map[string]interface{})).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	output, err := FindDevices(ctx, conn, &networkmanager.GetDevicesInput{
		GlobalNetworkId: aws.String(d.Get("global_network_id").(string)),
	})

	if err != nil {
		return diag.Errorf("error listing Network Manager Devices: %s", err)
	}

	var deviceIDs []string

	for _, v := range output {
		if len(tagsToMatch) > 0 {
			if !KeyValueTags(v.Tags).ContainsAll(tagsToMatch) {
				continue
			}
		}

		deviceIDs = append(deviceIDs, aws.StringValue(v.DeviceId))
	}

	d.SetId(meta.(*conns.AWSClient).Region)
	d.Set("ids", deviceIDs)

	return nil
}
