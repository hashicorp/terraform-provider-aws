package codestarconnections

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

func DataSourceConnection() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceConnectionRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Required: true,
			},
			"connection_status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"host_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"provider_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceConnectionRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CodeStarConnectionsConn
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	arn := d.Get("arn").(string)
	connection, err := FindConnectionByARN(conn, arn)

	if err != nil {
		return fmt.Errorf("reading CodeStar Connections Connection (%s): %w", arn, err)
	}

	d.SetId(arn)
	d.Set("connection_status", connection.ConnectionStatus)
	d.Set("host_arn", connection.HostArn)
	d.Set("name", connection.ConnectionName)
	d.Set("provider_type", connection.ProviderType)

	tags, err := ListTags(conn, arn)

	if err != nil {
		return fmt.Errorf("listing tags for CodeStar Connections Connection (%s): %w", arn, err)
	}

	if err := d.Set("tags", tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("setting tags: %w", err)
	}

	return nil
}
