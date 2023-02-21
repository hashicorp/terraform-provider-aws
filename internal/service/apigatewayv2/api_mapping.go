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
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

func ResourceAPIMapping() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAPIMappingCreate,
		ReadWithoutTimeout:   resourceAPIMappingRead,
		UpdateWithoutTimeout: resourceAPIMappingUpdate,
		DeleteWithoutTimeout: resourceAPIMappingDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceAPIMappingImport,
		},

		Schema: map[string]*schema.Schema{
			"api_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"api_mapping_key": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"domain_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"stage": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceAPIMappingCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayV2Conn()

	req := &apigatewayv2.CreateApiMappingInput{
		ApiId:      aws.String(d.Get("api_id").(string)),
		DomainName: aws.String(d.Get("domain_name").(string)),
		Stage:      aws.String(d.Get("stage").(string)),
	}
	if v, ok := d.GetOk("api_mapping_key"); ok {
		req.ApiMappingKey = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating API Gateway v2 API mapping: %s", req)
	resp, err := conn.CreateApiMappingWithContext(ctx, req)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating API Gateway v2 API mapping: %s", err)
	}

	d.SetId(aws.StringValue(resp.ApiMappingId))

	return append(diags, resourceAPIMappingRead(ctx, d, meta)...)
}

func resourceAPIMappingRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayV2Conn()

	resp, err := conn.GetApiMappingWithContext(ctx, &apigatewayv2.GetApiMappingInput{
		ApiMappingId: aws.String(d.Id()),
		DomainName:   aws.String(d.Get("domain_name").(string)),
	})
	if tfawserr.ErrCodeEquals(err, apigatewayv2.ErrCodeNotFoundException) && !d.IsNewResource() {
		log.Printf("[WARN] API Gateway v2 API mapping (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading API Gateway v2 API mapping: %s", err)
	}

	d.Set("api_id", resp.ApiId)
	d.Set("api_mapping_key", resp.ApiMappingKey)
	d.Set("stage", resp.Stage)

	return diags
}

func resourceAPIMappingUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayV2Conn()

	req := &apigatewayv2.UpdateApiMappingInput{
		ApiId:        aws.String(d.Get("api_id").(string)),
		ApiMappingId: aws.String(d.Id()),
		DomainName:   aws.String(d.Get("domain_name").(string)),
	}
	if d.HasChange("api_mapping_key") {
		req.ApiMappingKey = aws.String(d.Get("api_mapping_key").(string))
	}
	if d.HasChange("stage") {
		req.Stage = aws.String(d.Get("stage").(string))
	}

	log.Printf("[DEBUG] Updating API Gateway v2 API mapping: %s", req)
	_, err := conn.UpdateApiMappingWithContext(ctx, req)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating API Gateway v2 API mapping: %s", err)
	}

	return append(diags, resourceAPIMappingRead(ctx, d, meta)...)
}

func resourceAPIMappingDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayV2Conn()

	log.Printf("[DEBUG] Deleting API Gateway v2 API mapping (%s)", d.Id())
	_, err := conn.DeleteApiMappingWithContext(ctx, &apigatewayv2.DeleteApiMappingInput{
		ApiMappingId: aws.String(d.Id()),
		DomainName:   aws.String(d.Get("domain_name").(string)),
	})
	if tfawserr.ErrCodeEquals(err, apigatewayv2.ErrCodeNotFoundException) {
		return diags
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting API Gateway v2 API mapping: %s", err)
	}

	return diags
}

func resourceAPIMappingImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	parts := strings.Split(d.Id(), "/")
	if len(parts) != 2 {
		return []*schema.ResourceData{}, fmt.Errorf("wrong format of import ID (%s), use: 'api-mapping-id/domain-name'", d.Id())
	}

	d.SetId(parts[0])
	d.Set("domain_name", parts[1])

	return []*schema.ResourceData{d}, nil
}
