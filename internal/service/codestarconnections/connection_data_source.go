package codestarconnections

import (
	"fmt"

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
				ExactlyOneOf: []string{"arn", "name"},
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
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{"arn", "name"},
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

	var connection *codestarconnections.Connection
	var err error

	if v, ok := d.GetOk("arn"); ok {
		arn := v.(string)
		connection, err = FindConnectionByARN(conn, arn)

		if err != nil {
			return fmt.Errorf("reading CodeStar Connections Connection (%s): %w", arn, err)
		}
	} else if v, ok := d.GetOk("name"); ok {
		name := v.(string)

		err = conn.ListConnectionsPages(&codestarconnections.ListConnectionsInput{}, func(page *codestarconnections.ListConnectionsOutput, lastPage bool) bool {
			if page == nil {
				return !lastPage
			}

			for _, v := range page.Connections {
				if aws.StringValue(v.ConnectionName) == name {
					connection = v

					return false
				}
			}

			return !lastPage
		})

		if err != nil {
			return fmt.Errorf("listing CodeStar Connections Connections: %w", err)
		}

		if connection == nil {
			return fmt.Errorf("CodeStar Connections Connection (%s): not found", name)
		}
	}

	arn := aws.StringValue(connection.ConnectionArn)
	d.SetId(arn)
	d.Set("arn", arn)
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
