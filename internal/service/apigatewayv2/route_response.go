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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKResource("aws_apigatewayv2_route_response", name="Route Response")
func resourceRouteResponse() *schema.Resource {
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

	input := &apigatewayv2.CreateRouteResponseInput{
		ApiId:            aws.String(d.Get("api_id").(string)),
		RouteId:          aws.String(d.Get("route_id").(string)),
		RouteResponseKey: aws.String(d.Get("route_response_key").(string)),
	}

	if v, ok := d.GetOk("model_selection_expression"); ok {
		input.ModelSelectionExpression = aws.String(v.(string))
	}

	if v, ok := d.GetOk("response_models"); ok {
		input.ResponseModels = flex.ExpandStringValueMap(v.(map[string]interface{}))
	}

	output, err := conn.CreateRouteResponse(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating API Gateway v2 Route Response: %s", err)
	}

	d.SetId(aws.ToString(output.RouteResponseId))

	return append(diags, resourceRouteResponseRead(ctx, d, meta)...)
}

func resourceRouteResponseRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayV2Client(ctx)

	output, err := findRouteResponseByThreePartKey(ctx, conn, d.Get("api_id").(string), d.Get("route_id").(string), d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] API Gateway v2 Route Response (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading API Gateway v2 Route Response (%s): %s", d.Id(), err)
	}

	d.Set("model_selection_expression", output.ModelSelectionExpression)
	d.Set("response_models", output.ResponseModels)
	d.Set("route_response_key", output.RouteResponseKey)

	return diags
}

func resourceRouteResponseUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayV2Client(ctx)

	input := &apigatewayv2.UpdateRouteResponseInput{
		ApiId:           aws.String(d.Get("api_id").(string)),
		RouteId:         aws.String(d.Get("route_id").(string)),
		RouteResponseId: aws.String(d.Id()),
	}

	if d.HasChange("model_selection_expression") {
		input.ModelSelectionExpression = aws.String(d.Get("model_selection_expression").(string))
	}

	if d.HasChange("response_models") {
		input.ResponseModels = flex.ExpandStringValueMap(d.Get("response_models").(map[string]interface{}))
	}

	if d.HasChange("route_response_key") {
		input.RouteResponseKey = aws.String(d.Get("route_response_key").(string))
	}

	_, err := conn.UpdateRouteResponse(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating API Gateway v2 Route Response (%s): %s", d.Id(), err)
	}

	return append(diags, resourceRouteResponseRead(ctx, d, meta)...)
}

func resourceRouteResponseDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayV2Client(ctx)

	log.Printf("[DEBUG] Deleting API Gateway v2 Route Response: %s", d.Id())
	_, err := conn.DeleteRouteResponse(ctx, &apigatewayv2.DeleteRouteResponseInput{
		ApiId:           aws.String(d.Get("api_id").(string)),
		RouteId:         aws.String(d.Get("route_id").(string)),
		RouteResponseId: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.NotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting API Gateway v2 Route Response (%s): %s", d.Id(), err)
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

func findRouteResponseByThreePartKey(ctx context.Context, conn *apigatewayv2.Client, apiID, routeID, routeResponseID string) (*apigatewayv2.GetRouteResponseOutput, error) {
	input := &apigatewayv2.GetRouteResponseInput{
		ApiId:           aws.String(apiID),
		RouteId:         aws.String(routeID),
		RouteResponseId: aws.String(routeResponseID),
	}

	return findRouteResponse(ctx, conn, input)
}

func findRouteResponse(ctx context.Context, conn *apigatewayv2.Client, input *apigatewayv2.GetRouteResponseInput) (*apigatewayv2.GetRouteResponseOutput, error) {
	output, err := conn.GetRouteResponse(ctx, input)

	if errs.IsA[*awstypes.NotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}
