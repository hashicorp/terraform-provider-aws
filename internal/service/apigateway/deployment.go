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
	"github.com/aws/aws-sdk-go-v2/aws/arn"
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
			"canary_settings": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"percent_traffic": {
							Type:     schema.TypeFloat,
							Optional: true,
							Default:  0.0,
						},
						"stage_variable_overrides": {
							Type:     schema.TypeMap,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Optional: true,
						},
						"use_stage_cache": {
							Type:     schema.TypeBool,
							Optional: true,
						},
					},
				},
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

func resourceDeploymentCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayClient(ctx)

	input := &apigateway.CreateDeploymentInput{
		Description:      aws.String(d.Get(names.AttrDescription).(string)),
		RestApiId:        aws.String(d.Get("rest_api_id").(string)),
		StageDescription: aws.String(d.Get("stage_description").(string)),
		StageName:        aws.String(d.Get("stage_name").(string)),
		Variables:        flex.ExpandStringValueMap(d.Get("variables").(map[string]interface{})),
	}

	deployment, err := conn.CreateDeployment(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating API Gateway Deployment: %s", err)
	}

	d.SetId(aws.ToString(deployment.Id))
	if v, ok := d.GetOk("canary_settings"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.CanarySettings = expandDeploymentCanarySettings(v.([]interface{})[0].(map[string]interface{}))
	}

	return append(diags, resourceDeploymentRead(ctx, d, meta)...)
}

func resourceDeploymentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

	stageName := d.Get("stage_name").(string)
	d.Set(names.AttrCreatedDate, deployment.CreatedDate.Format(time.RFC3339))
	d.Set(names.AttrDescription, deployment.Description)
	executionARN := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "execute-api",
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("%s/%s", restAPIID, stageName),
	}.String()
	d.Set("execution_arn", executionARN)
	d.Set("invoke_url", meta.(*conns.AWSClient).APIGatewayInvokeURL(ctx, restAPIID, stageName))

	return diags
}

func resourceDeploymentUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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
		_, err := conn.UpdateDeployment(ctx, &apigateway.UpdateDeploymentInput{
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
	conn := meta.(*conns.AWSClient).APIGatewayClient(ctx)

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
		stage, err := findStageByTwoPartKey(ctx, conn, restAPIID, stageName)

		if err == nil {
			shouldDeleteStage = aws.ToString(stage.DeploymentId) == d.Id()
		} else if !tfresource.NotFound(err) {
			return sdkdiag.AppendErrorf(diags, "reading API Gateway Stage (%s): %s", stageName, err)
		}
	}

	if shouldDeleteStage {
		_, err := conn.DeleteStage(ctx, &apigateway.DeleteStageInput{
			StageName: aws.String(stageName),
			RestApiId: aws.String(restAPIID),
		})

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "deleting API Gateway Stage (%s): %s", stageName, err)
		}
	}

	log.Printf("[DEBUG] Deleting API Gateway Deployment: %s", d.Id())
	_, err := conn.DeleteDeployment(ctx, &apigateway.DeleteDeploymentInput{
		DeploymentId: aws.String(d.Id()),
		RestApiId:    aws.String(restAPIID),
	})

	if errs.IsA[*types.NotFoundException](err) {
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

func findDeploymentByTwoPartKey(ctx context.Context, conn *apigateway.Client, restAPIID, deploymentID string) (*apigateway.GetDeploymentOutput, error) {
	input := &apigateway.GetDeploymentInput{
		DeploymentId: aws.String(deploymentID),
		RestApiId:    aws.String(restAPIID),
	}

	output, err := conn.GetDeployment(ctx, input)

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

func expandDeploymentCanarySettings(tfMap map[string]interface{}) *types.DeploymentCanarySettings {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.DeploymentCanarySettings{}

	if v, ok := tfMap["percent_traffic"].(float64); ok {
		apiObject.PercentTraffic = v
	}

	if v, ok := tfMap["stage_variable_overrides"].(map[string]interface{}); ok && len(v) > 0 {
		apiObject.StageVariableOverrides = flex.ExpandStringValueMap(v)
	}

	if v, ok := tfMap["use_stage_cache"].(bool); ok {
		apiObject.UseStageCache = v
	}

	return apiObject
}
