// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apigateway

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/apigateway"
	"github.com/aws/aws-sdk-go-v2/service/apigateway/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_api_gateway_method_response", name="Method Response")
func resourceMethodResponse() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceMethodResponseCreate,
		ReadWithoutTimeout:   resourceMethodResponseRead,
		UpdateWithoutTimeout: resourceMethodResponseUpdate,
		DeleteWithoutTimeout: resourceMethodResponseDelete,

		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				idParts := strings.Split(d.Id(), "/")
				if len(idParts) != 4 || idParts[0] == "" || idParts[1] == "" || idParts[2] == "" || idParts[3] == "" {
					return nil, fmt.Errorf("Unexpected format of ID (%q), expected REST-API-ID/RESOURCE-ID/HTTP-METHOD/STATUS-CODE", d.Id())
				}
				restApiID := idParts[0]
				resourceID := idParts[1]
				httpMethod := idParts[2]
				statusCode := idParts[3]
				d.Set("http_method", httpMethod)
				d.Set(names.AttrStatusCode, statusCode)
				d.Set(names.AttrResourceID, resourceID)
				d.Set("rest_api_id", restApiID)
				d.SetId(fmt.Sprintf("agmr-%s-%s-%s-%s", restApiID, resourceID, httpMethod, statusCode))
				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"http_method": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validHTTPMethod(),
			},
			names.AttrResourceID: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"response_models": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"response_parameters": {
				Type:     schema.TypeMap,
				Elem:     &schema.Schema{Type: schema.TypeBool},
				Optional: true,
			},
			"rest_api_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrStatusCode: {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceMethodResponseCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayClient(ctx)

	input := &apigateway.PutMethodResponseInput{
		HttpMethod: aws.String(d.Get("http_method").(string)),
		ResourceId: aws.String(d.Get(names.AttrResourceID).(string)),
		RestApiId:  aws.String(d.Get("rest_api_id").(string)),
		StatusCode: aws.String(d.Get(names.AttrStatusCode).(string)),
	}

	if v, ok := d.GetOk("response_models"); ok && len(v.(map[string]interface{})) > 0 {
		input.ResponseModels = flex.ExpandStringValueMap(v.(map[string]interface{}))
	}

	if v, ok := d.GetOk("response_parameters"); ok && len(v.(map[string]interface{})) > 0 {
		input.ResponseParameters = flex.ExpandBoolValueMap(v.(map[string]interface{}))
	}

	mutexKey := "api-gateway-method-response"
	conns.GlobalMutexKV.Lock(mutexKey)
	defer conns.GlobalMutexKV.Unlock(mutexKey)

	const (
		timeout = 2 * time.Minute
	)
	_, err := tfresource.RetryWhenIsA[*types.ConflictException](ctx, timeout, func() (interface{}, error) {
		return conn.PutMethodResponse(ctx, input)
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating API Gateway Method Response: %s", err)
	}

	d.SetId(fmt.Sprintf("agmr-%s-%s-%s-%s", d.Get("rest_api_id").(string), d.Get(names.AttrResourceID).(string), d.Get("http_method").(string), d.Get(names.AttrStatusCode).(string)))

	return diags
}

func resourceMethodResponseRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayClient(ctx)

	methodResponse, err := findMethodResponseByFourPartKey(ctx, conn, d.Get("http_method").(string), d.Get(names.AttrResourceID).(string), d.Get("rest_api_id").(string), d.Get(names.AttrStatusCode).(string))

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] API Gateway Method Response (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading API Gateway Method Response (%s): %s", d.Id(), err)
	}

	if err := d.Set("response_models", methodResponse.ResponseModels); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting response_models: %s", err)
	}
	if err := d.Set("response_parameters", methodResponse.ResponseParameters); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting response_parameters: %s", err)
	}

	return diags
}

func resourceMethodResponseUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayClient(ctx)

	operations := make([]types.PatchOperation, 0)

	if d.HasChange("response_models") {
		operations = append(operations, expandRequestResponseModelOperations(d, "response_models", "responseModels")...)
	}

	if d.HasChange("response_parameters") {
		operations = append(operations, expandMethodParametersOperations(d, "response_parameters", "responseParameters")...)
	}

	input := &apigateway.UpdateMethodResponseInput{
		HttpMethod:      aws.String(d.Get("http_method").(string)),
		PatchOperations: operations,
		ResourceId:      aws.String(d.Get(names.AttrResourceID).(string)),
		RestApiId:       aws.String(d.Get("rest_api_id").(string)),
		StatusCode:      aws.String(d.Get(names.AttrStatusCode).(string)),
	}

	_, err := conn.UpdateMethodResponse(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating API Gateway Method Response (%s): %s", d.Id(), err)
	}

	return append(diags, resourceMethodResponseRead(ctx, d, meta)...)
}

func resourceMethodResponseDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayClient(ctx)

	log.Printf("[DEBUG] Deleting API Gateway Method Response: %s", d.Id())
	_, err := conn.DeleteMethodResponse(ctx, &apigateway.DeleteMethodResponseInput{
		HttpMethod: aws.String(d.Get("http_method").(string)),
		ResourceId: aws.String(d.Get(names.AttrResourceID).(string)),
		RestApiId:  aws.String(d.Get("rest_api_id").(string)),
		StatusCode: aws.String(d.Get(names.AttrStatusCode).(string)),
	})

	if errs.IsA[*types.NotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting API Gateway Method Response (%s): %s", d.Id(), err)
	}

	return diags
}

func findMethodResponseByFourPartKey(ctx context.Context, conn *apigateway.Client, httpMethod, resourceID, apiID, statusCode string) (*apigateway.GetMethodResponseOutput, error) {
	input := &apigateway.GetMethodResponseInput{
		HttpMethod: aws.String(httpMethod),
		ResourceId: aws.String(resourceID),
		RestApiId:  aws.String(apiID),
		StatusCode: aws.String(statusCode),
	}

	output, err := conn.GetMethodResponse(ctx, input)

	if errs.IsA[*types.NotFoundException](err) {
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
