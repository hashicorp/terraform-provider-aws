// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apigatewayv2

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/apigatewayv2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/apigatewayv2/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_apigatewayv2_deployment", name="Deployment")
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
			"api_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"auto_deployed": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			names.AttrDescription: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 1024),
			},
			names.AttrTriggers: {
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
	conn := meta.(*conns.AWSClient).APIGatewayV2Client(ctx)

	apiID := d.Get("api_id").(string)
	input := &apigatewayv2.CreateDeploymentInput{
		ApiId: aws.String(apiID),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	output, err := conn.CreateDeployment(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating API Gateway v2 Deployment: %s", err)
	}

	d.SetId(aws.ToString(output.DeploymentId))

	if _, err := waitDeploymentDeployed(ctx, conn, apiID, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for API Gateway v2 Deployment (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceDeploymentRead(ctx, d, meta)...)
}

func resourceDeploymentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayV2Client(ctx)

	output, err := findDeploymentByTwoPartKey(ctx, conn, d.Get("api_id").(string), d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] API Gateway v2 Deployment (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading API Gateway v2 Deployment (%s): %s", d.Id(), err)
	}

	d.Set("auto_deployed", output.AutoDeployed)
	d.Set(names.AttrDescription, output.Description)

	return diags
}

func resourceDeploymentUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayV2Client(ctx)

	apiID := d.Get("api_id").(string)
	input := &apigatewayv2.UpdateDeploymentInput{
		ApiId:        aws.String(apiID),
		DeploymentId: aws.String(d.Id()),
	}

	if d.HasChange(names.AttrDescription) {
		input.Description = aws.String(d.Get(names.AttrDescription).(string))
	}

	_, err := conn.UpdateDeployment(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating API Gateway v2 Deployment (%s): %s", d.Id(), err)
	}

	if _, err := waitDeploymentDeployed(ctx, conn, apiID, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for API Gateway v2 Deployment (%s) update: %s", d.Id(), err)
	}

	return append(diags, resourceDeploymentRead(ctx, d, meta)...)
}

func resourceDeploymentDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayV2Client(ctx)

	log.Printf("[DEBUG] Deleting API Gateway v2 Deployment (%s)", d.Id())
	_, err := conn.DeleteDeployment(ctx, &apigatewayv2.DeleteDeploymentInput{
		ApiId:        aws.String(d.Get("api_id").(string)),
		DeploymentId: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.NotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting API Gateway v2 Deployment (%s): %s", d.Id(), err)
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

func findDeploymentByTwoPartKey(ctx context.Context, conn *apigatewayv2.Client, apiID, deploymentID string) (*apigatewayv2.GetDeploymentOutput, error) {
	input := &apigatewayv2.GetDeploymentInput{
		ApiId:        aws.String(apiID),
		DeploymentId: aws.String(deploymentID),
	}

	return findDeployment(ctx, conn, input)
}

func findDeployment(ctx context.Context, conn *apigatewayv2.Client, input *apigatewayv2.GetDeploymentInput) (*apigatewayv2.GetDeploymentOutput, error) {
	output, err := conn.GetDeployment(ctx, input)

	if errs.IsA[*awstypes.NotFoundException](err) {
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

func statusDeployment(ctx context.Context, conn *apigatewayv2.Client, apiID, deploymentID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findDeploymentByTwoPartKey(ctx, conn, apiID, deploymentID)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.DeploymentStatus), nil
	}
}

func waitDeploymentDeployed(ctx context.Context, conn *apigatewayv2.Client, apiID, deploymentID string) (*apigatewayv2.GetDeploymentOutput, error) { //nolint:unparam
	const (
		timeout = 5 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.DeploymentStatusPending),
		Target:  enum.Slice(awstypes.DeploymentStatusDeployed),
		Refresh: statusDeployment(ctx, conn, apiID, deploymentID),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*apigatewayv2.GetDeploymentOutput); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.DeploymentStatusMessage)))

		return output, err
	}

	return nil, err
}
