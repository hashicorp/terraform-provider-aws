package apigateway

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
)

func ResourceMethod() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceMethodCreate,
		ReadWithoutTimeout:   resourceMethodRead,
		UpdateWithoutTimeout: resourceMethodUpdate,
		DeleteWithoutTimeout: resourceMethodDelete,
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				idParts := strings.Split(d.Id(), "/")
				if len(idParts) != 3 || idParts[0] == "" || idParts[1] == "" || idParts[2] == "" {
					return nil, fmt.Errorf("Unexpected format of ID (%q), expected REST-API-ID/RESOURCE-ID/HTTP-METHOD", d.Id())
				}
				restApiID := idParts[0]
				resourceID := idParts[1]
				httpMethod := idParts[2]
				d.Set("http_method", httpMethod)
				d.Set("resource_id", resourceID)
				d.Set("rest_api_id", restApiID)
				d.SetId(fmt.Sprintf("agm-%s-%s-%s", restApiID, resourceID, httpMethod))
				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"rest_api_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"resource_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"http_method": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validHTTPMethod(),
			},

			"authorization": {
				Type:     schema.TypeString,
				Required: true,
			},

			"authorizer_id": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"authorization_scopes": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
				Optional: true,
			},

			"api_key_required": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			"request_models": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			"request_parameters": {
				Type:     schema.TypeMap,
				Elem:     &schema.Schema{Type: schema.TypeBool},
				Optional: true,
			},

			"request_validator_id": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"operation_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func resourceMethodCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayConn()

	input := apigateway.PutMethodInput{
		AuthorizationType: aws.String(d.Get("authorization").(string)),
		HttpMethod:        aws.String(d.Get("http_method").(string)),
		ResourceId:        aws.String(d.Get("resource_id").(string)),
		RestApiId:         aws.String(d.Get("rest_api_id").(string)),
		ApiKeyRequired:    aws.Bool(d.Get("api_key_required").(bool)),
	}

	models := make(map[string]string)
	for k, v := range d.Get("request_models").(map[string]interface{}) {
		models[k] = v.(string)
	}
	if len(models) > 0 {
		input.RequestModels = aws.StringMap(models)
	}

	parameters := make(map[string]bool)
	if kv, ok := d.GetOk("request_parameters"); ok {
		for k, v := range kv.(map[string]interface{}) {
			parameters[k], ok = v.(bool)
			if !ok {
				value, _ := strconv.ParseBool(v.(string))
				parameters[k] = value
			}
		}
		input.RequestParameters = aws.BoolMap(parameters)
	}

	if v, ok := d.GetOk("authorizer_id"); ok {
		input.AuthorizerId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("authorization_scopes"); ok {
		input.AuthorizationScopes = flex.ExpandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("operation_name"); ok {
		input.OperationName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("request_validator_id"); ok {
		input.RequestValidatorId = aws.String(v.(string))
	}

	_, err := conn.PutMethodWithContext(ctx, &input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating API Gateway Method: %s", err)
	}

	d.SetId(fmt.Sprintf("agm-%s-%s-%s", d.Get("rest_api_id").(string), d.Get("resource_id").(string), d.Get("http_method").(string)))
	log.Printf("[DEBUG] API Gateway Method ID: %s", d.Id())

	return diags
}

func resourceMethodRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayConn()

	log.Printf("[DEBUG] Reading API Gateway Method %s", d.Id())
	out, err := conn.GetMethodWithContext(ctx, &apigateway.GetMethodInput{
		HttpMethod: aws.String(d.Get("http_method").(string)),
		ResourceId: aws.String(d.Get("resource_id").(string)),
		RestApiId:  aws.String(d.Get("rest_api_id").(string)),
	})
	if err != nil {
		if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, apigateway.ErrCodeNotFoundException) {
			log.Printf("[WARN] API Gateway Method (%s) not found, removing from state", d.Id())
			d.SetId("")
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "reading API Gateway Method (%s): %s", d.Id(), err)
	}
	log.Printf("[DEBUG] Received API Gateway Method: %s", out)

	d.Set("api_key_required", out.ApiKeyRequired)

	if err := d.Set("authorization_scopes", flex.FlattenStringList(out.AuthorizationScopes)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting authorization_scopes: %s", err)
	}

	d.Set("authorization", out.AuthorizationType)
	d.Set("authorizer_id", out.AuthorizerId)
	d.Set("operation_name", out.OperationName)

	if err := d.Set("request_models", aws.StringValueMap(out.RequestModels)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting request_models: %s", err)
	}

	if err := d.Set("request_parameters", aws.BoolValueMap(out.RequestParameters)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting request_parameters: %s", err)
	}

	d.Set("request_validator_id", out.RequestValidatorId)

	return diags
}

func resourceMethodUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayConn()

	log.Printf("[DEBUG] Updating API Gateway Method %s", d.Id())
	operations := make([]*apigateway.PatchOperation, 0)
	if d.HasChange("resource_id") {
		operations = append(operations, &apigateway.PatchOperation{
			Op:    aws.String(apigateway.OpReplace),
			Path:  aws.String("/resourceId"),
			Value: aws.String(d.Get("resource_id").(string)),
		})
	}

	if d.HasChange("request_models") {
		operations = append(operations, expandRequestResponseModelOperations(d, "request_models", "requestModels")...)
	}

	if d.HasChange("request_parameters") {
		parameters := make(map[string]bool)
		var ok bool
		for k, v := range d.Get("request_parameters").(map[string]interface{}) {
			parameters[k], ok = v.(bool)
			if !ok {
				value, _ := strconv.ParseBool(v.(string))
				parameters[k] = value
			}
		}
		ops := expandMethodParametersOperations(d, "request_parameters", "requestParameters")
		operations = append(operations, ops...)
	}

	if d.HasChange("authorization") {
		operations = append(operations, &apigateway.PatchOperation{
			Op:    aws.String(apigateway.OpReplace),
			Path:  aws.String("/authorizationType"),
			Value: aws.String(d.Get("authorization").(string)),
		})
	}

	if d.HasChange("authorizer_id") {
		operations = append(operations, &apigateway.PatchOperation{
			Op:    aws.String(apigateway.OpReplace),
			Path:  aws.String("/authorizerId"),
			Value: aws.String(d.Get("authorizer_id").(string)),
		})
	}

	if d.HasChange("authorization_scopes") {
		old, new := d.GetChange("authorization_scopes")
		path := "/authorizationScopes"

		os := old.(*schema.Set)
		ns := new.(*schema.Set)

		additionList := ns.Difference(os)
		for _, v := range additionList.List() {
			operations = append(operations, &apigateway.PatchOperation{
				Op:    aws.String(apigateway.OpAdd),
				Path:  aws.String(path),
				Value: aws.String(v.(string)),
			})
		}

		removalList := os.Difference(ns)
		for _, v := range removalList.List() {
			operations = append(operations, &apigateway.PatchOperation{
				Op:    aws.String(apigateway.OpRemove),
				Path:  aws.String(path),
				Value: aws.String(v.(string)),
			})
		}
	}

	if d.HasChange("api_key_required") {
		operations = append(operations, &apigateway.PatchOperation{
			Op:    aws.String(apigateway.OpReplace),
			Path:  aws.String("/apiKeyRequired"),
			Value: aws.String(fmt.Sprintf("%t", d.Get("api_key_required").(bool))),
		})
	}

	if d.HasChange("request_validator_id") {
		var request_validator_id *string
		if v, ok := d.GetOk("request_validator_id"); ok {
			// requestValidatorId cannot be an empty string; it must either be nil
			// or it must have some value. Otherwise, updating fails.
			if s := v.(string); len(s) > 0 {
				request_validator_id = &s
			}
		}
		operations = append(operations, &apigateway.PatchOperation{
			Op:    aws.String(apigateway.OpReplace),
			Path:  aws.String("/requestValidatorId"),
			Value: request_validator_id,
		})
	}

	if d.HasChange("operation_name") {
		var operation_name *string
		if v, ok := d.GetOk("operation_name"); ok {
			if s := v.(string); len(s) > 0 {
				operation_name = &s
			}
		}
		operations = append(operations, &apigateway.PatchOperation{
			Op:    aws.String("replace"),
			Path:  aws.String("/operationName"),
			Value: operation_name,
		})
	}

	_, err := conn.UpdateMethodWithContext(ctx, &apigateway.UpdateMethodInput{
		HttpMethod:      aws.String(d.Get("http_method").(string)),
		ResourceId:      aws.String(d.Get("resource_id").(string)),
		RestApiId:       aws.String(d.Get("rest_api_id").(string)),
		PatchOperations: operations,
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating API Gateway Method (%s): %s", d.Id(), err)
	}

	return append(diags, resourceMethodRead(ctx, d, meta)...)
}

func resourceMethodDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayConn()
	log.Printf("[DEBUG] Deleting API Gateway Method: %s", d.Id())

	_, err := conn.DeleteMethodWithContext(ctx, &apigateway.DeleteMethodInput{
		HttpMethod: aws.String(d.Get("http_method").(string)),
		ResourceId: aws.String(d.Get("resource_id").(string)),
		RestApiId:  aws.String(d.Get("rest_api_id").(string)),
	})

	if tfawserr.ErrCodeEquals(err, apigateway.ErrCodeNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting API Gateway Method (%s): %s", d.Id(), err)
	}

	return diags
}
