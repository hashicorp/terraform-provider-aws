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

// @SDKResource("aws_api_gateway_resource", name="Resource")
func resourceResource() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceResourceCreate,
		ReadWithoutTimeout:   resourceResourceRead,
		UpdateWithoutTimeout: resourceResourceUpdate,
		DeleteWithoutTimeout: resourceResourceDelete,

		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				idParts := strings.Split(d.Id(), "/")
				if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
					return nil, fmt.Errorf("Unexpected format of ID (%q), expected REST-API-ID/RESOURCE-ID", d.Id())
				}
				restApiID := idParts[0]
				resourceID := idParts[1]
				d.Set("rest_api_id", restApiID)
				d.SetId(resourceID)
				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"parent_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrPath: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"path_part": {
				Type:     schema.TypeString,
				Required: true,
			},
			"rest_api_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceResourceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayClient(ctx)

	input := &apigateway.CreateResourceInput{
		ParentId:  aws.String(d.Get("parent_id").(string)),
		PathPart:  aws.String(d.Get("path_part").(string)),
		RestApiId: aws.String(d.Get("rest_api_id").(string)),
	}

	output, err := conn.CreateResource(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating API Gateway Resource: %s", err)
	}

	d.SetId(aws.ToString(output.Id))

	return append(diags, resourceResourceRead(ctx, d, meta)...)
}

func resourceResourceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayClient(ctx)

	resource, err := findResourceByTwoPartKey(ctx, conn, d.Id(), d.Get("rest_api_id").(string))

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] API Gateway Resource (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading API Gateway Resource (%s): %s", d.Id(), err)
	}

	d.Set("parent_id", resource.ParentId)
	d.Set("path_part", resource.PathPart)
	d.Set(names.AttrPath, resource.Path)

	return diags
}

func resourceResourceUpdateOperations(d *schema.ResourceData) []types.PatchOperation {
	operations := make([]types.PatchOperation, 0)
	if d.HasChange("path_part") {
		operations = append(operations, types.PatchOperation{
			Op:    types.OpReplace,
			Path:  aws.String("/pathPart"),
			Value: aws.String(d.Get("path_part").(string)),
		})
	}

	if d.HasChange("parent_id") {
		operations = append(operations, types.PatchOperation{
			Op:    types.OpReplace,
			Path:  aws.String("/parentId"),
			Value: aws.String(d.Get("parent_id").(string)),
		})
	}
	return operations
}

func resourceResourceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayClient(ctx)

	input := &apigateway.UpdateResourceInput{
		ResourceId:      aws.String(d.Id()),
		RestApiId:       aws.String(d.Get("rest_api_id").(string)),
		PatchOperations: resourceResourceUpdateOperations(d),
	}

	_, err := conn.UpdateResource(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating API Gateway Resource (%s): %s", d.Id(), err)
	}

	return append(diags, resourceResourceRead(ctx, d, meta)...)
}

func resourceResourceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayClient(ctx)

	log.Printf("[DEBUG] Deleting API Gateway Resource: %s", d.Id())
	_, err := conn.DeleteResource(ctx, &apigateway.DeleteResourceInput{
		ResourceId: aws.String(d.Id()),
		RestApiId:  aws.String(d.Get("rest_api_id").(string)),
	})

	if errs.IsA[*types.NotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting API Gateway Resource (%s): %s", d.Id(), err)
	}

	return diags
}

func findResourceByTwoPartKey(ctx context.Context, conn *apigateway.Client, resourceID, apiID string) (*apigateway.GetResourceOutput, error) {
	input := &apigateway.GetResourceInput{
		ResourceId: aws.String(resourceID),
		RestApiId:  aws.String(apiID),
	}

	output, err := conn.GetResource(ctx, input)

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
