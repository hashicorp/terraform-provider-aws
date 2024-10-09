// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apigatewayv2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/apigatewayv2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/apigatewayv2/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_apigatewayv2_apis", name="APIs")
func dataSourceAPIs() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceAPIsRead,

		Schema: map[string]*schema.Schema{
			names.AttrIDs: {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Optional: true,
			},
			"protocol_type": {
				Type:     schema.TypeString,
				Optional: true,
			},
			names.AttrTags: tftags.TagsSchema(),
		},
	}
}

func dataSourceAPIsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayV2Client(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	tagsToMatch := tftags.New(ctx, d.Get(names.AttrTags).(map[string]interface{})).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	apis, err := findAPIs(ctx, conn, &apigatewayv2.GetApisInput{})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading API Gateway v2 APIs: %s", err)
	}

	var ids []*string

	for _, api := range apis {
		if v, ok := d.GetOk(names.AttrName); ok && v.(string) != aws.ToString(api.Name) {
			continue
		}

		if v, ok := d.GetOk("protocol_type"); ok && v.(string) != string(api.ProtocolType) {
			continue
		}

		if len(tagsToMatch) > 0 && !KeyValueTags(ctx, api.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).ContainsAll(tagsToMatch) {
			continue
		}

		ids = append(ids, api.ApiId)
	}

	d.SetId(meta.(*conns.AWSClient).Region)

	if err := d.Set(names.AttrIDs, flex.FlattenStringSet(ids)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting ids: %s", err)
	}

	return diags
}

func findAPIs(ctx context.Context, conn *apigatewayv2.Client, input *apigatewayv2.GetApisInput) ([]awstypes.Api, error) {
	var apis []awstypes.Api

	err := getAPIsPages(ctx, conn, input, func(page *apigatewayv2.GetApisOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		apis = append(apis, page.Items...)

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	return apis, nil
}
