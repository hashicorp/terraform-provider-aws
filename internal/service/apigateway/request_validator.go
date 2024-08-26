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
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_api_gateway_request_validator", name="Request Validator")
func resourceRequestValidator() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceRequestValidatorCreate,
		ReadWithoutTimeout:   resourceRequestValidatorRead,
		UpdateWithoutTimeout: resourceRequestValidatorUpdate,
		DeleteWithoutTimeout: resourceRequestValidatorDelete,

		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				idParts := strings.Split(d.Id(), "/")
				if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
					return nil, fmt.Errorf("Unexpected format of ID (%q), expected REST-API-ID/REQUEST-VALIDATOR-ID", d.Id())
				}
				restApiID := idParts[0]
				requestValidatorID := idParts[1]
				d.Set("rest_api_id", restApiID)
				d.SetId(requestValidatorID)
				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
			},
			"rest_api_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"validate_request_body": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"validate_request_parameters": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
		},
	}
}

func resourceRequestValidatorCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &apigateway.CreateRequestValidatorInput{
		Name:                      aws.String(name),
		RestApiId:                 aws.String(d.Get("rest_api_id").(string)),
		ValidateRequestBody:       d.Get("validate_request_body").(bool),
		ValidateRequestParameters: d.Get("validate_request_parameters").(bool),
	}

	output, err := conn.CreateRequestValidator(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating API Gateway Request Validator (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.Id))

	return diags
}

func resourceRequestValidatorRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayClient(ctx)

	output, err := findRequestValidatorByTwoPartKey(ctx, conn, d.Id(), d.Get("rest_api_id").(string))

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] API Gateway Request Validator (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading API Gateway Request Validator (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrName, output.Name)
	d.Set("validate_request_body", output.ValidateRequestBody)
	d.Set("validate_request_parameters", output.ValidateRequestParameters)

	return diags
}

func resourceRequestValidatorUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayClient(ctx)

	operations := make([]types.PatchOperation, 0)

	if d.HasChange(names.AttrName) {
		operations = append(operations, types.PatchOperation{
			Op:    types.OpReplace,
			Path:  aws.String("/name"),
			Value: aws.String(d.Get(names.AttrName).(string)),
		})
	}

	if d.HasChange("validate_request_body") {
		operations = append(operations, types.PatchOperation{
			Op:    types.OpReplace,
			Path:  aws.String("/validateRequestBody"),
			Value: aws.String(fmt.Sprintf("%t", d.Get("validate_request_body").(bool))),
		})
	}

	if d.HasChange("validate_request_parameters") {
		operations = append(operations, types.PatchOperation{
			Op:    types.OpReplace,
			Path:  aws.String("/validateRequestParameters"),
			Value: aws.String(fmt.Sprintf("%t", d.Get("validate_request_parameters").(bool))),
		})
	}

	input := &apigateway.UpdateRequestValidatorInput{
		RequestValidatorId: aws.String(d.Id()),
		RestApiId:          aws.String(d.Get("rest_api_id").(string)),
		PatchOperations:    operations,
	}

	_, err := conn.UpdateRequestValidator(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating API Gateway Request Validator (%s): %s", d.Id(), err)
	}

	return append(diags, resourceRequestValidatorRead(ctx, d, meta)...)
}

func resourceRequestValidatorDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayClient(ctx)

	log.Printf("[DEBUG] Deleting API Gateway Request Validator: %s", d.Id())
	_, err := conn.DeleteRequestValidator(ctx, &apigateway.DeleteRequestValidatorInput{
		RequestValidatorId: aws.String(d.Id()),
		RestApiId:          aws.String(d.Get("rest_api_id").(string)),
	})

	// XXX: Figure out a way to delete the method that depends on the request validator first
	// otherwise the validator will be dangling until the API is deleted.
	if errs.IsA[*types.ConflictException](err) || errs.IsA[*types.NotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting API Gateway Request Validator (%s): %s", d.Id(), err)
	}

	return diags
}

func findRequestValidatorByTwoPartKey(ctx context.Context, conn *apigateway.Client, requestValidatorID, apiID string) (*apigateway.GetRequestValidatorOutput, error) {
	input := &apigateway.GetRequestValidatorInput{
		RequestValidatorId: aws.String(requestValidatorID),
		RestApiId:          aws.String(apiID),
	}

	output, err := conn.GetRequestValidator(ctx, input)

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
