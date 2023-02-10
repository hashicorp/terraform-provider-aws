package apigateway

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

func DataSourceAPIKey() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceAPIKeyRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"last_updated_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"value": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
			"tags": tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceAPIKeyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayConn()
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	id := d.Get("id").(string)
	apiKey, err := FindAPIKeyByID(ctx, conn, id)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading API Gateway API Key (%s): %s", id, err)
	}

	d.SetId(aws.StringValue(apiKey.Id))
	d.Set("name", apiKey.Name)
	d.Set("value", apiKey.Value)
	d.Set("created_date", aws.TimeValue(apiKey.CreatedDate).Format(time.RFC3339))
	d.Set("description", apiKey.Description)
	d.Set("enabled", apiKey.Enabled)
	d.Set("last_updated_date", aws.TimeValue(apiKey.LastUpdatedDate).Format(time.RFC3339))

	if err := d.Set("tags", KeyValueTags(apiKey.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	return diags
}
