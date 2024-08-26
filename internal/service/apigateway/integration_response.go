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

// @SDKResource("aws_api_gateway_integration_response", name="Integration Response")
func resourceIntegrationResponse() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceIntegrationResponsePut,
		ReadWithoutTimeout:   resourceIntegrationResponseRead,
		UpdateWithoutTimeout: resourceIntegrationResponsePut,
		DeleteWithoutTimeout: resourceIntegrationResponseDelete,

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
				d.SetId(fmt.Sprintf("agir-%s-%s-%s-%s", restApiID, resourceID, httpMethod, statusCode))
				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"content_handling": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: validIntegrationContentHandling(),
			},
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
			"response_parameters": {
				Type:     schema.TypeMap,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Optional: true,
			},
			"response_templates": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"rest_api_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"selection_pattern": {
				Type:     schema.TypeString,
				Optional: true,
			},
			names.AttrStatusCode: {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceIntegrationResponsePut(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayClient(ctx)

	input := &apigateway.PutIntegrationResponseInput{
		HttpMethod: aws.String(d.Get("http_method").(string)),
		ResourceId: aws.String(d.Get(names.AttrResourceID).(string)),
		RestApiId:  aws.String(d.Get("rest_api_id").(string)),
		StatusCode: aws.String(d.Get(names.AttrStatusCode).(string)),
	}

	if v, ok := d.GetOk("content_handling"); ok {
		input.ContentHandling = types.ContentHandlingStrategy(v.(string))
	}

	if v, ok := d.GetOk("response_parameters"); ok && len(v.(map[string]interface{})) > 0 {
		input.ResponseParameters = flex.ExpandStringValueMap(v.(map[string]interface{}))
	}

	if v, ok := d.GetOk("response_templates"); ok && len(v.(map[string]interface{})) > 0 {
		input.ResponseTemplates = flex.ExpandStringValueMap(v.(map[string]interface{}))
	}

	if v, ok := d.GetOk("selection_pattern"); ok {
		input.SelectionPattern = aws.String(v.(string))
	}

	_, err := conn.PutIntegrationResponse(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "putting API Gateway Integration Response: %s", err)
	}

	if d.IsNewResource() {
		d.SetId(fmt.Sprintf("agir-%s-%s-%s-%s", d.Get("rest_api_id").(string), d.Get(names.AttrResourceID).(string), d.Get("http_method").(string), d.Get(names.AttrStatusCode).(string)))
	}

	return append(diags, resourceIntegrationResponseRead(ctx, d, meta)...)
}

func resourceIntegrationResponseRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayClient(ctx)

	integrationResponse, err := findIntegrationResponseByFourPartKey(ctx, conn, d.Get("http_method").(string), d.Get(names.AttrResourceID).(string), d.Get("rest_api_id").(string), d.Get(names.AttrStatusCode).(string))

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] API Gateway Integration Response (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading API Gateway Integration Response (%s): %s", d.Id(), err)
	}

	d.Set("content_handling", integrationResponse.ContentHandling)
	d.Set("response_parameters", integrationResponse.ResponseParameters)
	// We need to explicitly convert key = nil values into key = "", which aws.StringValueMap() removes.
	responseTemplates := make(map[string]string)
	for k, v := range integrationResponse.ResponseTemplates {
		responseTemplates[k] = v
	}
	d.Set("response_templates", responseTemplates)
	d.Set("selection_pattern", integrationResponse.SelectionPattern)

	return diags
}

func resourceIntegrationResponseDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayClient(ctx)

	log.Printf("[DEBUG] Deleting API Gateway Integration Response: %s", d.Id())
	_, err := conn.DeleteIntegrationResponse(ctx, &apigateway.DeleteIntegrationResponseInput{
		HttpMethod: aws.String(d.Get("http_method").(string)),
		ResourceId: aws.String(d.Get(names.AttrResourceID).(string)),
		RestApiId:  aws.String(d.Get("rest_api_id").(string)),
		StatusCode: aws.String(d.Get(names.AttrStatusCode).(string)),
	})

	if errs.IsA[*types.NotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting API Gateway Integration Response (%s): %s", d.Id(), err)
	}

	return diags
}

func findIntegrationResponseByFourPartKey(ctx context.Context, conn *apigateway.Client, httpMethod, resourceID, apiID, statusCode string) (*apigateway.GetIntegrationResponseOutput, error) {
	input := &apigateway.GetIntegrationResponseInput{
		HttpMethod: aws.String(httpMethod),
		ResourceId: aws.String(resourceID),
		RestApiId:  aws.String(apiID),
		StatusCode: aws.String(statusCode),
	}

	output, err := conn.GetIntegrationResponse(ctx, input)

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
