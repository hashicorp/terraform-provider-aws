package apigatewayv2

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigatewayv2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

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
	conn := meta.(*conns.AWSClient).APIGatewayV2Conn()

	req := &apigatewayv2.CreateDeploymentInput{
		ApiId: aws.String(d.Get("api_id").(string)),
	}
	if v, ok := d.GetOk("description"); ok {
		req.Description = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating API Gateway v2 deployment: %s", req)
	resp, err := conn.CreateDeploymentWithContext(ctx, req)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating API Gateway v2 deployment: %s", err)
	}

	d.SetId(aws.StringValue(resp.DeploymentId))

	if _, err := WaitDeploymentDeployed(ctx, conn, d.Get("api_id").(string), d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for API Gateway v2 deployment (%s) creation: %s", d.Id(), err)
	}

	return append(diags, resourceDeploymentRead(ctx, d, meta)...)
}

func resourceDeploymentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayV2Conn()

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
	conn := meta.(*conns.AWSClient).APIGatewayV2Conn()

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
	conn := meta.(*conns.AWSClient).APIGatewayV2Conn()

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
