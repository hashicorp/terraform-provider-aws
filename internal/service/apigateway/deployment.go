// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apigateway

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKResource("aws_api_gateway_deployment")
func ResourceDeployment() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceDeploymentCreate,
		ReadWithoutTimeout:   resourceDeploymentRead,
		UpdateWithoutTimeout: resourceDeploymentUpdate,
		DeleteWithoutTimeout: resourceDeploymentDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceDeploymentImport,
		},

		Schema: map[string]*schema.Schema{
			"created_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"execution_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"invoke_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"rest_api_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"stage_description": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"stage_name": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"triggers": {
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

func resourceDeploymentCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayConn(ctx)

	input := &apigateway.CreateDeploymentInput{
		Description:      aws.String(d.Get("description").(string)),
		RestApiId:        aws.String(d.Get("rest_api_id").(string)),
		StageDescription: aws.String(d.Get("stage_description").(string)),
		StageName:        aws.String(d.Get("stage_name").(string)),
		Variables:        flex.ExpandStringMap(d.Get("variables").(map[string]interface{})),
	}

	deployment, err := conn.CreateDeploymentWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating API Gateway Deployment: %s", err)
	}

	d.SetId(aws.StringValue(deployment.Id))

	return append(diags, resourceDeploymentRead(ctx, d, meta)...)
}

func resourceDeploymentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayConn(ctx)

	restAPIID := d.Get("rest_api_id").(string)
	deployment, err := FindDeploymentByTwoPartKey(ctx, conn, restAPIID, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] API Gateway Deployment (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading API Gateway Deployment (%s): %s", d.Id(), err)
	}

	stageName := d.Get("stage_name").(string)
	d.Set("created_date", deployment.CreatedDate.Format(time.RFC3339))
	d.Set("description", deployment.Description)
	executionARN := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "execute-api",
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("%s/%s", restAPIID, stageName),
	}.String()
	d.Set("execution_arn", executionARN)
	d.Set("invoke_url", meta.(*conns.AWSClient).APIGatewayInvokeURL(restAPIID, stageName))

	return diags
}

func resourceDeploymentUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

	if len(operations) > 0 {
		_, err := conn.UpdateDeploymentWithContext(ctx, &apigateway.UpdateDeploymentInput{
			DeploymentId:    aws.String(d.Id()),
			PatchOperations: operations,
			RestApiId:       aws.String(d.Get("rest_api_id").(string)),
		})

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating API Gateway Deployment (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceDeploymentRead(ctx, d, meta)...)
}

func resourceDeploymentDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayConn(ctx)

	// If the stage has been updated to point at a different deployment, then
	// the stage should not be removed when this deployment is deleted.
	shouldDeleteStage := false
	// API Gateway allows an empty state name (e.g. ""), but the AWS Go SDK
	// now has validation for the parameter, so we must check first.
	// InvalidParameter: 1 validation error(s) found.
	//  - minimum field size of 1, GetStageInput.StageName.
	restAPIID := d.Get("rest_api_id").(string)
	stageName := d.Get("stage_name").(string)
	if stageName != "" {
		stage, err := FindStageByTwoPartKey(ctx, conn, restAPIID, stageName)

		if err == nil {
			shouldDeleteStage = aws.StringValue(stage.DeploymentId) == d.Id()
		} else if !tfresource.NotFound(err) {
			return sdkdiag.AppendErrorf(diags, "reading API Gateway Stage (%s): %s", stageName, err)
		}
	}

	if shouldDeleteStage {
		_, err := conn.DeleteStageWithContext(ctx, &apigateway.DeleteStageInput{
			StageName: aws.String(stageName),
			RestApiId: aws.String(restAPIID),
		})

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "deleting API Gateway Stage (%s): %s", stageName, err)
		}
	}

	log.Printf("[DEBUG] Deleting API Gateway Deployment: %s", d.Id())
	_, err := conn.DeleteDeploymentWithContext(ctx, &apigateway.DeleteDeploymentInput{
		DeploymentId: aws.String(d.Id()),
		RestApiId:    aws.String(restAPIID),
	})

	if tfawserr.ErrCodeEquals(err, apigateway.ErrCodeNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting API Gateway Deployment (%s): %s", d.Id(), err)
	}

	return diags
}

func resourceDeploymentImport(_ context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
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

func FindDeploymentByTwoPartKey(ctx context.Context, conn *apigateway.APIGateway, restAPIID, deploymentID string) (*apigateway.Deployment, error) {
	input := &apigateway.GetDeploymentInput{
		DeploymentId: aws.String(deploymentID),
		RestApiId:    aws.String(restAPIID),
	}

	output, err := conn.GetDeploymentWithContext(ctx, input)

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
