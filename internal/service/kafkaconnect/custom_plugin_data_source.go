package kafkaconnect

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kafkaconnect"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func DataSourceCustomPlugin() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceCustomPluginRead,

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

func dataSourceCustomPluginRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).KafkaConnectConn()

	name := d.Get("name")
	var output []*kafkaconnect.CustomPluginSummary

	err := conn.ListCustomPluginsPagesWithContext(ctx, &kafkaconnect.ListCustomPluginsInput{}, func(page *kafkaconnect.ListCustomPluginsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.CustomPlugins {
			if aws.StringValue(v.Name) == name {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if err != nil {
		return diag.Errorf("error listing MSK Connect Custom Plugins: %s", err)
	}

	if len(output) == 0 || output[0] == nil {
		err = tfresource.NewEmptyResultError(name)
	} else if count := len(output); count > 1 {
		err = tfresource.NewTooManyResultsError(count, name)
	}

	if err != nil {
		return diag.FromErr(tfresource.SingularDataSourceFindError("MSK Connect Custom Plugin", err))
	}

	plugin := output[0]

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
