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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourceUsagePlanKey() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceUsagePlanKeyCreate,
		ReadWithoutTimeout:   resourceUsagePlanKeyRead,
		DeleteWithoutTimeout: resourceUsagePlanKeyDelete,

		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				idParts := strings.Split(d.Id(), "/")
				if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
					return nil, fmt.Errorf("Unexpected format of ID (%q), expected USAGE-PLAN-ID/USAGE-PLAN-KEY-ID", d.Id())
				}
				usagePlanId := idParts[0]
				usagePlanKeyId := idParts[1]
				d.Set("usage_plan_id", usagePlanId)
				d.Set("key_id", usagePlanKeyId)
				d.SetId(usagePlanKeyId)
				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"key_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"key_type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"usage_plan_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"value": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceUsagePlanKeyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayConn()

	input := &apigateway.CreateUsagePlanKeyInput{
		KeyId:       aws.String(d.Get("key_id").(string)),
		KeyType:     aws.String(d.Get("key_type").(string)),
		UsagePlanId: aws.String(d.Get("usage_plan_id").(string)),
	}

	output, err := conn.CreateUsagePlanKeyWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating API Gateway Usage Plan Key: %s", err)
	}

	d.SetId(aws.StringValue(output.Id))

	return append(diags, resourceUsagePlanKeyRead(ctx, d, meta)...)
}

func resourceUsagePlanKeyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayConn()

	upk, err := FindUsagePlanKeyByTwoPartKey(ctx, conn, d.Get("usage_plan_id").(string), d.Get("key_id").(string))

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] API Gateway Usage Plan Key (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading API Gateway Usage Plan Key (%s): %s", d.Id(), err)
	}

	d.Set("key_type", upk.Type)
	d.Set("name", upk.Name)
	d.Set("value", upk.Value)

	return diags
}

func resourceUsagePlanKeyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayConn()

	log.Printf("[DEBUG] Deleting API Gateway Usage Plan Key: %s", d.Id())
	_, err := conn.DeleteUsagePlanKeyWithContext(ctx, &apigateway.DeleteUsagePlanKeyInput{
		KeyId:       aws.String(d.Get("key_id").(string)),
		UsagePlanId: aws.String(d.Get("usage_plan_id").(string)),
	})

	if tfawserr.ErrCodeEquals(err, apigateway.ErrCodeNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting API Gateway Usage Plan Key (%s): %s", d.Id(), err)
	}

	return diags
}

func FindUsagePlanKeyByTwoPartKey(ctx context.Context, conn *apigateway.APIGateway, usagePlanID, keyID string) (*apigateway.UsagePlanKey, error) {
	input := &apigateway.GetUsagePlanKeyInput{
		KeyId:       aws.String(keyID),
		UsagePlanId: aws.String(usagePlanID),
	}

	output, err := conn.GetUsagePlanKeyWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, apigateway.ErrCodeNotFoundException) {
		return nil, &resource.NotFoundError{
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
