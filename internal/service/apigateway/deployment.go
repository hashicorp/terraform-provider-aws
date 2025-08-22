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

// @SDKResource("aws_api_gateway_deployment", name="Deployment")
func resourceDeployment() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceDeploymentCreate,
		ReadWithoutTimeout:   resourceDeploymentRead,
		UpdateWithoutTimeout: resourceDeploymentUpdate,
		DeleteWithoutTimeout: resourceDeploymentDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceDeploymentImport,
		},

		Schema: map[string]*schema.Schema{
			names.AttrCreatedDate: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Optional: true,
			},
			"rest_api_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrTriggers: {
				Type:     schema.TypeMap,
				Optional: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"variables": {
				Type:     schema.TypeMap,
				Optional: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func resourceDeploymentCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayClient(ctx)

	input := apigateway.CreateDeploymentInput{
		Description: aws.String(d.Get(names.AttrDescription).(string)),
		RestApiId:   aws.String(d.Get("rest_api_id").(string)),
		Variables:   flex.ExpandStringValueMap(d.Get("variables").(map[string]any)),
	}

	deployment, err := conn.CreateDeployment(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating API Gateway Deployment: %s", err)
	}

	d.SetId(aws.ToString(deployment.Id))

	return append(diags, resourceDeploymentRead(ctx, d, meta)...)
}

func resourceDeploymentRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayClient(ctx)

	restAPIID := d.Get("rest_api_id").(string)
	deployment, err := findDeploymentByTwoPartKey(ctx, conn, restAPIID, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] API Gateway Deployment (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading API Gateway Deployment (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrCreatedDate, deployment.CreatedDate.Format(time.RFC3339))
	d.Set(names.AttrDescription, deployment.Description)

	return diags
}

func resourceDeploymentUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
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

	if len(operations) > 0 {
		input := apigateway.UpdateDeploymentInput{
			DeploymentId:    aws.String(d.Id()),
			PatchOperations: operations,
			RestApiId:       aws.String(d.Get("rest_api_id").(string)),
		}
		_, err := conn.UpdateDeployment(ctx, &input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating API Gateway Deployment (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceDeploymentRead(ctx, d, meta)...)
}

func resourceDeploymentDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayClient(ctx)

	restAPIID := d.Get("rest_api_id").(string)

	log.Printf("[DEBUG] Deleting API Gateway Deployment: %s", d.Id())
	input := apigateway.DeleteDeploymentInput{
		DeploymentId: aws.String(d.Id()),
		RestApiId:    aws.String(restAPIID),
	}
	_, err := conn.DeleteDeployment(ctx, &input)

	if errs.IsAErrorMessageContains[*types.BadRequestException](err, "Active stages with canary settings pointing to this deployment must be moved or deleted") {
		deploymentInput := apigateway.DeleteDeploymentInput{
			DeploymentId: aws.String(d.Id()),
			RestApiId:    aws.String(restAPIID),
		}
		_, err = conn.DeleteDeployment(ctx, &deploymentInput)
	}

	if errs.IsA[*types.NotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting API Gateway Deployment (%s): %s", d.Id(), err)
	}

	return diags
}

func resourceDeploymentImport(_ context.Context, d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
	idParts := strings.Split(d.Id(), "/")
	if len(idParts) != 2 {
		return nil, fmt.Errorf("Unexpected format of ID (%s), use: 'REST-API-ID/DEPLOYMENT-ID'", d.Id())
	}

	restApiID := idParts[0]
	deploymentID := idParts[1]

	d.SetId(deploymentID)
	d.Set("rest_api_id", restApiID)

	return []*schema.ResourceData{d}, nil
}

func findDeploymentByTwoPartKey(ctx context.Context, conn *apigateway.Client, restAPIID, deploymentID string) (*apigateway.GetDeploymentOutput, error) {
	input := apigateway.GetDeploymentInput{
		DeploymentId: aws.String(deploymentID),
		RestApiId:    aws.String(restAPIID),
	}

	output, err := conn.GetDeployment(ctx, &input)

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
