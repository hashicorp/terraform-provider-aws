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
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
)

// @SDKResource("aws_apigatewayv2_route_response")
func ResourceRouteResponse() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceRouteResponseCreate,
		ReadWithoutTimeout:   resourceRouteResponseRead,
		UpdateWithoutTimeout: resourceRouteResponseUpdate,
		DeleteWithoutTimeout: resourceRouteResponseDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceRouteResponseImport,
		},

		Schema: map[string]*schema.Schema{
			"api_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"model_selection_expression": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"response_models": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"route_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"route_response_key": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceRouteResponseCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayV2Client(ctx)

	req := &apigatewayv2.CreateRouteResponseInput{
		ApiId:            aws.String(d.Get("api_id").(string)),
		RouteId:          aws.String(d.Get("route_id").(string)),
		RouteResponseKey: aws.String(d.Get("route_response_key").(string)),
	}
	if v, ok := d.GetOk("model_selection_expression"); ok {
		req.ModelSelectionExpression = aws.String(v.(string))
	}
	if v, ok := d.GetOk("response_models"); ok {
		req.ResponseModels = flex.ExpandStringValueMap(v.(map[string]interface{}))
	}

	log.Printf("[DEBUG] Creating API Gateway v2 route response: %+v", req)
	resp, err := conn.CreateRouteResponse(ctx, req)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating API Gateway v2 route response: %s", err)
	}

	d.SetId(aws.ToString(resp.RouteResponseId))

	return append(diags, resourceRouteResponseRead(ctx, d, meta)...)
}

func resourceRouteResponseRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayV2Client(ctx)

	resp, err := conn.GetRouteResponse(ctx, &apigatewayv2.GetRouteResponseInput{
		ApiId:           aws.String(d.Get("api_id").(string)),
		RouteId:         aws.String(d.Get("route_id").(string)),
		RouteResponseId: aws.String(d.Id()),
	})
	if errs.IsA[*awstypes.NotFoundException](err) && !d.IsNewResource() {
		log.Printf("[WARN] API Gateway v2 route response (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading API Gateway v2 route response: %s", err)
	}

	d.Set("model_selection_expression", resp.ModelSelectionExpression)
	if err := d.Set("response_models", resp.ResponseModels); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting response_models: %s", err)
	}
	d.Set("route_response_key", resp.RouteResponseKey)

	return diags
}

func resourceRouteResponseUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayV2Client(ctx)

	req := &apigatewayv2.UpdateRouteResponseInput{
		ApiId:           aws.String(d.Get("api_id").(string)),
		RouteId:         aws.String(d.Get("route_id").(string)),
		RouteResponseId: aws.String(d.Id()),
	}
	if d.HasChange("model_selection_expression") {
		req.ModelSelectionExpression = aws.String(d.Get("model_selection_expression").(string))
	}
	if d.HasChange("response_models") {
		req.ResponseModels = flex.ExpandStringValueMap(d.Get("response_models").(map[string]interface{}))
	}
	if d.HasChange("route_response_key") {
		req.RouteResponseKey = aws.String(d.Get("route_response_key").(string))
	}

	log.Printf("[DEBUG] Updating API Gateway v2 route response: %+v", req)
	_, err := conn.UpdateRouteResponse(ctx, req)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating API Gateway v2 route response: %s", err)
	}

	return append(diags, resourceRouteResponseRead(ctx, d, meta)...)
}

func resourceRouteResponseDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayV2Client(ctx)

	log.Printf("[DEBUG] Deleting API Gateway v2 route response (%s)", d.Id())
	_, err := conn.DeleteRouteResponse(ctx, &apigatewayv2.DeleteRouteResponseInput{
		ApiId:           aws.String(d.Get("api_id").(string)),
		RouteId:         aws.String(d.Get("route_id").(string)),
		RouteResponseId: aws.String(d.Id()),
	})
	if errs.IsA[*awstypes.NotFoundException](err) {
		return diags
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting API Gateway v2 route response: %s", err)
	}

	return diags
}

func resourceRouteResponseImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	parts := strings.Split(d.Id(), "/")
	if len(parts) != 3 {
		return []*schema.ResourceData{}, fmt.Errorf("wrong format of import ID (%s), use: 'api-id/route-id/route-response-id'", d.Id())
	}

	d.SetId(parts[2])
	d.Set("api_id", parts[0])
	d.Set("route_id", parts[1])

	return []*schema.ResourceData{d}, nil
}
