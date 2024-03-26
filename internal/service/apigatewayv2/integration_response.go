// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apigatewayv2

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/apigatewayv2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/apigatewayv2/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
)

// @SDKResource("aws_apigatewayv2_integration_response")
func ResourceIntegrationResponse() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceIntegrationResponseCreate,
		ReadWithoutTimeout:   resourceIntegrationResponseRead,
		UpdateWithoutTimeout: resourceIntegrationResponseUpdate,
		DeleteWithoutTimeout: resourceIntegrationResponseDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceIntegrationResponseImport,
		},

		Schema: map[string]*schema.Schema{
			"api_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"content_handling_strategy": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(enum.Slice(awstypes.ContentHandlingStrategyConvertToBinary, awstypes.ContentHandlingStrategyConvertToText), false),
			},
			"integration_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"integration_response_key": {
				Type:     schema.TypeString,
				Required: true,
			},
			"response_templates": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"template_selection_expression": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func resourceIntegrationResponseCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayV2Client(ctx)

	req := &apigatewayv2.CreateIntegrationResponseInput{
		ApiId:                  aws.String(d.Get("api_id").(string)),
		IntegrationId:          aws.String(d.Get("integration_id").(string)),
		IntegrationResponseKey: aws.String(d.Get("integration_response_key").(string)),
	}
	if v, ok := d.GetOk("content_handling_strategy"); ok {
		req.ContentHandlingStrategy = awstypes.ContentHandlingStrategy(v.(string))
	}
	if v, ok := d.GetOk("response_templates"); ok {
		req.ResponseTemplates = flex.ExpandStringValueMap(v.(map[string]interface{}))
	}
	if v, ok := d.GetOk("template_selection_expression"); ok {
		req.TemplateSelectionExpression = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating API Gateway v2 integration response: %+v", req)
	resp, err := conn.CreateIntegrationResponse(ctx, req)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating API Gateway v2 integration response: %s", err)
	}

	d.SetId(aws.ToString(resp.IntegrationResponseId))

	return append(diags, resourceIntegrationResponseRead(ctx, d, meta)...)
}

func resourceIntegrationResponseRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayV2Client(ctx)

	resp, err := conn.GetIntegrationResponse(ctx, &apigatewayv2.GetIntegrationResponseInput{
		ApiId:                 aws.String(d.Get("api_id").(string)),
		IntegrationId:         aws.String(d.Get("integration_id").(string)),
		IntegrationResponseId: aws.String(d.Id()),
	})
	if errs.IsA[*awstypes.NotFoundException](err) && !d.IsNewResource() {
		log.Printf("[WARN] API Gateway v2 integration response (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading API Gateway v2 integration response: %s", err)
	}

	d.Set("content_handling_strategy", resp.ContentHandlingStrategy)
	d.Set("integration_response_key", resp.IntegrationResponseKey)
	err = d.Set("response_templates", resp.ResponseTemplates)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "setting response_templates: %s", err)
	}
	d.Set("template_selection_expression", resp.TemplateSelectionExpression)

	return diags
}

func resourceIntegrationResponseUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayV2Client(ctx)

	req := &apigatewayv2.UpdateIntegrationResponseInput{
		ApiId:                 aws.String(d.Get("api_id").(string)),
		IntegrationId:         aws.String(d.Get("integration_id").(string)),
		IntegrationResponseId: aws.String(d.Id()),
	}
	if d.HasChange("content_handling_strategy") {
		req.ContentHandlingStrategy = awstypes.ContentHandlingStrategy(d.Get("content_handling_strategy").(string))
	}
	if d.HasChange("integration_response_key") {
		req.IntegrationResponseKey = aws.String(d.Get("integration_response_key").(string))
	}
	if d.HasChange("response_templates") {
		req.ResponseTemplates = flex.ExpandStringValueMap(d.Get("response_templates").(map[string]interface{}))
	}
	if d.HasChange("template_selection_expression") {
		req.TemplateSelectionExpression = aws.String(d.Get("template_selection_expression").(string))
	}

	log.Printf("[DEBUG] Updating API Gateway v2 integration response: %+v", req)
	_, err := conn.UpdateIntegrationResponse(ctx, req)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating API Gateway v2 integration response: %s", err)
	}

	return append(diags, resourceIntegrationResponseRead(ctx, d, meta)...)
}

func resourceIntegrationResponseDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayV2Client(ctx)

	log.Printf("[DEBUG] Deleting API Gateway v2 integration response (%s)", d.Id())
	_, err := conn.DeleteIntegrationResponse(ctx, &apigatewayv2.DeleteIntegrationResponseInput{
		ApiId:                 aws.String(d.Get("api_id").(string)),
		IntegrationId:         aws.String(d.Get("integration_id").(string)),
		IntegrationResponseId: aws.String(d.Id()),
	})
	if errs.IsA[*awstypes.NotFoundException](err) {
		return diags
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting API Gateway v2 integration response: %s", err)
	}

	return diags
}

func resourceIntegrationResponseImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	parts := strings.Split(d.Id(), "/")
	if len(parts) != 3 {
		return []*schema.ResourceData{}, fmt.Errorf("wrong format of import ID (%s), use: 'api-id/integration-id/integration-response-id'", d.Id())
	}

	d.SetId(parts[2])
	d.Set("api_id", parts[0])
	d.Set("integration_id", parts[1])

	return []*schema.ResourceData{d}, nil
}
