// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apigateway

import (
	"context"
	"fmt"
	"log"
	"strings"

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

// @SDKResource("aws_api_gateway_gateway_response", name="Gateway Response")
func resourceGatewayResponse() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceGatewayResponsePut,
		ReadWithoutTimeout:   resourceGatewayResponseRead,
		UpdateWithoutTimeout: resourceGatewayResponsePut,
		DeleteWithoutTimeout: resourceGatewayResponseDelete,

		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				idParts := strings.Split(d.Id(), "/")
				if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
					return nil, fmt.Errorf("Unexpected format of ID (%q), expected REST-API-ID/RESPONSE-TYPE", d.Id())
				}
				restApiID := idParts[0]
				responseType := idParts[1]
				d.Set("response_type", responseType)
				d.Set("rest_api_id", restApiID)
				d.SetId(fmt.Sprintf("aggr-%s-%s", restApiID, responseType))
				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"response_parameters": {
				Type:     schema.TypeMap,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Optional: true,
			},
			"response_templates": {
				Type:     schema.TypeMap,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Optional: true,
			},
			"response_type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"rest_api_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrStatusCode: {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func resourceGatewayResponsePut(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayClient(ctx)

	input := &apigateway.PutGatewayResponseInput{
		ResponseType: types.GatewayResponseType(d.Get("response_type").(string)),
		RestApiId:    aws.String(d.Get("rest_api_id").(string)),
	}

	if v, ok := d.GetOk("response_parameters"); ok && len(v.(map[string]interface{})) > 0 {
		input.ResponseParameters = flex.ExpandStringValueMap(v.(map[string]interface{}))
	}

	if v, ok := d.GetOk("response_templates"); ok && len(v.(map[string]interface{})) > 0 {
		input.ResponseTemplates = flex.ExpandStringValueMap(v.(map[string]interface{}))
	}

	if v, ok := d.GetOk(names.AttrStatusCode); ok {
		input.StatusCode = aws.String(v.(string))
	}

	_, err := conn.PutGatewayResponse(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "putting API Gateway Gateway Response: %s", err)
	}

	if d.IsNewResource() {
		d.SetId(fmt.Sprintf("aggr-%s-%s", d.Get("rest_api_id").(string), d.Get("response_type").(string)))
	}

	return append(diags, resourceGatewayResponseRead(ctx, d, meta)...)
}

func resourceGatewayResponseRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayClient(ctx)

	gatewayResponse, err := findGatewayResponseByTwoPartKey(ctx, conn, d.Get("response_type").(string), d.Get("rest_api_id").(string))

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] API Gateway Gateway Response (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading API Gateway Gateway Response (%s): %s", d.Id(), err)
	}

	d.Set("response_parameters", gatewayResponse.ResponseParameters)
	d.Set("response_templates", gatewayResponse.ResponseTemplates)
	d.Set("response_type", gatewayResponse.ResponseType)
	d.Set(names.AttrStatusCode, gatewayResponse.StatusCode)

	return diags
}

func resourceGatewayResponseDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayClient(ctx)

	log.Printf("[DEBUG] Deleting API Gateway Gateway Response: %s", d.Id())
	_, err := conn.DeleteGatewayResponse(ctx, &apigateway.DeleteGatewayResponseInput{
		ResponseType: types.GatewayResponseType(d.Get("response_type").(string)),
		RestApiId:    aws.String(d.Get("rest_api_id").(string)),
	})

	if errs.IsA[*types.NotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting API Gateway Gateway Response (%s): %s", d.Id(), err)
	}

	return diags
}

func findGatewayResponseByTwoPartKey(ctx context.Context, conn *apigateway.Client, responseType, apiID string) (*apigateway.GetGatewayResponseOutput, error) {
	input := &apigateway.GetGatewayResponseInput{
		ResponseType: types.GatewayResponseType(responseType),
		RestApiId:    aws.String(apiID),
	}

	output, err := conn.GetGatewayResponse(ctx, input)

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
