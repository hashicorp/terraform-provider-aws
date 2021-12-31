package kafkaconnect

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kafkaconnect"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func DataSourceCustomPlugin() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceCustomPluginRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"latest_revision": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"state": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceCustomPluginRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).KafkaConnectConn

	pluginName := d.Get("name")

	input := &kafkaconnect.ListCustomPluginsInput{}

	var plugin *kafkaconnect.CustomPluginSummary

	err := conn.ListCustomPluginsPages(input, func(page *kafkaconnect.ListCustomPluginsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, pluginSummary := range page.CustomPlugins {
			if aws.StringValue(pluginSummary.Name) == pluginName {
				plugin = pluginSummary

				return false
			}
		}

		return !lastPage
	})

	if err != nil {
		return fmt.Errorf("error listing MSK Connect Custom Plugins: %w", err)
	}

	if plugin == nil {
		return fmt.Errorf("error reading MSK Connect Custom Plugin (%s): no results found", pluginName)
	}

	d.SetId(aws.StringValue(plugin.CustomPluginArn))
	d.Set("arn", plugin.CustomPluginArn)
	d.Set("description", plugin.Description)
	d.Set("name", plugin.Name)
	d.Set("state", plugin.CustomPluginState)
	if plugin.LatestRevision != nil {
		d.Set("latest_revision", plugin.LatestRevision.Revision)
	} else {
		d.Set("latest_revision", nil)
	}

	return nil
}
