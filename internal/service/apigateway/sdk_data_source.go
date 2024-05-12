// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apigateway

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/apigateway"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_api_gateway_sdk", name="SDK")
func dataSourceSDK() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceSDKRead,

		Schema: map[string]*schema.Schema{
			"body": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrContentType: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"content_disposition": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrParameters: {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"rest_api_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"sdk_type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"java", "javascript", "android", "objectivec", "swift", "ruby"}, false),
			},
			"stage_name": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func dataSourceSDKRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayClient(ctx)

	apiID := d.Get("rest_api_id").(string)
	stageName := d.Get("stage_name").(string)
	sdkType := d.Get("sdk_type").(string)
	input := &apigateway.GetSdkInput{
		RestApiId: aws.String(apiID),
		SdkType:   aws.String(sdkType),
		StageName: aws.String(stageName),
	}

	if v, ok := d.GetOk(names.AttrParameters); ok && len(v.(map[string]interface{})) > 0 {
		input.Parameters = flex.ExpandStringValueMap(v.(map[string]interface{}))
	}

	id := apiID + ":" + stageName + ":" + sdkType

	sdk, err := conn.GetSdk(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading API Gateway SDK (%s): %s", id, err)
	}

	d.SetId(id)
	d.Set("body", string(sdk.Body))
	d.Set("content_disposition", sdk.ContentDisposition)
	d.Set(names.AttrContentType, sdk.ContentType)

	return diags
}
