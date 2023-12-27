// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apigateway

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKResource("aws_api_gateway_model")
func ResourceModel() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceModelCreate,
		ReadWithoutTimeout:   resourceModelRead,
		UpdateWithoutTimeout: resourceModelUpdate,
		DeleteWithoutTimeout: resourceModelDelete,

		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				idParts := strings.Split(d.Id(), "/")
				if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
					return nil, fmt.Errorf("Unexpected format of ID (%q), expected REST-API-ID/NAME", d.Id())
				}
				restApiID := idParts[0]
				name := idParts[1]
				d.Set("name", name)
				d.Set("rest_api_id", restApiID)

				conn := meta.(*conns.AWSClient).APIGatewayConn(ctx)

				output, err := conn.GetModelWithContext(ctx, &apigateway.GetModelInput{
					ModelName: aws.String(name),
					RestApiId: aws.String(restApiID),
				})

				if err != nil {
					return nil, err
				}

				d.SetId(aws.StringValue(output.Id))

				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"content_type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"rest_api_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"schema": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateFunc:     validation.StringIsJSON,
				DiffSuppressFunc: verify.SuppressEquivalentJSONDiffs,
				StateFunc: func(v interface{}) string {
					json, _ := structure.NormalizeJsonString(v)
					return json
				},
			},
		},
	}
}

func resourceModelCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayConn(ctx)

	name := d.Get("name").(string)
	input := &apigateway.CreateModelInput{
		ContentType: aws.String(d.Get("content_type").(string)),
		Name:        aws.String(name),
		RestApiId:   aws.String(d.Get("rest_api_id").(string)),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("schema"); ok {
		input.Schema = aws.String(v.(string))
	}

	output, err := conn.CreateModelWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating API Gateway Model (%s): %s", name, err)
	}

	d.SetId(aws.StringValue(output.Id))

	return diags
}

func resourceModelRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayConn(ctx)

	model, err := FindModelByTwoPartKey(ctx, conn, d.Get("name").(string), d.Get("rest_api_id").(string))

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] API Gateway Model (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading API Gateway Model (%s): %s", d.Id(), err)
	}

	d.Set("content_type", model.ContentType)
	d.Set("description", model.Description)
	d.Set("schema", model.Schema)

	return diags
}

func resourceModelUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayConn(ctx)

	operations := make([]*apigateway.PatchOperation, 0)

	if d.HasChange("description") {
		operations = append(operations, &apigateway.PatchOperation{
			Op:    aws.String(apigateway.OpReplace),
			Path:  aws.String("/description"),
			Value: aws.String(d.Get("description").(string)),
		})
	}

	if d.HasChange("schema") {
		operations = append(operations, &apigateway.PatchOperation{
			Op:    aws.String(apigateway.OpReplace),
			Path:  aws.String("/schema"),
			Value: aws.String(d.Get("schema").(string)),
		})
	}

	input := &apigateway.UpdateModelInput{
		ModelName:       aws.String(d.Get("name").(string)),
		PatchOperations: operations,
		RestApiId:       aws.String(d.Get("rest_api_id").(string)),
	}

	_, err := conn.UpdateModelWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating API Gateway Model (%s): %s", d.Id(), err)
	}

	return append(diags, resourceModelRead(ctx, d, meta)...)
}

func resourceModelDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayConn(ctx)

	log.Printf("[DEBUG] Deleting API Gateway Model: %s", d.Id())
	_, err := conn.DeleteModelWithContext(ctx, &apigateway.DeleteModelInput{
		ModelName: aws.String(d.Get("name").(string)),
		RestApiId: aws.String(d.Get("rest_api_id").(string)),
	})

	if tfawserr.ErrCodeEquals(err, apigateway.ErrCodeNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting API Gateway Model (%s): %s", d.Id(), err)
	}

	return diags
}

func FindModelByTwoPartKey(ctx context.Context, conn *apigateway.APIGateway, name, apiID string) (*apigateway.Model, error) {
	input := &apigateway.GetModelInput{
		ModelName: aws.String(name),
		RestApiId: aws.String(apiID),
	}

	output, err := conn.GetModelWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, apigateway.ErrCodeNotFoundException) {
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
