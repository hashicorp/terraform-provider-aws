package codestarconnections

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codestarconnections"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func DataSourceConnection() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceConnectionRead,

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

func dataSourceConnectionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CodeStarConnectionsConn()
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	var connection *codestarconnections.Connection
	var err error

	if v, ok := d.GetOk("arn"); ok {
		arn := v.(string)
		connection, err = FindConnectionByARN(ctx, conn, arn)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading CodeStar Connections Connection (%s): %s", arn, err)
		}
	} else if v, ok := d.GetOk("name"); ok {
		name := v.(string)

		err = conn.ListConnectionsPagesWithContext(ctx, &codestarconnections.ListConnectionsInput{}, func(page *codestarconnections.ListConnectionsOutput, lastPage bool) bool {
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
			return sdkdiag.AppendErrorf(diags, "listing CodeStar Connections Connections: %s", err)
		}

		if connection == nil {
			return sdkdiag.AppendErrorf(diags, "CodeStar Connections Connection (%s): not found", name)
		}
	}

	arn := aws.StringValue(connection.ConnectionArn)
	d.SetId(arn)
	d.Set("arn", arn)
	d.Set("connection_status", connection.ConnectionStatus)
	d.Set("host_arn", connection.HostArn)
	d.Set("name", connection.ConnectionName)
	d.Set("provider_type", connection.ProviderType)

	tags, err := ListTags(ctx, conn, arn)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing tags for CodeStar Connections Connection (%s): %s", arn, err)
	}

	if err := d.Set("tags", tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	return diags
}
