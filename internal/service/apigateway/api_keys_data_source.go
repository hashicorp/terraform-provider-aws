// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apigateway

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/apigateway"
	apigwtypes "github.com/aws/aws-sdk-go-v2/service/apigateway/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_api_gateway_api_keys", name="API Keys")
func dataSourceAPIKeys() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceAPIKeysRead,

		Schema: map[string]*schema.Schema{
			"customer_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"include_values": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"items": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrCreatedDate: {
							Type:     schema.TypeString,
							Computed: true,
						},
						"customer_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrDescription: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrEnabled: {
							Type:     schema.TypeBool,
							Computed: true,
						},
						names.AttrID: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrLastUpdatedDate: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrName: {
							Type:     schema.TypeString,
							Computed: true,
						},
						"stage_keys": {
							Type:     schema.TypeList,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						names.AttrTags: tftags.TagsSchema(),
						names.AttrValue: {
							Type:      schema.TypeString,
							Computed:  true,
							Sensitive: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceAPIKeysRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayClient(ctx)

	input := &apigateway.GetApiKeysInput{}

	if v, ok := d.GetOk("customer_id"); ok {
		input.CustomerId = aws.String(v.(string))
	}

	input.IncludeValues = aws.Bool(d.Get("include_values").(bool))

	var apiKeyItems []map[string]interface{}

	pages := apigateway.NewGetApiKeysPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "getting keys: %s", err)
		}

		apiKeyItems = append(apiKeyItems, readAPIKeyItems(page.Items)...)
	}

	d.SetId(meta.(*conns.AWSClient).Region)
	if err := d.Set("items", apiKeyItems); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting apiKeyItems: %s", err)
	}

	return diags
}

func readAPIKeyItems(apiKeyItems []apigwtypes.ApiKey) []map[string]interface{} {
	apiKeys := make([]map[string]interface{}, 0, len(apiKeyItems))
	for _, apiKey := range apiKeyItems {
		k := make(map[string]interface{})

		k[names.AttrCreatedDate] = aws.ToTime(apiKey.CreatedDate).Format(time.RFC3339)
		k["customer_id"] = aws.ToString(apiKey.CustomerId)
		k[names.AttrDescription] = aws.ToString(apiKey.Description)
		k[names.AttrEnabled] = aws.ToBool(&apiKey.Enabled)
		k[names.AttrID] = aws.ToString(apiKey.Id)
		k[names.AttrLastUpdatedDate] = aws.ToTime(apiKey.LastUpdatedDate).Format(time.RFC3339)
		k[names.AttrName] = aws.ToString(apiKey.Name)
		k["stage_keys"] = apiKey.StageKeys
		k[names.AttrValue] = aws.ToString(apiKey.Value)
		k[names.AttrTags] = apiKey.Tags

		apiKeys = append(apiKeys, k)
	}

	return apiKeys
}
