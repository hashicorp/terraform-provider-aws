// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apigatewayv2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/apigatewayv2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

// @SDKDataSource("aws_apigatewayv2_export", name="Export")
func dataSourceExport() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceExportRead,

		Schema: map[string]*schema.Schema{
			"api_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"body": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"export_version": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"1.0"}, false),
			},
			"include_extensions": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"specification": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"OAS30"}, false),
			},
			"stage_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"output_type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"JSON", "YAML"}, false),
			},
		},
	}
}

func dataSourceExportRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayV2Client(ctx)

	apiID := d.Get("api_id").(string)
	input := &apigatewayv2.ExportApiInput{
		ApiId:             aws.String(apiID),
		IncludeExtensions: aws.Bool(d.Get("include_extensions").(bool)),
		OutputType:        aws.String(d.Get("output_type").(string)),
		Specification:     aws.String(d.Get("specification").(string)),
	}

	if v, ok := d.GetOk("export_version"); ok {
		input.ExportVersion = aws.String(v.(string))
	}

	if v, ok := d.GetOk("stage_name"); ok {
		input.StageName = aws.String(v.(string))
	}

	export, err := conn.ExportApi(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "exporting Gateway v2 API (%s): %s", apiID, err)
	}

	d.SetId(apiID)
	d.Set("body", string(export.Body))

	return diags
}
