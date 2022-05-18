package apigatewayv2

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigatewayv2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
)

func ResourceRoute() *schema.Resource {
	return &schema.Resource{
		Create: resourceRouteCreate,
		Read:   resourceRouteRead,
		Update: resourceRouteUpdate,
		Delete: resourceRouteDelete,
		Importer: &schema.ResourceImporter{
			State: resourceRouteImport,
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
				Type:         schema.TypeString,
				Optional:     true,
				Default:      apigatewayv2.AuthorizationTypeNone,
				ValidateFunc: validation.StringInSlice(apigatewayv2.AuthorizationType_Values(), false),
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
			"target": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 128),
			},
		},
	}
}

func resourceRouteCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).APIGatewayV2Conn

	req := &apigatewayv2.CreateRouteInput{
		ApiId:             aws.String(d.Get("api_id").(string)),
		ApiKeyRequired:    aws.Bool(d.Get("api_key_required").(bool)),
		AuthorizationType: aws.String(d.Get("authorization_type").(string)),
		RouteKey:          aws.String(d.Get("route_key").(string)),
	}
	if v, ok := d.GetOk("authorization_scopes"); ok {
		req.AuthorizationScopes = flex.ExpandStringSet(v.(*schema.Set))
	}
	if v, ok := d.GetOk("authorizer_id"); ok {
		req.AuthorizerId = aws.String(v.(string))
	}
	if v, ok := d.GetOk("model_selection_expression"); ok {
		req.ModelSelectionExpression = aws.String(v.(string))
	}
	if v, ok := d.GetOk("operation_name"); ok {
		req.OperationName = aws.String(v.(string))
	}
	if v, ok := d.GetOk("request_models"); ok {
		req.RequestModels = flex.ExpandStringMap(v.(map[string]interface{}))
	}
	if v, ok := d.GetOk("request_parameter"); ok && v.(*schema.Set).Len() > 0 {
		req.RequestParameters = expandRouteRequestParameters(v.(*schema.Set).List())
	}
	if v, ok := d.GetOk("route_response_selection_expression"); ok {
		req.RouteResponseSelectionExpression = aws.String(v.(string))
	}
	if v, ok := d.GetOk("target"); ok {
		req.Target = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating API Gateway v2 route: %s", req)
	resp, err := conn.CreateRoute(req)
	if err != nil {
		return fmt.Errorf("error creating API Gateway v2 route: %w", err)
	}

	d.SetId(aws.StringValue(resp.RouteId))

	return resourceRouteRead(d, meta)
}

func resourceRouteRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).APIGatewayV2Conn

	resp, err := conn.GetRoute(&apigatewayv2.GetRouteInput{
		ApiId:   aws.String(d.Get("api_id").(string)),
		RouteId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, apigatewayv2.ErrCodeNotFoundException) && !d.IsNewResource() {
		log.Printf("[WARN] API Gateway v2 route (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading API Gateway v2 route (%s): %w", d.Id(), err)
	}

	d.Set("api_key_required", resp.ApiKeyRequired)
	if err := d.Set("authorization_scopes", flex.FlattenStringSet(resp.AuthorizationScopes)); err != nil {
		return fmt.Errorf("error setting authorization_scopes: %w", err)
	}
	d.Set("authorization_type", resp.AuthorizationType)
	d.Set("authorizer_id", resp.AuthorizerId)
	d.Set("model_selection_expression", resp.ModelSelectionExpression)
	d.Set("operation_name", resp.OperationName)
	if err := d.Set("request_models", flex.PointersMapToStringList(resp.RequestModels)); err != nil {
		return fmt.Errorf("error setting request_models: %w", err)
	}
	if err := d.Set("request_parameter", flattenRouteRequestParameters(resp.RequestParameters)); err != nil {
		return fmt.Errorf("error setting request_parameter: %w", err)
	}
	d.Set("route_key", resp.RouteKey)
	d.Set("route_response_selection_expression", resp.RouteResponseSelectionExpression)
	d.Set("target", resp.Target)

	return nil
}

func resourceRouteUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).APIGatewayV2Conn

	var requestParameters map[string]*apigatewayv2.ParameterConstraints

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
				log.Printf("[DEBUG] Deleting API Gateway v2 route (%s) request parameter (%s)", d.Id(), v)
				_, err := conn.DeleteRouteRequestParameter(&apigatewayv2.DeleteRouteRequestParameterInput{
					ApiId:               aws.String(d.Get("api_id").(string)),
					RequestParameterKey: aws.String(v),
					RouteId:             aws.String(d.Id()),
				})

				if tfawserr.ErrCodeEquals(err, apigatewayv2.ErrCodeNotFoundException) {
					continue
				}

				if err != nil {
					return fmt.Errorf("error deleting API Gateway v2 route (%s) request parameter (%s): %w", d.Id(), v, err)
				}
			}
		}

		requestParameters = expandRouteRequestParameters(ns.List())
	}

	if d.HasChangesExcept("request_parameter") || len(requestParameters) > 0 {
		req := &apigatewayv2.UpdateRouteInput{
			ApiId:   aws.String(d.Get("api_id").(string)),
			RouteId: aws.String(d.Id()),
		}
		if d.HasChange("api_key_required") {
			req.ApiKeyRequired = aws.Bool(d.Get("api_key_required").(bool))
		}
		if d.HasChange("authorization_scopes") {
			req.AuthorizationScopes = flex.ExpandStringSet(d.Get("authorization_scopes").(*schema.Set))
		}
		if d.HasChange("authorization_type") {
			req.AuthorizationType = aws.String(d.Get("authorization_type").(string))
		}
		if d.HasChange("authorizer_id") {
			req.AuthorizerId = aws.String(d.Get("authorizer_id").(string))
		}
		if d.HasChange("model_selection_expression") {
			req.ModelSelectionExpression = aws.String(d.Get("model_selection_expression").(string))
		}
		if d.HasChange("operation_name") {
			req.OperationName = aws.String(d.Get("operation_name").(string))
		}
		if d.HasChange("request_models") {
			req.RequestModels = flex.ExpandStringMap(d.Get("request_models").(map[string]interface{}))
		}
		if d.HasChange("request_parameter") {
			req.RequestParameters = requestParameters
		}
		if d.HasChange("route_key") {
			req.RouteKey = aws.String(d.Get("route_key").(string))
		}
		if d.HasChange("route_response_selection_expression") {
			req.RouteResponseSelectionExpression = aws.String(d.Get("route_response_selection_expression").(string))
		}
		if d.HasChange("target") {
			req.Target = aws.String(d.Get("target").(string))
		}

		log.Printf("[DEBUG] Updating API Gateway v2 route: %s", req)
		_, err := conn.UpdateRoute(req)

		if err != nil {
			return fmt.Errorf("error updating API Gateway v2 route (%s): %w", d.Id(), err)
		}
	}

	return resourceRouteRead(d, meta)
}

func resourceRouteDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).APIGatewayV2Conn

	log.Printf("[DEBUG] Deleting API Gateway v2 route (%s)", d.Id())
	_, err := conn.DeleteRoute(&apigatewayv2.DeleteRouteInput{
		ApiId:   aws.String(d.Get("api_id").(string)),
		RouteId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, apigatewayv2.ErrCodeNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting API Gateway v2 route (%s): %w", d.Id(), err)
	}

	return nil
}

func resourceRouteImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	parts := strings.Split(d.Id(), "/")
	if len(parts) != 2 {
		return []*schema.ResourceData{}, fmt.Errorf("Wrong format of resource: %s. Please follow 'api-id/route-id'", d.Id())
	}

	apiId := parts[0]
	routeId := parts[1]

	conn := meta.(*conns.AWSClient).APIGatewayV2Conn

	resp, err := conn.GetRoute(&apigatewayv2.GetRouteInput{
		ApiId:   aws.String(apiId),
		RouteId: aws.String(routeId),
	})
	if err != nil {
		return nil, err
	}

	if aws.BoolValue(resp.ApiGatewayManaged) {
		return nil, fmt.Errorf("API Gateway v2 route (%s) was created via quick create", routeId)
	}

	d.SetId(routeId)
	d.Set("api_id", apiId)

	return []*schema.ResourceData{d}, nil
}

func expandRouteRequestParameters(tfList []interface{}) map[string]*apigatewayv2.ParameterConstraints {
	if len(tfList) == 0 {
		return nil
	}

	apiObjects := map[string]*apigatewayv2.ParameterConstraints{}

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := &apigatewayv2.ParameterConstraints{}

		if v, ok := tfMap["required"].(bool); ok {
			apiObject.Required = aws.Bool(v)
		}

		if v, ok := tfMap["request_parameter_key"].(string); ok && v != "" {
			apiObjects[v] = apiObject
		}
	}

	return apiObjects
}

func flattenRouteRequestParameters(apiObjects map[string]*apigatewayv2.ParameterConstraints) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for k, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, map[string]interface{}{
			"request_parameter_key": k,
			"required":              aws.BoolValue(apiObject.Required),
		})
	}

	return tfList
}
