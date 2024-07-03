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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_apigatewayv2_route", name="Route")
func resourceRoute() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceRouteCreate,
		ReadWithoutTimeout:   resourceRouteRead,
		UpdateWithoutTimeout: resourceRouteUpdate,
		DeleteWithoutTimeout: resourceRouteDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceRouteImport,
		},

		Schema: map[string]*schema.Schema{
			"api_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"api_key_required": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"authorization_scopes": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"authorization_type": {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          awstypes.AuthorizationTypeNone,
				ValidateDiagFunc: enum.Validate[awstypes.AuthorizationType](),
			},
			"authorizer_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"model_selection_expression": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"operation_name": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 64),
			},
			"request_models": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"request_parameter": {
				Type:     schema.TypeSet,
				Optional: true,
				MinItems: 0,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"request_parameter_key": {
							Type:     schema.TypeString,
							Required: true,
						},
						"required": {
							Type:     schema.TypeBool,
							Required: true,
						},
					},
				},
			},
			"route_key": {
				Type:     schema.TypeString,
				Required: true,
			},
			"route_response_selection_expression": {
				Type:     schema.TypeString,
				Optional: true,
			},
			names.AttrTarget: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 128),
			},
		},
	}
}

func resourceRouteCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayV2Client(ctx)

	input := &apigatewayv2.CreateRouteInput{
		ApiId:             aws.String(d.Get("api_id").(string)),
		ApiKeyRequired:    aws.Bool(d.Get("api_key_required").(bool)),
		AuthorizationType: awstypes.AuthorizationType(d.Get("authorization_type").(string)),
		RouteKey:          aws.String(d.Get("route_key").(string)),
	}

	if v, ok := d.GetOk("authorization_scopes"); ok {
		input.AuthorizationScopes = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("authorizer_id"); ok {
		input.AuthorizerId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("model_selection_expression"); ok {
		input.ModelSelectionExpression = aws.String(v.(string))
	}

	if v, ok := d.GetOk("operation_name"); ok {
		input.OperationName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("request_models"); ok {
		input.RequestModels = flex.ExpandStringValueMap(v.(map[string]interface{}))
	}

	if v, ok := d.GetOk("request_parameter"); ok && v.(*schema.Set).Len() > 0 {
		input.RequestParameters = expandRouteRequestParameters(v.(*schema.Set).List())
	}

	if v, ok := d.GetOk("route_response_selection_expression"); ok {
		input.RouteResponseSelectionExpression = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrTarget); ok {
		input.Target = aws.String(v.(string))
	}

	output, err := conn.CreateRoute(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating API Gateway v2 Route: %s", err)
	}

	d.SetId(aws.ToString(output.RouteId))

	return append(diags, resourceRouteRead(ctx, d, meta)...)
}

func resourceRouteRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayV2Client(ctx)

	output, err := findRouteByTwoPartKey(ctx, conn, d.Get("api_id").(string), d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] API Gateway v2 Route (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading API Gateway v2 Route (%s): %s", d.Id(), err)
	}

	d.Set("api_key_required", output.ApiKeyRequired)
	d.Set("authorization_scopes", output.AuthorizationScopes)
	d.Set("authorization_type", output.AuthorizationType)
	d.Set("authorizer_id", output.AuthorizerId)
	d.Set("model_selection_expression", output.ModelSelectionExpression)
	d.Set("operation_name", output.OperationName)
	d.Set("request_models", output.RequestModels)
	if err := d.Set("request_parameter", flattenRouteRequestParameters(output.RequestParameters)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting request_parameter: %s", err)
	}
	d.Set("route_key", output.RouteKey)
	d.Set("route_response_selection_expression", output.RouteResponseSelectionExpression)
	d.Set(names.AttrTarget, output.Target)

	return diags
}

func resourceRouteUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayV2Client(ctx)

	var requestParameters map[string]awstypes.ParameterConstraints

	if d.HasChange("request_parameter") {
		o, n := d.GetChange("request_parameter")
		os := o.(*schema.Set)
		ns := n.(*schema.Set)

		for _, tfMapRaw := range os.Difference(ns).List() {
			tfMap, ok := tfMapRaw.(map[string]interface{})

			if !ok {
				continue
			}

			if v, ok := tfMap["request_parameter_key"].(string); ok && v != "" {
				input := &apigatewayv2.DeleteRouteRequestParameterInput{
					ApiId:               aws.String(d.Get("api_id").(string)),
					RequestParameterKey: aws.String(v),
					RouteId:             aws.String(d.Id()),
				}

				_, err := conn.DeleteRouteRequestParameter(ctx, input)

				if errs.IsA[*awstypes.NotFoundException](err) {
					continue
				}

				if err != nil {
					return sdkdiag.AppendErrorf(diags, "deleting API Gateway v2 Route (%s) request parameter (%s): %s", d.Id(), v, err)
				}
			}
		}

		requestParameters = expandRouteRequestParameters(ns.List())
	}

	if d.HasChangesExcept("request_parameter") || len(requestParameters) > 0 {
		input := &apigatewayv2.UpdateRouteInput{
			ApiId:   aws.String(d.Get("api_id").(string)),
			RouteId: aws.String(d.Id()),
		}

		if d.HasChange("api_key_required") {
			input.ApiKeyRequired = aws.Bool(d.Get("api_key_required").(bool))
		}

		if d.HasChange("authorization_scopes") {
			input.AuthorizationScopes = flex.ExpandStringValueSet(d.Get("authorization_scopes").(*schema.Set))
		}

		if d.HasChange("authorization_type") {
			input.AuthorizationType = awstypes.AuthorizationType(d.Get("authorization_type").(string))
		}

		if d.HasChange("authorizer_id") {
			input.AuthorizerId = aws.String(d.Get("authorizer_id").(string))
			input.AuthorizationType = awstypes.AuthorizationType(d.Get("authorization_type").(string))
		}

		if d.HasChange("model_selection_expression") {
			input.ModelSelectionExpression = aws.String(d.Get("model_selection_expression").(string))
		}

		if d.HasChange("operation_name") {
			input.OperationName = aws.String(d.Get("operation_name").(string))
		}

		if d.HasChange("request_models") {
			input.RequestModels = flex.ExpandStringValueMap(d.Get("request_models").(map[string]interface{}))
		}

		if d.HasChange("request_parameter") {
			input.RequestParameters = requestParameters
		}

		if d.HasChange("route_key") {
			input.RouteKey = aws.String(d.Get("route_key").(string))
		}

		if d.HasChange("route_response_selection_expression") {
			input.RouteResponseSelectionExpression = aws.String(d.Get("route_response_selection_expression").(string))
		}

		if d.HasChange(names.AttrTarget) {
			input.Target = aws.String(d.Get(names.AttrTarget).(string))
		}

		_, err := conn.UpdateRoute(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating API Gateway v2 Route (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceRouteRead(ctx, d, meta)...)
}

func resourceRouteDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayV2Client(ctx)

	log.Printf("[DEBUG] Deleting API Gateway v2 Route: %s", d.Id())
	_, err := conn.DeleteRoute(ctx, &apigatewayv2.DeleteRouteInput{
		ApiId:   aws.String(d.Get("api_id").(string)),
		RouteId: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.NotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting API Gateway v2 Route (%s): %s", d.Id(), err)
	}

	return diags
}

func resourceRouteImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	parts := strings.Split(d.Id(), "/")
	if len(parts) != 2 {
		return []*schema.ResourceData{}, fmt.Errorf("wrong format of import ID (%s), use: 'api-id/route-id'", d.Id())
	}

	apiID := parts[0]
	routeID := parts[1]

	conn := meta.(*conns.AWSClient).APIGatewayV2Client(ctx)

	output, err := findRouteByTwoPartKey(ctx, conn, apiID, routeID)

	if err != nil {
		return nil, err
	}

	if aws.ToBool(output.ApiGatewayManaged) {
		return nil, fmt.Errorf("API Gateway v2 Route (%s) was created via quick create", routeID)
	}

	d.SetId(routeID)
	d.Set("api_id", apiID)

	return []*schema.ResourceData{d}, nil
}

func findRouteByTwoPartKey(ctx context.Context, conn *apigatewayv2.Client, apiID, routeID string) (*apigatewayv2.GetRouteOutput, error) {
	input := &apigatewayv2.GetRouteInput{
		ApiId:   aws.String(apiID),
		RouteId: aws.String(routeID),
	}

	return findRoute(ctx, conn, input)
}

func findRoute(ctx context.Context, conn *apigatewayv2.Client, input *apigatewayv2.GetRouteInput) (*apigatewayv2.GetRouteOutput, error) {
	output, err := conn.GetRoute(ctx, input)

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

func expandRouteRequestParameters(tfList []interface{}) map[string]awstypes.ParameterConstraints {
	if len(tfList) == 0 {
		return nil
	}

	apiObjects := map[string]awstypes.ParameterConstraints{}

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := awstypes.ParameterConstraints{}

		if v, ok := tfMap["required"].(bool); ok {
			apiObject.Required = aws.Bool(v)
		}

		if v, ok := tfMap["request_parameter_key"].(string); ok && v != "" {
			apiObjects[v] = apiObject
		}
	}

	return apiObjects
}

func flattenRouteRequestParameters(apiObjects map[string]awstypes.ParameterConstraints) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for k, apiObject := range apiObjects {
		tfList = append(tfList, map[string]interface{}{
			"request_parameter_key": k,
			"required":              aws.ToBool(apiObject.Required),
		})
	}

	return tfList
}
