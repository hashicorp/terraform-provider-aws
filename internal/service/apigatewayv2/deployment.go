// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apigatewayv2

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigatewayv2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

// @SDKResource("aws_apigatewayv2_deployment")
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
			"api_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"auto_deployed": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 1024),
			},
			"triggers": {
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
	conn := meta.(*conns.AWSClient).APIGatewayV2Conn(ctx)

	input := &apigatewayv2.CreateDeploymentInput{
		ApiId: aws.String(d.Get("api_id").(string)),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	output, err := conn.CreateDeploymentWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating API Gateway v2 Deployment: %s", err)
	}

	d.SetId(aws.StringValue(output.DeploymentId))

	if _, err := WaitDeploymentDeployed(ctx, conn, d.Get("api_id").(string), d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for API Gateway v2 Deployment (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceDeploymentRead(ctx, d, meta)...)
}

func resourceDeploymentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayV2Conn(ctx)

	outputRaw, _, err := StatusDeployment(ctx, conn, d.Get("api_id").(string), d.Id())()
	if tfawserr.ErrCodeEquals(err, apigatewayv2.ErrCodeNotFoundException) && !d.IsNewResource() {
		log.Printf("[WARN] API Gateway v2 deployment (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading API Gateway v2 deployment: %s", err)
	}

	output := outputRaw.(*apigatewayv2.GetDeploymentOutput)
	d.Set("auto_deployed", output.AutoDeployed)
	d.Set("description", output.Description)

	return diags
}

func resourceDeploymentUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayV2Conn(ctx)

	req := &apigatewayv2.UpdateDeploymentInput{
		ApiId:        aws.String(d.Get("api_id").(string)),
		DeploymentId: aws.String(d.Id()),
	}
	if d.HasChange("description") {
		req.Description = aws.String(d.Get("description").(string))
	}

	log.Printf("[DEBUG] Updating API Gateway v2 deployment: %s", req)
	_, err := conn.UpdateDeploymentWithContext(ctx, req)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating API Gateway v2 deployment: %s", err)
	}

	if _, err := WaitDeploymentDeployed(ctx, conn, d.Get("api_id").(string), d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for API Gateway v2 deployment (%s) update: %s", d.Id(), err)
	}

	return append(diags, resourceDeploymentRead(ctx, d, meta)...)
}

func resourceDeploymentDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayV2Conn(ctx)

	log.Printf("[DEBUG] Deleting API Gateway v2 deployment (%s)", d.Id())
	_, err := conn.DeleteDeploymentWithContext(ctx, &apigatewayv2.DeleteDeploymentInput{
		ApiId:        aws.String(d.Get("api_id").(string)),
		DeploymentId: aws.String(d.Id()),
	})
	if tfawserr.ErrCodeEquals(err, apigatewayv2.ErrCodeNotFoundException) {
		return diags
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting API Gateway v2 deployment: %s", err)
	}

	return diags
}

func resourceDeploymentImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	parts := strings.Split(d.Id(), "/")
	if len(parts) != 2 {
		return []*schema.ResourceData{}, fmt.Errorf("wrong format of import ID (%s), use: 'api-id/deployment-id'", d.Id())
	}

	d.SetId(parts[1])
	d.Set("api_id", parts[0])

	return []*schema.ResourceData{d}, nil
}

// StatusDeployment fetches the Deployment and its Status
func StatusDeployment(ctx context.Context, conn *apigatewayv2.ApiGatewayV2, apiId, deploymentId string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &apigatewayv2.GetDeploymentInput{
			ApiId:        aws.String(apiId),
			DeploymentId: aws.String(deploymentId),
		}

		output, err := conn.GetDeploymentWithContext(ctx, input)

		if err != nil {
			return nil, apigatewayv2.DeploymentStatusFailed, err
		}

		// Error messages can also be contained in the response with FAILED status

		if aws.StringValue(output.DeploymentStatus) == apigatewayv2.DeploymentStatusFailed {
			return output, apigatewayv2.DeploymentStatusFailed, fmt.Errorf("%s", aws.StringValue(output.DeploymentStatusMessage))
		}

		return output, aws.StringValue(output.DeploymentStatus), nil
	}
}

const (
	// Maximum amount of time to wait for a Deployment to return Deployed
	DeploymentDeployedTimeout = 5 * time.Minute
)

// WaitDeploymentDeployed waits for a Deployment to return Deployed
func WaitDeploymentDeployed(ctx context.Context, conn *apigatewayv2.ApiGatewayV2, apiId, deploymentId string) (*apigatewayv2.GetDeploymentOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{apigatewayv2.DeploymentStatusPending},
		Target:  []string{apigatewayv2.DeploymentStatusDeployed},
		Refresh: StatusDeployment(ctx, conn, apiId, deploymentId),
		Timeout: DeploymentDeployedTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if v, ok := outputRaw.(*apigatewayv2.GetDeploymentOutput); ok {
		return v, err
	}

	return nil, err
}
