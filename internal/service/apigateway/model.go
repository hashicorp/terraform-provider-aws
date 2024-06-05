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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_api_gateway_model", name="Model")
func resourceModel() *schema.Resource {
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
				apiID := idParts[0]
				name := idParts[1]
				d.Set(names.AttrName, name)
				d.Set("rest_api_id", apiID)

				conn := meta.(*conns.AWSClient).APIGatewayClient(ctx)

				model, err := findModelByTwoPartKey(ctx, conn, name, apiID)

				if err != nil {
					return nil, err
				}

				d.SetId(aws.ToString(model.Id))

				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			names.AttrContentType: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Optional: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"rest_api_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrSchema: {
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
	conn := meta.(*conns.AWSClient).APIGatewayClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &apigateway.CreateModelInput{
		ContentType: aws.String(d.Get(names.AttrContentType).(string)),
		Name:        aws.String(name),
		RestApiId:   aws.String(d.Get("rest_api_id").(string)),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrSchema); ok {
		input.Schema = aws.String(v.(string))
	}

	output, err := conn.CreateModel(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating API Gateway Model (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.Id))

	return diags
}

func resourceModelRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayClient(ctx)

	model, err := findModelByTwoPartKey(ctx, conn, d.Get(names.AttrName).(string), d.Get("rest_api_id").(string))

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] API Gateway Model (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading API Gateway Model (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrContentType, model.ContentType)
	d.Set(names.AttrDescription, model.Description)
	d.Set(names.AttrSchema, model.Schema)

	return diags
}

func resourceModelUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayClient(ctx)

	operations := make([]types.PatchOperation, 0)

	if d.HasChange(names.AttrDescription) {
		operations = append(operations, types.PatchOperation{
			Op:    types.OpReplace,
			Path:  aws.String("/description"),
			Value: aws.String(d.Get(names.AttrDescription).(string)),
		})
	}

	if d.HasChange(names.AttrSchema) {
		operations = append(operations, types.PatchOperation{
			Op:    types.OpReplace,
			Path:  aws.String("/schema"),
			Value: aws.String(d.Get(names.AttrSchema).(string)),
		})
	}

	input := &apigateway.UpdateModelInput{
		ModelName:       aws.String(d.Get(names.AttrName).(string)),
		PatchOperations: operations,
		RestApiId:       aws.String(d.Get("rest_api_id").(string)),
	}

	_, err := conn.UpdateModel(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating API Gateway Model (%s): %s", d.Id(), err)
	}

	return append(diags, resourceModelRead(ctx, d, meta)...)
}

func resourceModelDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayClient(ctx)

	log.Printf("[DEBUG] Deleting API Gateway Model: %s", d.Id())
	_, err := conn.DeleteModel(ctx, &apigateway.DeleteModelInput{
		ModelName: aws.String(d.Get(names.AttrName).(string)),
		RestApiId: aws.String(d.Get("rest_api_id").(string)),
	})

	if errs.IsA[*types.NotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting API Gateway Model (%s): %s", d.Id(), err)
	}

	return diags
}

func findModelByTwoPartKey(ctx context.Context, conn *apigateway.Client, name, apiID string) (*apigateway.GetModelOutput, error) {
	input := &apigateway.GetModelInput{
		ModelName: aws.String(name),
		RestApiId: aws.String(apiID),
	}

	output, err := conn.GetModel(ctx, input)

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
