package codestarconnections

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codestarconnections"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func DataSourceConnection() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceConnectionRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: verify.ValidARN,
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
				Optional: true,
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

	log.Printf("[DEBUG] Getting CodeStar Connection")

	var connection *codestarconnections.Connection
	var err error

	if v, ok := d.GetOk("arn"); ok {
		arn := v.(string)
		connection, err = findConnectionByARN(conn, arn)
		if err != nil {
			return fmt.Errorf("error getting CodeStar Connection (%s): %w", arn, err)
		}
		if connection == nil {
			return fmt.Errorf("Could not find CodeStar connection with arn (%s)", arn)
		}
	} else if v, ok := d.GetOk("name"); ok {
		name := v.(string)
		connection, err = findConnectionByName(conn, name)
		if err != nil {
			return fmt.Errorf("error getting CodeStar Connection (%s): %w", name, err)
		}
		if connection == nil {
			return fmt.Errorf("Could not find CodeStar connection with name (%s)", name)
		}
	} else {
		return fmt.Errorf("Either arn or name must be specified")
	}

	log.Printf("[DEBUG] CodeStar Connection: %#v", connection)

	arn := aws.StringValue(connection.ConnectionArn)
	d.SetId(arn)
	d.Set("arn", arn)
	d.Set("connection_status", connection.ConnectionStatus)
	d.Set("host_arn", connection.HostArn)
	d.Set("name", connection.ConnectionName)
	d.Set("provider_type", connection.ProviderType)

	tags, err := ListTags(conn, arn)
	if err != nil {
		return fmt.Errorf("error listing tags for CodeStar Connection (%s): %w", arn, err)
	}

	if err := d.Set("tags", tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags for CodeStar Connection (%s): %w", arn, err)
	}

	return nil
}
